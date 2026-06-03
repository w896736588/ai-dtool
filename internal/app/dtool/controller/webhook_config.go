package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// WebhookConfigList 返回所有 Webhook 配置列表
func WebhookConfigList(c *gin.Context) {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_webhook_config ORDER BY id`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	items := make([]define.WebhookConfigItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, define.WebhookConfigItem{
			Id:         cast.ToInt(row["id"]),
			Name:       cast.ToString(row["name"]),
			Type:       cast.ToString(row["type"]),
			WebhookUrl: cast.ToString(row["webhook_url"]),
			Secret:     cast.ToString(row["secret"]),
			CreatedAt:  cast.ToInt64(row["created_at"]),
			UpdatedAt:  cast.ToInt64(row["updated_at"]),
		})
	}

	gsgin.GinResponseSuccess(c, "", gin.H{"list": items})
}

// WebhookConfigSave 新增/编辑 Webhook 配置
func WebhookConfigSave(c *gin.Context) {
	var req define.WebhookConfigSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if req.Name == "" {
		gsgin.GinResponseError(c, "名称不能为空", nil)
		return
	}
	if req.WebhookUrl == "" {
		gsgin.GinResponseError(c, "Webhook 地址不能为空", nil)
		return
	}
	if req.Type == "" {
		req.Type = define.WebhookTypeDingtalk
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.TrimSpace(req.Type)
	req.WebhookUrl = strings.TrimSpace(req.WebhookUrl)
	req.Secret = strings.TrimSpace(req.Secret)
	if !isSupportedWebhookType(req.Type) {
		gsgin.GinResponseError(c, "不支持的 Webhook 类型", nil)
		return
	}

	now := time.Now().Unix()

	if req.Id > 0 {
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_webhook_config SET name = ?, type = ?, webhook_url = ?, secret = ?, updated_at = ? WHERE id = ?`,
			req.Name, req.Type, req.WebhookUrl, req.Secret, now, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", define.WebhookConfigItem{
			Id:         req.Id,
			Name:       req.Name,
			Type:       req.Type,
			WebhookUrl: req.WebhookUrl,
			Secret:     req.Secret,
			UpdatedAt:  now,
		})
		return
	}

	lastId, err := common.DbMain.Client.InsertBySql(
		`INSERT INTO tbl_webhook_config (name, type, webhook_url, secret, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		req.Name, req.Type, req.WebhookUrl, req.Secret, now, now,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", define.WebhookConfigItem{
		Id:         int(lastId),
		Name:       req.Name,
		Type:       req.Type,
		WebhookUrl: req.WebhookUrl,
		Secret:     req.Secret,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
}

// WebhookConfigDelete 删除 Webhook 配置
func WebhookConfigDelete(c *gin.Context) {
	var req define.WebhookConfigDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if req.Id <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}

	// 检查是否有 agent_cli 正在使用该配置
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT COUNT(*) as cnt FROM tbl_agent_cli WHERE webhook_config_id = ?`, req.Id,
	).One()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if cast.ToInt(row["cnt"]) > 0 {
		gsgin.GinResponseError(c, "该配置正在被 Agent CLI 使用，无法删除", nil)
		return
	}

	_, err = common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_webhook_config WHERE id = ?`, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", nil)
}

// WebhookConfigTest 测试发送 Webhook 通知
func WebhookConfigTest(c *gin.Context) {
	var req define.WebhookConfigTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.TrimSpace(req.Type)
	req.WebhookUrl = strings.TrimSpace(req.WebhookUrl)
	req.Secret = strings.TrimSpace(req.Secret)

	if req.Type == "" {
		req.Type = define.WebhookTypeDingtalk
	}
	if req.Name == "" {
		req.Name = "Webhook 测试"
	}
	if req.WebhookUrl == "" {
		gsgin.GinResponseError(c, "Webhook 地址不能为空", nil)
		return
	}
	if !isSupportedWebhookType(req.Type) {
		gsgin.GinResponseError(c, "不支持的 Webhook 类型", nil)
		return
	}

	config := &define.WebhookConfigItem{
		Name:       req.Name,
		Type:       req.Type,
		WebhookUrl: req.WebhookUrl,
		Secret:     req.Secret,
	}
	title := fmt.Sprintf("Webhook 测试通知 - %s", req.Name)
	text := "这是一条测试消息，用于验证当前 Webhook 配置是否可用。"
	if req.SingleURL != "" {
		text = text + "\n\n已附带详情链接按钮，请点击验证卡片跳转。"
	}
	if err := business.SendWebhookNotify(config, title, text, req.SingleURL); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "测试发送成功", nil)
}

func isSupportedWebhookType(webhookType string) bool {
	switch webhookType {
	case define.WebhookTypeDingtalk, define.WebhookTypeFeishu, define.WebhookTypeWecom:
		return true
	default:
		return false
	}
}
