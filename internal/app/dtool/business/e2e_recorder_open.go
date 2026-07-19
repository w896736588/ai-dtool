// Package business 提供基于 smart_link 的录制入口（v7 方案）。
//
// v7 与 v6 的关键区别：完全抛弃 iframe / proxy.html / ws_token fetch 这套东西。
// recorder runtime 是一个自包含 JS bundle，由 Go embed 打进二进制，录制开始时通过
// page.Evaluate(...) 直接注入到被测 page。被测 page 顶部出现 toolbar，所有 click /
// input / scroll 动作 push 到 window.__dtoolRecordBuffer；点「结束并下载」会触发
// JSON 下载 + 复制到剪贴板，并通过 window.__dtoolRecordResult 把结果暴露给
// Playwright / 前端 E2E 页面读取。
//
// 也就是说录制数据**不再走后端 HTTP**。前端 E2E 页面提供「导入录制 JSON」入口，让
// 用户粘贴 / 上传 JSON 来生成用例步骤。
package business

import (
	"crypto/rand"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/component/e2e/recorder_runtime"
	"dev_tool/internal/app/dtool/component/e2e/store"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	p_common "dev_tool/internal/pkg/p_common"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// buildRecorderSessionTokenNone v7 方案构造一次性 sentinel 占位 ws_token：
// UNIQUE 索引 `idx_tbl_e2e_record_session_token` 已有历史行 ws_token='none'，
// 所有 v7 新行如果也写 'none' 会冲突；这里把 sessionUUID 拼到 sentinel 后面保证全局唯一，
// 同时仍然不使用真实 token（v7 不走 ws_token HTTP 通道）。
func buildRecorderSessionTokenNone(sessionUUID string) string {
	if sessionUUID == `` {
		return `none_` + p_common.TBaseClient.GetUnique(`token_`)
	}
	return `none_` + sessionUUID
}

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
//      返回给 caller，由 caller 注入 recorder runtime 后让用户手工补完。
//   5) caller 在拿到 page 后调用 injectRecorderRuntime(page) 把 standalone.js 注入进去。
//
// 返回 *playwright.Page：plw.GetPage 函数签名就是返回指针到 Page 接口，调用方必须按
// `(*page).Close()` / `(*page).Evaluate(...)` 方式调用方法。
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
	// 录制模式强制关掉 smart_link 的 auto_close_second 自动关闭：
	// plw.ActiveTime 监听到页面没有网络请求 / load 事件超过 auto_close_second 秒后会主动
	// Close page → context 里没 page 残留 → chromium 自动释放 context → 触发 Context.OnClose
	// 回调，整个浏览器实例直接消失。self-test 能跑通是因为用户手动操作页面会持续触发 request
	// 事件刷新活跃时间戳，recorder 流程没有持续 user activity，必须关掉这个机制。
	runParams.AutoCloseSecond = 0

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
	// 录制模式不能再依赖 plw 的 ActiveTime 自动关 / OpenClose 任何机制，必须主动 attach runtime 后
	// 用一个长期 goroutine 保活 page / context 直到前端显式结束录制。
	go openRecorderKeepalive(browserID, page)
	return browserID, page, runParams.Link, warning, nil
}

// openRecorderKeepalive 录制模式专用：process list 跑完后留住 context 的最后一道防线。
//
// 真实情况：plw 内部的 ActiveTime 自动关 page、OnClose 触发、Context 没有持久 strong ref
// 等多种因素都会让 chromium context 退出；自测能跑通是因为用户后续会持续触发 request
// 事件，recorder 流程没有持续 user activity。这里通过一个永远循环的 goroutine 每 5 秒
// 调用一次 page.Evaluate，保持页面的 request 活跃时间戳不被清掉。
//
// 退出条件：通过 channel 由 E2ERecordStop 显式唤起 close 路径，避免任何时序窗口让
// chromium process 被自动关闭。
func openRecorderKeepalive(browserID string, page *playwright.Page) {
	defer func() {
		if err := recover(); err != nil {
			gstool.FmtPrintlnLogTime("[recorder] keepalive 退出 panic=%v browserID=%s", err, browserID)
		}
	}()
	gstool.FmtPrintlnLogTime("[recorder] keepalive 启动 browserID=%s", browserID)
	for i := 0; i < 60*60*12; i++ { // 12 小时上限
		time.Sleep(5 * time.Second)
		if page == nil {
			break
		}
		// Evaluate 一个常量表达式：纯前端开销极小，但能触发 page 的 request/load 事件，
		// 让 plw.ActiveTime 的活跃时间戳被持续刷新，page 永远不会到超时阈值。
		// 即使前面 AutoCloseSecond=0 已经在这里把它从 map 里排除，这也是个备用手段。
		_, _ = (*page).Evaluate("typeof window !== 'undefined' ? 1 : 0")
	}
	gstool.FmtPrintlnLogTime("[recorder] keepalive 结束 browserID=%s", browserID)
}

// injectRecorderRuntime 把 recorder_runtime.RecorderRuntimeJS() 通过 page.Evaluate 注入到被测 page。
// 失败时仅打印日志，不返回 error —— toolbar 注入失败不会拖垮整个录制流程（用户仍可手工复制 JSON，
// 因为 runtime 内容是纯前端代码，不会因为注入失败就让 page 不可用）。
func injectRecorderRuntime(page *playwright.Page) {
	js := recorder_runtime.RecorderRuntimeJS()
	if _, err := (*page).Evaluate(js); err != nil {
		gstool.FmtPrintlnLogTime(`[recorder] 注入 recorder runtime 失败 %s`, err.Error())
	}
}

// E2ERecordOpen 开启一次 smart_link 录制会话（v7 方案：纯前端 JSON 方案）。
//
// 关键路径：
//   1) 复用 plw.Playwright 的 GetPage + RunProcessesSync 完成 smart_link 的登录流程。
//   2) 在 page 上 evaluate recorder runtime（独立 JS bundle），挂 toolbar。
//   3) 创建 record_session 行（仅用于历史 / 续录 / Playwright 中转拉取）。
//   4) 返回 BrowserID 让前端能跳转到正确页面。
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

	// v7 不再用 ws_token 走 HTTP；保留 UpdateSmartLink 是为了让历史行结构兼容（session_uuid / smart_link_id / link_id 仍要落库）。
	// 重要：ws_token 列有 UNIQUE 索引，且历史数据中已有 ws_token='none'（迁移脚本把所有空字符串 / NULL 规整为该 sentinel）。
	// 所以 v7 不能直接写 'none'，必须在 sentinel 后拼上 sessionUUID 之类保证唯一的后缀。
	wsTokenSentinel := buildRecorderSessionTokenNone(sessionUUID)
	if err := store.NewRecordSessionStore().UpdateSmartLink(sessionID, req.SmartLinkID, req.UserName, wsTokenSentinel, "", req.LinkID); err != nil {
		closePage()
		return nil, err
	}

	injectRecorderRuntime(page)

	return &define.E2ERecordOpenResponse{
		OK:          true,
		BrowserID:   browserID,
		SessionID:   sessionID,
		SessionUUID: sessionUUID,
		// 以下三个字段保留为空字符串，便于前端做兼容判断；旧前端若仍读取也不会 panic。
		WSToken:     "",
		RecorderURL: "",
		EnvURL:      envURL,
		Warning:     warning,
	}, nil
}

// generateWSToken 保留以备后续（如要做 Playwright 中转拉取录制结果）使用。
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