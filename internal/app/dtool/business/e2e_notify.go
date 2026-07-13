package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// E2ENotifyRunCompleted 异步执行完成通知。
// 为避免在 webhook 不可用时阻塞主流程，发送失败仅记录日志，不影响用例结果。
func E2ENotifyRunCompletedInner(runID int64) {
	time.Sleep(1 * time.Second) // 等待 run 状态稳定
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT r.*, g.webhook_config_id, c.name AS case_name, g.name AS group_name
		 FROM tbl_e2e_run r
		 LEFT JOIN tbl_e2e_case c ON c.id = r.case_id
		 LEFT JOIN tbl_e2e_group g ON g.id = r.group_id
		 WHERE r.id = ?`, runID,
	).One()
	if err != nil || len(row) == 0 {
		return
	}
	webhookID := cast.ToInt(row["webhook_config_id"])
	if webhookID <= 0 {
		return
	}
	cfgRow, _ := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_webhook_config WHERE id = ?`, webhookID,
	).One()
	if cfgRow == nil {
		return
	}
	cfg := &define.WebhookConfigItem{
		Id:         webhookID,
		Name:       cast.ToString(cfgRow["name"]),
		Type:       cast.ToString(cfgRow["type"]),
		WebhookUrl: cast.ToString(cfgRow["webhook_url"]),
		Secret:     cast.ToString(cfgRow["secret"]),
	}

	status := cast.ToString(row["status"])
	title := buildE2ENotifyTitle(status)
	caseName := cast.ToString(row["case_name"])
	groupName := cast.ToString(row["group_name"])
	passedSteps := cast.ToInt(row["passed_steps"])
	failedSteps := cast.ToInt(row["failed_steps"])

	text := fmt.Sprintf(
		"**用例**: %s\n**分组**: %s\n**状态**: %s\n**步骤通过**: %d | 失败 %d\n**耗时**: %d ms\n**Run ID**: %d",
		caseName, groupName, strings.ToUpper(status),
		passedSteps, failedSteps,
		cast.ToInt(row["duration_ms"]), runID,
	)
	// 调用同包内 webhook_notify.go 中的 SendWebhookNotify
	_ = SendWebhookNotify(cfg, title, text, "")
}

func buildE2ENotifyTitle(status string) string {
	switch status {
	case define.E2ERunStatusPassed:
		return "✅ E2E 用例执行通过"
	case define.E2ERunStatusFailed:
		return "❌ E2E 用例执行失败"
	case define.E2ERunStatusStopped:
		return "⏹️ E2E 用例已停止"
	case define.E2ERunStatusError:
		return "⚠️ E2E 用例执行异常"
	default:
		return "ℹ️ E2E 用例状态更新"
	}
}

// E2ENotifyRunCompleted 是异步入口。
func E2ENotifyRunCompleted(runID int64) {
	go E2ENotifyRunCompletedInner(runID)
}
