// Package business 提供基于 smart_link + ws_token 的录制入口。
package business

import (
	"crypto/rand"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/component/e2e/store"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// recorderProxyPath recorder iframe proxy 同源路径。
// 与 router.go 的 /api/e2e/recorder/proxy.html 保持一致。
const recorderProxyPath = "/api/e2e/recorder/proxy.html"

// recorderInitScriptFmt AddInitScript 注入到被录 page 的 JS 模板。
// %q 会自动做 JSON 字符串转义，避免 ws_token 中的特殊字符破坏脚本。
// 同时兼容：
//   - DOMContentLoaded 还没触发（page 加载中）：监听器等待事件。
//   - DOMContentLoaded 已经过去（Evaluate 立即执行时）：直接挂 toolbar。
const recorderInitScriptFmt = `
(function(){
  window.__dtoolRecorder = {wsToken:%q, recorderUrl:%q, sessionUUID:%q};
  function mountRecorder(){
    var iframe = document.createElement('iframe');
    iframe.src = window.__dtoolRecorder.recorderUrl;
    iframe.style.cssText = 'position:fixed;width:1px;height:1px;opacity:0;pointer-events:none;border:0;right:0;bottom:0;';
    document.body.appendChild(iframe);
  }
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', mountRecorder);
  } else {
    // DOMContentLoaded 已经过去，立即挂；不影响后续 navigation，由 AddInitScript 重新触发
    mountRecorder();
  }
})();
`

// recorderStream 录制专用的 StreamFunc：
// - 把 plw Playwright.Open 的关键节点（构建run_params / 打开浏览器 / 登录 process）实时打到日志。
// - 与 controller/smart_link.go 中 SmartLinkRunPlaywright 走一样的套路。
func recorderStream() func(string, string) {
	return func(stage, msg string) {
		gstool.FmtPrintlnLogTime("[recorder] %s %s", stage, msg)
	}
}

// querySmartLinkLabel 复刻 controller/ai_browser.go::querySmartLinkLabel 的语义：
// 在业务层直接查 smart_link.label，避免反向依赖 controller 包。
func querySmartLinkLabel(smartLinkID int) (string, error) {
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT label FROM smart_link WHERE id = ? AND status = ?`,
		smartLinkID, define.SmartLinkStatusNormal,
	).One()
	if err != nil || row == nil {
		return "", fmt.Errorf("smart_link %d 不存在或已失效", smartLinkID)
	}
	return cast.ToString(row["label"]), nil
}

// openSmartLinkRecorder 真正复用 smart_link 的 plw 流程：
//   1) GetRunParams -> 把 smart_link 行 + process_id 列表装配成 PlaywrightRunParams。
//   2) NewPlaywright -> 创建 plw.Playwright（含 ContextPageList）。
//   3) GetPage       -> 复用带用户数据目录的 BrowserContext，NewPage + Navigate 到 link。
//   4) RunProcessesSync -> 跑登录 process 列表（输入账号 / 输入密码 / 登录 / 进入后台...）；
//      即使失败（典型情况：SPA 跳转慢于 wait_mills 但 form submit 已发起），也会把 page 一并
//      返回给 caller，由 controller 决定是否继续注入 recorder，让用户手工补完登录。
//   5) 注：不在此处 AddInitScript，由 E2ERecordOpen 在拿到 page 后注入，使 init script
//      在用户开始操作之前就被加入，但不会被登录流程破坏。
//
// 返回 *playwright.Page：plw.GetPage 函数签名就是返回指针到 Page 接口，调用方必须按
// `(*page).Close()` / `(*page).AddInitScript(...)` 方式调用方法。
func openSmartLinkRecorder(smartLinkID int, userName, password string) (string, *playwright.Page, string, string, error) {
	label, err := querySmartLinkLabel(smartLinkID)
	if err != nil {
		return "", nil, "", "", err
	}
	stream := recorderStream()
	stream(`构建run_params`, `开始`)
	// openType=2 与 controller/smart_link.go SmartLinkRunPlaywright 一致，使用内置浏览器核心；
	// 实际上 openType 决定 Channel selection，传入 0 让 GetRunParams 从 smart_link 表的 open_type 字段读。
	// replaceList 必须传非 nil map：getRunParamsFromNewTable 内部会向其写入 {user_name}/{password}，
	// 传 nil 会触发 "assignment to entry in nil map" panic。
	runParams, runErr := plw.GetRunParams(smartLinkID, label, userName, password, 0, 1, map[string]string{})
	if runErr != nil {
		stream(`构建run_params`, `失败:`+runErr.Error())
		return "", nil, "", "", runErr
	}
	stream(`构建run_params`, fmt.Sprintf(`成功 link=%s label=%s user=%s`, runParams.Link, runParams.Label, userName))
	runParams.StreamFunc = stream

	playwrightClient := plw.NewPlaywright(runParams, component.PlaywrightClient.Log)
	page, pageErr := playwrightClient.GetPage(common.GetCall())
	if pageErr != nil {
		stream(`打开浏览器实例`, `失败:`+pageErr.Error())
		return "", nil, "", "", pageErr
	}
	stream(`打开浏览器实例`, `完成，准备执行 process list`)
	warning := ``
	if procErr := playwrightClient.RunProcessesSync(page); procErr != nil {
		// 重要：不再 Close page 丢弃浏览器上下文。process list 失败通常意味着 SPA 跳转还没完成
		// 但 form submit 已经发起，cookies 已经在写入。这种情况让用户手工补完就能继续录制，
		// 比让他重头再来友好得多。
		warning = `smart_link 的 process list 部分失败（SPA 跳转慢于 wait_mills 是常见原因）：` + procErr.Error() +
			`。浏览器已经打开，请手工完成剩余步骤后再开始录制。`
		stream(`执行process`, `失败但保留 page 等用户手工补完: `+procErr.Error())
	} else {
		stream(`执行process`, `全部完成`)
	}

	browserID := fmt.Sprintf("rec_%d_%d", smartLinkID, time.Now().UnixNano())
	return browserID, page, runParams.Link, warning, nil
}

// E2ERecordOpen 开启一次 smart_link 录制会话。
// 关键路径：复用 plw.Playwright 的 GetPage + RunProcessesSync，让 smart_link 的登录/进入后台流程
// 全部跑完后再注入 recorder init script。用户最终看到的浏览器就是 smart_link 登录后的真实页面。
func E2ERecordOpen(req *define.E2ERecordOpenRequest) (*define.E2ERecordOpenResponse, error) {
	if req == nil || req.SmartLinkID <= 0 {
		return nil, errors.New("smart_link_id 必须为正数")
	}

	browserID, page, envURL, warning, err := openSmartLinkRecorder(req.SmartLinkID, req.UserName, req.Password)
	if err != nil {
		return &define.E2ERecordOpenResponse{OK: false, Error: err.Error()}, nil
	}
	closePage := func() { _ = (*page).Close() }

	sessionID, sessionUUID, err := newRecordSessionForRecorder(req, browserID, envURL)
	if err != nil {
		closePage()
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	wsToken, err := generateWSToken()
	if err != nil {
		closePage()
		return nil, err
	}
	recorderURL := recorderProxyPath
	if err := store.NewRecordSessionStore().UpdateSmartLink(sessionID, req.SmartLinkID, req.UserName, wsToken, recorderURL, req.LinkID); err != nil {
		closePage()
		return nil, err
	}

	initBody := fmt.Sprintf(recorderInitScriptFmt, wsToken, recorderURL, sessionUUID)
	if err := (*page).AddInitScript(playwright.Script{Content: &initBody}); err != nil {
		// init script 失败：仍返回 session，但前端会提示
		return &define.E2ERecordOpenResponse{
			OK:          false,
			Error:       err.Error(),
			SessionID:   sessionID,
			SessionUUID: sessionUUID,
			WSToken:     wsToken,
			RecorderURL: recorderURL,
			EnvURL:      envURL,
			BrowserID:   browserID,
			Warning:     warning,
		}, nil
	}

	// process list 部分失败时（典型：SPA 跳转慢于 wait_mills），当前 page 大概率还是登录页或
	// 半完成页。AddInitScript 只在新 navigation 才会跑，所以这里额外 Evaluate 把 toolbar
	// 立刻挂到当前 DOMContentLoaded —— 即使用户还没手工完成登录，toolbar 也先出现，避免他
	// 登录完成后还要再 reload 一次。
	if warning != `` {
		if _, evalErr := (*page).Evaluate(initBody); evalErr != nil {
			gstool.FmtPrintlnLogTime(`[recorder] 直接 evaluate 注入 toolbar 失败 %s`, evalErr.Error())
		}
	}

	return &define.E2ERecordOpenResponse{
		OK:          true,
		BrowserID:   browserID,
		SessionID:   sessionID,
		SessionUUID: sessionUUID,
		WSToken:     wsToken,
		RecorderURL: recorderURL,
		EnvURL:      envURL,
		Warning:     warning,
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

// E2ERecordResume 按 session 行（row_id）续录：清掉旧 ws_token 后调用 E2ERecordOpen 重新分配 token 并启动浏览器。
func E2ERecordResume(sessionID int64) (*define.E2ERecordOpenResponse, error) {
	if sessionID <= 0 {
		return nil, errors.New("session_id 必须为正数")
	}
	rs := store.NewRecordSessionStore()
	row, err := rs.GetByID(sessionID)
	if err != nil || row == nil {
		return nil, errors.New("会话不存在")
	}
	req := &define.E2ERecordOpenRequest{
		SmartLinkID: cast.ToInt(row["smart_link_id"]),
		LinkID:      cast.ToInt(row["link_id"]),
		UserName:    cast.ToString(row["user_name"]),
		SessionName: cast.ToString(row["name"]),
		GroupID:     cast.ToInt(row["group_id"]),
		CaseID:      cast.ToInt(row["case_id"]),
	}
	if err := rs.UpdateWSToken(sessionID, ""); err != nil {
		return nil, err
	}
	return E2ERecordOpen(req)
}
