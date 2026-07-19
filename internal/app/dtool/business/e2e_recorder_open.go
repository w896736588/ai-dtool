// Package business 提供基于 smart_link + ws_token 的录制入口。
package business

import (
	"crypto/rand"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component/e2e/store"
	"dev_tool/internal/app/dtool/define"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
)

// recorderProxyPath recorder iframe proxy 同源路径。
// 本任务阶段 controller/router 还未挂载；任务 10 会注册。
const recorderProxyPath = "/api/e2e/recorder/proxy.html"

// recorderInitScriptFmt AddInitScript 注入到被录 page 的 JS 模板。
// %q 会自动做 JSON 字符串转义，避免 ws_token 中的特殊字符破坏脚本。
const recorderInitScriptFmt = `
(function(){
  window.__dtoolRecorder = {wsToken:%q, recorderUrl:%q, sessionUUID:%q};
  document.addEventListener('DOMContentLoaded', function(){
    var iframe = document.createElement('iframe');
    iframe.src = window.__dtoolRecorder.recorderUrl;
    iframe.style.cssText = 'position:fixed;width:1px;height:1px;opacity:0;pointer-events:none;border:0;right:0;bottom:0;';
    document.body.appendChild(iframe);
  });
})();
`

// E2ERecordOpen 开启一次 smart_link 录制会话：开浏览器 → 写 DB → 注入 init script → 跳到 env_url。
// 返回 *E2ERecordOpenResponse，错误通过 Response.Error 透出给 controller（避免吞掉用户可见提示）。
func E2ERecordOpen(req *define.E2ERecordOpenRequest) (*define.E2ERecordOpenResponse, error) {
	if req == nil || req.SmartLinkID <= 0 {
		return nil, errors.New("smart_link_id 必须为正数")
	}

	engine := GetE2EEngine()
	browserID, page, err := engine.OpenRecorder(req.SmartLinkID, req.UserName)
	if err != nil {
		return &define.E2ERecordOpenResponse{OK: false, Error: err.Error()}, nil
	}

	envURL, _ := fetchSmartLinkEnvURL(req.SmartLinkID)
	if envURL == "" {
		_ = page.Close()
		return &define.E2ERecordOpenResponse{OK: false, Error: "未找到 smart_link 对应 link"}, nil
	}

	sessionID, sessionUUID, err := newRecordSessionForRecorder(req, browserID, envURL)
	if err != nil {
		_ = page.Close()
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	wsToken, err := generateWSToken()
	if err != nil {
		_ = page.Close()
		return nil, err
	}
	recorderURL := recorderProxyPath
	if err := store.NewRecordSessionStore().UpdateSmartLink(sessionID, req.SmartLinkID, req.UserName, wsToken, recorderURL, req.LinkID); err != nil {
		_ = page.Close()
		return nil, err
	}

	initBody := fmt.Sprintf(recorderInitScriptFmt, wsToken, recorderURL, sessionUUID)
	if err := page.AddInitScript(playwright.Script{Content: &initBody}); err != nil {
		// init script 失败：仍返回 session，但前端会提示
		return &define.E2ERecordOpenResponse{
			OK:          false,
			Error:       err.Error(),
			SessionID:   sessionID,
			SessionUUID: sessionUUID,
			WSToken:     wsToken,
			RecorderURL: recorderURL,
			EnvURL:      envURL,
		}, nil
	}

	if _, err := page.Goto(envURL); err != nil {
		return &define.E2ERecordOpenResponse{OK: false, Error: err.Error()}, nil
	}

	return &define.E2ERecordOpenResponse{
		OK:          true,
		BrowserID:   browserID,
		SessionID:   sessionID,
		SessionUUID: sessionUUID,
		WSToken:     wsToken,
		RecorderURL: recorderURL,
		EnvURL:      envURL,
	}, nil
}

// generateWSToken 生成一次性 ws_token（32 字节随机，base64-url 编码）。
func generateWSToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// newRecordSessionForRecorder 预创建 record_session 行。
func newRecordSessionForRecorder(req *define.E2ERecordOpenRequest, browserID, envURL string) (int64, string, error) {
	rs := store.NewRecordSessionStore()
	name := strings.TrimSpace(req.SessionName)
	if name == "" {
		name = fmt.Sprintf("录制 %s", time.Now().Format("20060102 150405"))
	}
	sessionUUID := fmt.Sprintf("rec_%d", time.Now().UnixNano())
	id, err := rs.Create(name, sessionUUID, envURL, "", req.CaseID, req.GroupID, browserID)
	if err != nil {
		return 0, "", err
	}
	return id, sessionUUID, nil
}

// fetchSmartLinkEnvURL 查 smart_link 表取 link 字段，作为 env_url。
// 复用 controller/smart_link_item.go 中的常量 define.SmartLinkStatusNormal。
func fetchSmartLinkEnvURL(smartLinkID int) (string, error) {
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT link FROM smart_link WHERE id = ? AND status = ?`,
		smartLinkID, define.SmartLinkStatusNormal,
	).One()
	if err != nil || row == nil {
		return "", err
	}
	return cast.ToString(row["link"]), nil
}