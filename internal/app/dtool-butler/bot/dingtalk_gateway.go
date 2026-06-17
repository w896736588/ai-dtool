package bot

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"dev_tool/internal/app/dtool-butler/define"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// IncomingMessage 接收到的机器人消息，统一格式后投递给管家核心处理。
type IncomingMessage struct {
	ConversationId   string
	ConversationType string // 1 单聊 2 群聊
	SenderNick       string
	SenderStaffId    string
	Text             string
	SessionWebhook   string // 用于回复该消息的临时 webhook
}

// DingTalkGateway 钉钉 Stream 模式网关实现。
type DingTalkGateway struct {
	botConfig *define.BotConfigItem
	cli       *client.StreamClient
	msgChan   chan<- IncomingMessage
}

// NewDingTalkGateway 创建钉钉网关，msgChan 为消息投递通道。
func NewDingTalkGateway(botConfig *define.BotConfigItem, msgChan chan<- IncomingMessage) *DingTalkGateway {
	return &DingTalkGateway{
		botConfig: botConfig,
		msgChan:   msgChan,
	}
}

// Start 建立 Stream 长连接并注册机器人消息回调，非阻塞。
func (g *DingTalkGateway) Start() error {
	if g.botConfig == nil || g.botConfig.AppKey == `` || g.botConfig.AppSecret == `` {
		return fmt.Errorf(`钉钉机器人配置缺失 app_key/app_secret`)
	}
	logger.SetLogger(logger.NewStdTestLoggerWithDebug())
	g.cli = client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(g.botConfig.AppKey, g.botConfig.AppSecret)),
	)
	g.cli.RegisterChatBotCallbackRouter(g.onChatBotMessage)
	if err := g.cli.Start(context.Background()); err != nil {
		return fmt.Errorf(`钉钉 Stream 启动失败 %w`, err)
	}
	gstool.FmtPrintlnLogTime(`[butler-bot] 钉钉 Stream 连接成功`)
	return nil
}

// Close 关闭 Stream 连接。
func (g *DingTalkGateway) Close() {
	if g.cli != nil {
		g.cli.Close()
	}
}

// SendText 通过群机器人 webhook 主动发送文本（用于打招呼/休眠通知等无 incoming 的场景）。
// webhook_url 未配置时跳过。
func (g *DingTalkGateway) SendText(conversationId, text string) error {
	if g.botConfig == nil || g.botConfig.WebhookUrl == `` {
		gstool.FmtPrintlnLogTime(`[butler-bot] webhook_url 未配置，跳过主动发送`)
		return nil
	}
	return sendDingtalkWebhookText(g.botConfig.WebhookUrl, g.botConfig.Secret, text)
}

// onChatBotMessage 钉钉机器人消息回调，解析后投递到消息通道。
func (g *DingTalkGateway) onChatBotMessage(ctx context.Context, data *chatbot.BotCallbackDataModel) ([]byte, error) {
	if data == nil {
		return []byte(``), nil
	}
	msg := IncomingMessage{
		ConversationId:   data.ConversationId,
		ConversationType: data.ConversationType,
		SenderNick:       data.SenderNick,
		SenderStaffId:    data.SenderStaffId,
		Text:             cast.ToString(data.Text.Content),
		SessionWebhook:   data.SessionWebhook,
	}
	gstool.FmtPrintlnLogTime(`[butler-bot] 收到消息 会话=%s 发送者=%s 内容=%s`,
		msg.ConversationId, msg.SenderNick, msg.Text)
	go func() {
		g.msgChan <- msg
	}()
	return []byte(``), nil
}

// sendDingtalkWebhookText 通过钉钉群机器人 webhook 发送文本消息，支持加签。
func sendDingtalkWebhookText(webhookUrl, secret, text string) error {
	url := strings.TrimSpace(webhookUrl)
	if secret != `` {
		timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
		signStr := timestamp + "\n" + secret
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(signStr))
		sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		url = fmt.Sprintf("%s&timestamp=%s&sign=%s", url, timestamp, sign)
	}
	body := map[string]any{
		`msgtype`: `text`,
		`text`:    map[string]string{`content`: text},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf(`json marshal: %w`, err)
	}
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Post(url, "application/json; charset=utf-8", strings.NewReader(string(bodyBytes)))
	if err != nil {
		return fmt.Errorf(`http post: %w`, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf(`response parse: %w`, err)
	}
	if cast.ToInt(result["errcode"]) != 0 {
		return fmt.Errorf(`dingtalk error: errcode=%v errmsg=%v`, result["errcode"], result["errmsg"])
	}
	return nil
}
