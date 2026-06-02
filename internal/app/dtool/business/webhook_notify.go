package business

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"

	"github.com/spf13/cast"
)

const webhookRequestTimeout = 10 * time.Second

// GetWebhookConfigByAgentCliId 通过 agent_cli_id 获取关联的 webhook 配置，未配置返回 nil。
func GetWebhookConfigByAgentCliId(agentCliId int) *define.WebhookConfigItem {
	if agentCliId <= 0 {
		return nil
	}
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT webhook_config_id FROM tbl_agent_cli WHERE id = ?`, agentCliId,
	).One()
	if err != nil || len(row) == 0 {
		return nil
	}
	webhookConfigId := cast.ToInt(row["webhook_config_id"])
	if webhookConfigId <= 0 {
		return nil
	}
	wRow, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_webhook_config WHERE id = ?`, webhookConfigId,
	).One()
	if err != nil || len(wRow) == 0 {
		return nil
	}
	return &define.WebhookConfigItem{
		Id:         cast.ToInt(wRow["id"]),
		Name:       cast.ToString(wRow["name"]),
		Type:       cast.ToString(wRow["type"]),
		WebhookUrl: cast.ToString(wRow["webhook_url"]),
		Secret:     cast.ToString(wRow["secret"]),
	}
}

// webhook 消息类型与按钮文案常量。
const (
	feishuMsgTypeText          = "text"
	feishuMsgTypeInteractive   = "interactive"
	feishuActionCardBtnTitle   = "查看详情"
	dingtalkMsgTypeText        = "text"
	dingtalkMsgTypeActionCard  = "actionCard"
	dingtalkActionCardBtnTitle = "查看详情"
)

// SendWebhookNotify 根据 webhook 配置发送通知。
// title 为消息标题；text 为消息正文（支持 Markdown）；singleURL 为查看详情按钮跳转地址，
// 不为空时优先发送带跳转能力的消息，为空时回退为普通 text 消息。
func SendWebhookNotify(config *define.WebhookConfigItem, title, text, singleURL string) error {
	if config == nil || config.WebhookUrl == "" {
		return nil
	}
	switch config.Type {
	case define.WebhookTypeDingtalk:
		return sendDingtalkNotify(config.WebhookUrl, config.Secret, title, text, singleURL)
	case define.WebhookTypeFeishu:
		return sendFeishuNotify(config.WebhookUrl, config.Secret, title, text, singleURL)
	default:
		log.Printf("[webhook-notify] 暂不支持的类型: %s", config.Type)
		return nil
	}
}

// sendDingtalkNotify 发送钉钉消息，支持 HMAC-SHA256 加签。
// 当 singleURL 不为空时使用 actionCard（带"查看详情"按钮），否则回退到 text 类型。
func sendDingtalkNotify(webhookUrl, secret, title, text, singleURL string) error {
	url := strings.TrimSpace(webhookUrl)

	if secret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
		signStr := timestamp + "\n" + secret
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(signStr))
		sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		url = fmt.Sprintf("%s&timestamp=%s&sign=%s", url, timestamp, sign)
	}

	var body map[string]any
	if strings.TrimSpace(singleURL) != "" {
		// 有跳转链接：使用 actionCard，钉钉官方字段名为驼峰（singleTitle/singleURL），不能用 snake_case。
		body = map[string]any{
			"msgtype": dingtalkMsgTypeActionCard,
			"actionCard": map[string]any{
				"title":       title,
				"text":        text,
				"singleTitle": dingtalkActionCardBtnTitle,
				"singleURL":   singleURL,
			},
		}
		log.Printf("[webhook-notify][dingtalk] 使用 actionCard, title_len=%d text_len=%d single_url=%s", len(title), len(text), singleURL)
	} else {
		// 无跳转链接：回退到普通文本，避免 actionCard 缺少必填 single_url 导致发送失败。
		fallbackContent := text
		if strings.TrimSpace(title) != "" {
			fallbackContent = title + "\n" + text
		}
		body = map[string]any{
			"msgtype": dingtalkMsgTypeText,
			"text": map[string]string{
				"content": fallbackContent,
			},
		}
		log.Printf("[webhook-notify][dingtalk] singleURL 为空,回退到 text 类型 content_len=%d", len(fallbackContent))
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	resp, err := postWebhookJSON(url, bodyBytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("response parse: %w", err)
	}
	if cast.ToInt(result["errcode"]) != 0 {
		return fmt.Errorf("dingtalk error: errcode=%v errmsg=%v", result["errcode"], result["errmsg"])
	}
	return nil
}

// sendFeishuNotify 发送飞书群自定义机器人消息，支持签名校验。
// 有跳转链接时使用 interactive 卡片，否则发送 text。
func sendFeishuNotify(webhookUrl, secret, title, text, singleURL string) error {
	url := strings.TrimSpace(webhookUrl)
	body := map[string]any{}
	if secret != "" {
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		sign, err := buildFeishuWebhookSign(secret, timestamp)
		if err != nil {
			return err
		}
		body["timestamp"] = timestamp
		body["sign"] = sign
	}

	if strings.TrimSpace(singleURL) != "" {
		body["msg_type"] = feishuMsgTypeInteractive
		body["card"] = map[string]any{
			"config": map[string]any{
				"wide_screen_mode": true,
			},
			"header": map[string]any{
				"template": "blue",
				"title": map[string]any{
					"tag":     "plain_text",
					"content": strings.TrimSpace(title),
				},
			},
			"elements": []map[string]any{
				{
					"tag":     "markdown",
					"content": strings.TrimSpace(text),
				},
				{
					"tag": "action",
					"actions": []map[string]any{
						{
							"tag": "button",
							"text": map[string]any{
								"tag":     "plain_text",
								"content": feishuActionCardBtnTitle,
							},
							"type": "primary",
							"url":  strings.TrimSpace(singleURL),
						},
					},
				},
			},
		}
		log.Printf("[webhook-notify][feishu] 使用 interactive card, title_len=%d text_len=%d single_url=%s", len(title), len(text), singleURL)
	} else {
		body["msg_type"] = feishuMsgTypeText
		content := strings.TrimSpace(text)
		if strings.TrimSpace(title) != "" {
			content = strings.TrimSpace(title) + "\n" + content
		}
		body["content"] = map[string]string{
			"text": content,
		}
		log.Printf("[webhook-notify][feishu] singleURL 为空,回退到 text 类型 content_len=%d", len(content))
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	resp, err := postWebhookJSON(url, bodyBytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("response parse: %w", err)
	}
	if cast.ToInt(result["code"]) != 0 {
		return fmt.Errorf("feishu error: code=%v msg=%v", result["code"], result["msg"])
	}
	return nil
}

func buildFeishuWebhookSign(secret, timestamp string) (string, error) {
	signKey := timestamp + "\n" + strings.TrimSpace(secret)
	mac := hmac.New(sha256.New, []byte(signKey))
	if _, err := mac.Write([]byte{}); err != nil {
		return "", fmt.Errorf("generate feishu sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

func postWebhookJSON(url string, body []byte) (*http.Response, error) {
	client := &http.Client{Timeout: webhookRequestTimeout}
	resp, err := client.Post(url, "application/json; charset=utf-8", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	return resp, nil
}
