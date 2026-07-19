// E2E 执行引擎（属于 business 包）。
package business

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/component/e2e/assertion"
	"dev_tool/internal/app/dtool/component/e2e/interceptor"
	"dev_tool/internal/app/dtool/component/e2e/store"
	"dev_tool/internal/app/dtool/component/e2e/step_executor"
	"dev_tool/internal/app/dtool/component/e2e/variable"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

// E2EEngine 执行引擎。
type E2EEngine struct {
	stepReg     *step_executor.Registry
	assertReg   *assertion.Registry
	caseStore   *store.CaseStore
	runStore    *store.RunStore
	stepStore   *store.StepStore
	reqStore    *store.RequestStore

	mu        sync.Mutex
	runStates map[int64]*E2ERunState

	// v6 录制入口按 smart_link + browserID 索引 page，替代旧的 lastPage。
	recPageMu    sync.Mutex
	recorderPages map[string]recordedPage
}

// E2ERunState 正在运行的执行实例。
type E2ERunState struct {
	RunID    int64
	CaseID   int
	StopFlag bool
	Output   *step_executor.OutputBuffer
}

// NewE2EEngine 创建并注册所有内置步骤/断言处理器。
func NewE2EEngine() *E2EEngine {
	e := &E2EEngine{
		stepReg:      step_executor.NewRegistry(),
		assertReg:    assertion.NewRegistry(),
		caseStore:    store.NewCaseStore(),
		runStore:     store.NewRunStore(),
		stepStore:    store.NewStepStore(),
		reqStore:     store.NewRequestStore(),
		runStates:    make(map[int64]*E2ERunState),
		recorderPages: make(map[string]recordedPage),
	}
	e.registerDefaultSteps()
	e.registerDefaultAsserters()
	return e
}

func (e *E2EEngine) registerDefaultSteps() {
	e.stepReg.Register(&step_executor.OpenEnvExecutor{})
	e.stepReg.Register(&step_executor.ClickV1Executor{})
	e.stepReg.Register(&step_executor.InputV1Executor{})
	e.stepReg.Register(&step_executor.InputV2Executor{})   // 多种输入源
	e.stepReg.Register(&step_executor.HoverV1Executor{})
	e.stepReg.Register(&step_executor.SelectV1Executor{})
	e.stepReg.Register(&step_executor.NavigateV1Executor{})
	e.stepReg.Register(&step_executor.WaitElementV1Executor{})
	e.stepReg.Register(&step_executor.WaitTimeoutV1Executor{})
	e.stepReg.Register(&step_executor.ExtractTextV1Executor{})
	e.stepReg.Register(&step_executor.ExtractAttrV1Executor{})
	e.stepReg.Register(&step_executor.ExtractAPIV1Executor{}) // 从捕获的 API 响应提取
	e.stepReg.Register(&step_executor.ScriptV1Executor{})
	e.stepReg.Register(&step_executor.GoBackV1Executor{})
	e.stepReg.Register(&step_executor.ReloadV1Executor{})
	e.stepReg.Register(&step_executor.PressKeyV1Executor{})
	// v5.0 录制专用步骤
	e.stepReg.Register(&step_executor.ClickByPositionV1Executor{})
	e.stepReg.Register(&step_executor.RightClickV1Executor{})
	e.stepReg.Register(&step_executor.ScrollV1Executor{})
	e.stepReg.Register(&step_executor.WaitAfterV1Executor{})
}

func (e *E2EEngine) registerDefaultAsserters() {
	e.assertReg.Register(&assertion.TextV1Asserter{})
	e.assertReg.Register(&assertion.TextV2Asserter{})
	e.assertReg.Register(&assertion.ElementV1Asserter{})
	e.assertReg.Register(&assertion.URLV1Asserter{})
	e.assertReg.Register(&assertion.TitleV1Asserter{})
	e.assertReg.Register(&assertion.VariableV1Asserter{})
	e.assertReg.Register(assertion.NewAPIResponseV1Asserter())
	e.assertReg.Register(&assertion.APIRequestV1Asserter{})
}

// StepRegistry 步骤注册表（暴露给测试 / 业务）。
func (e *E2EEngine) StepRegistry() *step_executor.Registry { return e.stepReg }

// AssertionRegistry 断言注册表。
func (e *E2EEngine) AssertionRegistry() *assertion.Registry { return e.assertReg }

// NewStringOutput 创建一个新的 OutputBuffer（用于录制单步 / 整段回放）。
func (e *E2EEngine) NewStringOutput() *step_executor.OutputBuffer {
	return &step_executor.OutputBuffer{Lines: []string{}}
}

// recordedPage v6 录制入口缓存的 page + 元数据（按 browserID 索引）。
type recordedPage struct {
	page        playwright.Page
	smartLinkID int
	userName    string
	createdAt   time.Time
}

// OpenRecorder 基于 smart_link 打开一个 Playwright page，并缓存到 recorderPages。
// 流程不再依赖 plw.Playwright.GetPage，避免 PlaywrightRunParams / *p_common.Call 强耦合。
// smartLinkID 必须为正数；userName 可空。
func (e *E2EEngine) OpenRecorder(smartLinkID int, userName string) (string, playwright.Page, error) {
	if smartLinkID <= 0 {
		return "", nil, errors.New("smart_link_id 必须为正数")
	}
	browser := component.PlaywrightClient.BrowserWebkitChrome
	if browser == nil {
		browser = component.PlaywrightClient.BrowserWebkitSilence
	}
	if browser == nil {
		return "", nil, errors.New("Playwright 浏览器未启动，请先安装核心")
	}
	page, err := browser.NewPage()
	if err != nil {
		return "", nil, fmt.Errorf("NewPage 失败: %w", err)
	}
	browserID := fmt.Sprintf("rec_%d_%d", smartLinkID, time.Now().UnixNano())
	e.recPageMu.Lock()
	e.recorderPages[browserID] = recordedPage{
		page:        page,
		smartLinkID: smartLinkID,
		userName:    userName,
		createdAt:   time.Now(),
	}
	e.recPageMu.Unlock()
	return browserID, page, nil
}

// GetRecorderPage 根据 browserID 取回之前 OpenRecorder 缓存的 page。
func (e *E2EEngine) GetRecorderPage(browserID string) (playwright.Page, error) {
	e.recPageMu.Lock()
	defer e.recPageMu.Unlock()
	rp, ok := e.recorderPages[browserID]
	if !ok || rp.page == nil {
		return nil, fmt.Errorf("未找到 recorder page: %s", browserID)
	}
	return rp.page, nil
}

// ---- v6 已移除的旧 API 留 stub，等待 controller/router 改造（任务 10） ----

// TODO(removed-by-v6-refactor): OpenRecorderBrowser 已被 E2ERecordOpen 取代。
// 保留 stub 仅用于兼容 controller/E2ERecordOpenBrowser 的临时调用，主代理应在任务 10 删除。
func (e *E2EEngine) OpenRecorderBrowser(envURL string) (int64, error) {
	_ = envURL
	return 0, errors.New("OpenRecorderBrowser 已废弃，请使用 E2ERecordOpen + smart_link")
}

// TODO(removed-by-v6-refactor): GetBrowserPage 已被 GetRecorderPage(browserID) 取代。
func (e *E2EEngine) GetBrowserPage(browserID string) (playwright.Page, error) {
	_ = browserID
	return nil, nil
}

// TODO(removed-by-v6-refactor): GetAnyPage 已被 GetRecorderPage(browserID) 取代。
func (e *E2EEngine) GetAnyPage() playwright.Page {
	return nil
}

// TODO(removed-by-v6-refactor): SetLastPage 已不再需要，page 通过 recorderPages 直接索引。
func (e *E2EEngine) SetLastPage(p playwright.Page) { _ = p }

// TODO(removed-by-v6-refactor): ExecuteStepForTest 旧录制回放暴露方法，待任务 10 移除调用。
func (e *E2EEngine) ExecuteStepForTest(ctx *step_executor.ExecuteContext, step define.E2EStep) *step_executor.StepResult {
	return &step_executor.StepResult{Success: false, ErrorMsg: "ExecuteStepForTest 已废弃"}
}

// TODO(removed-by-v6-refactor): ApplyPostStepWaitForTest 旧录制回放暴露方法，待任务 10 移除调用。
func (e *E2EEngine) ApplyPostStepWaitForTest(step define.E2EStep, ctx *step_executor.ExecuteContext) {
	_ = step
	_ = ctx
}

// RunCase 异步执行用例。
func (e *E2EEngine) RunCase(caseID int, triggerType string) (int64, error) {
	caseRow, err := e.caseStore.Get(caseID)
	if err != nil {
		return 0, err
	}
	if len(caseRow) == 0 {
		return 0, fmt.Errorf("用例不存在: %d", caseID)
	}

	steps := e2eParseStepArray(caseRow["steps"])
	assertions := e2eParseAssertionArray(caseRow["assertions"])
	variables := e2eParseVarMap(caseRow["variables"])
	timeoutSeconds := e2eToInt(caseRow["timeout_seconds"])
	if timeoutSeconds <= 0 {
		timeoutSeconds = 600
	}

	runID, err := e.runStore.Create(caseID, e2eToInt(caseRow["group_id"]), len(steps), triggerType)
	if err != nil {
		return 0, err
	}

	state := &E2ERunState{RunID: runID, CaseID: caseID, Output: &step_executor.OutputBuffer{}}
	e.mu.Lock()
	e.runStates[runID] = state
	e.mu.Unlock()

	go e.runAsync(state, caseRow, steps, assertions, variables, timeoutSeconds)
	return runID, nil
}

// StopRun 标记停止。
func (e *E2EEngine) StopRun(runID int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if state, ok := e.runStates[runID]; ok {
		state.StopFlag = true
		return nil
	}
	return errors.New("执行实例未在运行中")
}

func (e *E2EEngine) runAsync(state *E2ERunState, caseRow map[string]any,
	steps []define.E2EStep, assertions []define.E2EAssertion,
	initialVars map[string]string, timeoutSeconds int) {

	defer func() {
		e.mu.Lock()
		delete(e.runStates, state.RunID)
		e.mu.Unlock()
	}()

	startedAtMs := time.Now().UnixMilli()
	_ = e.runStore.UpdateStatus(state.RunID, define.E2ERunStatusRunning, startedAtMs, nil)

	envURL := e2eToStr(caseRow["env_url"])
	envBaseURL := e2eToStr(caseRow["env_base_url"])

	varCtx := variable.NewContext(initialVars)
	resolver := variable.NewResolver(varCtx)
	repo := interceptor.NewRequestRepository()

	browser := component.PlaywrightClient.BrowserWebkitChrome
	if browser == nil {
		browser = component.PlaywrightClient.BrowserWebkitSilence
	}
	if browser == nil {
		errMsg := "Playwright 浏览器未启动，请先安装并启动浏览器核心"
		_ = e.runStore.UpdateStatus(state.RunID, define.E2ERunStatusError, startedAtMs, map[string]any{
			"error_message": errMsg,
		})
		state.Output.Writef("[engine] " + errMsg)
		return
	}

	page, err := browser.NewPage()
	if err != nil {
		_ = e.runStore.UpdateStatus(state.RunID, define.E2ERunStatusError, startedAtMs, map[string]any{
			"error_message": "创建 page 失败: " + err.Error(),
		})
		return
	}
	defer func() { _ = page.Close() }()

	// 监听器在构造时自动注册
	_ = NewE2ERequestCatcher(page, repo)

	execCtx := &step_executor.ExecuteContext{
		RunID:       state.RunID,
		CaseID:      state.CaseID,
		EnvURL:      envURL,
		EnvBaseURL:  envBaseURL,
		Page:        page,
		Browser:     browser,
		VarContext:  varCtx,
		Resolver:    resolver,
		RequestRepo: repo,
		Output:      state.Output,
	}

	timeoutCh := time.After(time.Duration(timeoutSeconds) * time.Second)

	for i, step := range steps {
		select {
		case <-timeoutCh:
			_ = e.runStore.UpdateStatus(state.RunID, define.E2ERunStatusStopped, startedAtMs, map[string]any{
				"error_message": "用例执行超时",
			})
			state.Output.Writef("[engine] 执行超时")
			e.flushCapturedRequests(state.RunID, repo)
			return
		default:
		}
		if state.StopFlag {
			_ = e.runStore.UpdateStatus(state.RunID, define.E2ERunStatusStopped, startedAtMs, map[string]any{
				"error_message": "用户取消",
			})
			state.Output.Writef("[engine] 用户停止执行")
			e.flushCapturedRequests(state.RunID, repo)
			return
		}

		stepStart := time.Now()
		stepResult := e.executeStep(execCtx, step)
		stepDuration := time.Since(stepStart).Milliseconds()

		runStepDBID, _ := e.stepStore.CreateStep(
			state.RunID, i, step.ID, string(step.Type), e2eFirstNonEmpty(step.Version, "1.0"),
			string(step.Config), step.Description,
			e2eBoolToStatus(stepResult.Success),
			stepResult.ErrorMsg, stepResult.Screenshot,
			stepDuration,
		)
		execCtx.RunStepDBID = int(runStepDBID)

		if stepResult.Success {
			_ = e.runStore.IncrementStepStats(state.RunID, 1, 0)
		} else {
			_ = e.runStore.IncrementStepStats(state.RunID, 0, 1)
		}

		// 录制场景下，步骤可能带有 wait_after_ms 字段（来自前端步骤确认弹窗）。
		// 也支持 wait_after_v1 步骤类型（更通用）。
		if stepResult.Success {
			e.applyPostStepWait(step, execCtx)
		}

		if !stepResult.Success {
			_ = e.runStore.UpdateStatus(state.RunID, define.E2ERunStatusFailed, startedAtMs, map[string]any{
				"error_message": stepResult.ErrorMsg,
			})
			state.Output.Writef("[step] %s 失败: %s", step.Type, stepResult.ErrorMsg)
			e.flushCapturedRequests(state.RunID, repo)
			return
		}

		for _, ass := range assertions {
			if ass.StepID != step.ID {
				continue
			}
			assResult := e.executeAssertion(&assertion.AssertionContext{
				Page:        page,
				RequestRepo: repo,
				VarContext:  varCtx,
				Resolver:    resolver,
			}, ass)

			_, _ = e.stepStore.CreateAssertion(
				state.RunID, runStepDBID, ass.ID, string(ass.Type),
				e2eFirstNonEmpty(ass.Version, "1.0"), string(ass.Config),
				e2eBoolToStatus(assResult.Success),
				assResult.Expected, assResult.Actual, assResult.ErrorMsg,
				assResult.MatchedURL, assResult.MatchedReqID,
			)
			if assResult.Success {
				_ = e.runStore.IncrementAssertStats(state.RunID, 1, 0)
			} else {
				_ = e.runStore.IncrementAssertStats(state.RunID, 0, 1)
				_ = e.runStore.UpdateStatus(state.RunID, define.E2ERunStatusFailed, startedAtMs, map[string]any{
					"error_message": "断言失败: " + ass.Description,
				})
				state.Output.Writef("[assertion] %s 失败: %s", ass.Type, assResult.ErrorMsg)
				e.flushCapturedRequests(state.RunID, repo)
				return
			}
		}
	}

	e.flushCapturedRequests(state.RunID, repo)
	_ = e.runStore.UpdateStatus(state.RunID, define.E2ERunStatusPassed, startedAtMs, nil)
	state.Output.Writef("[engine] 执行完成")
}

func (e *E2EEngine) executeStep(ctx *step_executor.ExecuteContext, step define.E2EStep) *step_executor.StepResult {
	exec, ok := e.stepReg.Get(step.Type)
	if !ok {
		return &step_executor.StepResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("未注册的步骤类型: %s", step.Type),
		}
	}
	if err := exec.Validate(&step); err != nil {
		return &step_executor.StepResult{Success: false, ErrorMsg: err.Error()}
	}
	resolved := step
	resolved.Config = step_executor.ResolveVariablesInConfig(step.Config, ctx.Resolver)
	return exec.Execute(ctx, &resolved)
}

func (e *E2EEngine) executeAssertion(ctx *assertion.AssertionContext, ass define.E2EAssertion) *assertion.AssertionResult {
	a, ok := e.assertReg.Get(ass.Type)
	if !ok {
		return &assertion.AssertionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("未注册的断言类型: %s", ass.Type),
		}
	}
	if err := a.Validate(&ass); err != nil {
		return &assertion.AssertionResult{Success: false, ErrorMsg: err.Error()}
	}
	return a.Assert(ctx, &ass)
}

// flushCapturedRequests 批量写入捕获请求。
func (e *E2EEngine) flushCapturedRequests(runID int64, repo *interceptor.RequestRepository) {
	for _, req := range repo.GetAll() {
		req := req
		_ = e.reqStore.Insert(req, runID, 0)
	}
}

// ---- 辅助 ----

func e2eParseStepArray(raw any) []define.E2EStep {
	if raw == nil {
		return nil
	}
	str := e2eToStr(raw)
	if str == "" || str == "[]" {
		return nil
	}
	var steps []define.E2EStep
	if err := json.Unmarshal([]byte(str), &steps); err != nil {
		return nil
	}
	return steps
}

func e2eParseAssertionArray(raw any) []define.E2EAssertion {
	if raw == nil {
		return nil
	}
	str := e2eToStr(raw)
	if str == "" || str == "[]" {
		return nil
	}
	var arr []define.E2EAssertion
	if err := json.Unmarshal([]byte(str), &arr); err != nil {
		return nil
	}
	return arr
}

func e2eParseVarMap(raw any) map[string]string {
	if raw == nil {
		return nil
	}
	str := e2eToStr(raw)
	if str == "" || str == "{}" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(str), &m); err != nil {
		return nil
	}
	return m
}

func e2eFirstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func e2eBoolToStatus(b bool) string {
	if b {
		return define.E2ERunStatusPassed
	}
	return define.E2ERunStatusFailed
}

// applyPostStepWait 步骤执行后等待：支持录制场景下用户在步骤确认弹窗设置的 wait_after_ms。
// 实现：直接在当前 page 上调用 waitForTimeout，避免增加一个 wait_after_v1 子步骤。
func (e *E2EEngine) applyPostStepWait(step define.E2EStep, execCtx *step_executor.ExecuteContext) {
	dur := 0
	if step.WaitAfterMs > 0 {
		dur = step.WaitAfterMs
	} else {
		// 兼容 wait_after_v1 步骤类型（用户手工拼接）
		var cfg define.WaitAfterV1Config
		if step.Type == define.E2EStepWaitAfterV1 {
			_ = json.Unmarshal(step.Config, &cfg)
			dur = cfg.DurationMs
		}
	}
	if dur <= 0 {
		return
	}
	if execCtx == nil || execCtx.Page == nil {
		return
	}
	execCtx.Page.WaitForTimeout(float64(dur))
	execCtx.Output.Writef("[step] wait_after %dms", dur)
}
