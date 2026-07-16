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
		stepReg:   step_executor.NewRegistry(),
		assertReg: assertion.NewRegistry(),
		caseStore: store.NewCaseStore(),
		runStore:  store.NewRunStore(),
		stepStore: store.NewStepStore(),
		reqStore:  store.NewRequestStore(),
		runStates: make(map[int64]*E2ERunState),
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
