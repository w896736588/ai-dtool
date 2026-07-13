package assertion

import (
	"dev_tool/internal/app/dtool/component/e2e/interceptor"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// APIResponseV1Asserter assert_api_response_v1：基于已捕获请求的响应断言。
type APIResponseV1Asserter struct {
	matcher *interceptor.Matcher
}

func NewAPIResponseV1Asserter() *APIResponseV1Asserter {
	return &APIResponseV1Asserter{matcher: interceptor.NewMatcher()}
}

func (a *APIResponseV1Asserter) Type() define.E2EAssertionType { return define.E2EAssertAPIResponseV1 }

func (a *APIResponseV1Asserter) Validate(assertion *define.E2EAssertion) error {
	var cfg define.APIResponseAssertionV1Config
	if err := json.Unmarshal(assertion.Config, &cfg); err != nil {
		return err
	}
	if cfg.FindByURL == "" && cfg.FindByPattern == "" && cfg.FindByMethod == "" && cfg.FindByResponseContains == "" {
		return errors.New("assert_api_response_v1: 至少需要 1 个查找条件")
	}
	return nil
}

func (a *APIResponseV1Asserter) Assert(ctx *AssertionContext, assertion *define.E2EAssertion) *AssertionResult {
	var cfg define.APIResponseAssertionV1Config
	_ = json.Unmarshal(assertion.Config, &cfg)

	matchCfg := &interceptor.MatchConfig{
		URL:              cfg.FindByURL,
		Contains:         cfg.FindByPattern,
		Method:           cfg.FindByMethod,
		ResponseContains: cfg.FindByResponseContains,
	}
	if matchCfg.Contains != "" && !strings.Contains(matchCfg.Contains, "*") {
		// 简单纯字符串保持 contains；通配符交给 matchWildcard
	}
	if matchCfg.Contains != "" && strings.Contains(matchCfg.Contains, "*") {
		matchCfg.Regex = matchCfg.Contains
		matchCfg.Contains = ""
	}

	matched := a.matcher.Match(ctx.RequestRepo, matchCfg)
	if len(matched) == 0 {
		return &AssertionResult{
			Success:  false,
			Expected: fmt.Sprintf("URL: %s, Method: %s", cfg.FindByURL, cfg.FindByMethod),
			ErrorMsg: "未找到匹配的捕获请求",
		}
	}

	targetList := matched
	if cfg.MatchIndex > 0 && cfg.MatchIndex <= len(matched) {
		targetList = []*interceptor.CapturedRequest{matched[cfg.MatchIndex-1]}
	}

	if cfg.MatchAll {
		for _, req := range targetList {
			r := a.assertOne(req, &cfg)
			if !r.Success {
				return r
			}
		}
		return &AssertionResult{Success: true}
	}

	for _, req := range targetList {
		r := a.assertOne(req, &cfg)
		if r.Success {
			ctx.RequestRepo.MarkMatched(req.ID, assertion.ID)
			r.MatchedURL = req.URL
			r.MatchedReqID = req.ID
			return r
		}
	}
	// 全部失败：返回第一个的失败结果
	r := a.assertOne(targetList[0], &cfg)
	r.MatchedURL = targetList[0].URL
	r.MatchedReqID = targetList[0].ID
	return r
}

func (a *APIResponseV1Asserter) assertOne(req *interceptor.CapturedRequest, cfg *define.APIResponseAssertionV1Config) *AssertionResult {
	if req.Response == nil {
		return &AssertionResult{
			Success:  false,
			Expected: "响应非空",
			Actual:   "无响应（请求可能还在进行中）",
			ErrorMsg: "捕获请求还未收到响应",
		}
	}
	if cfg.ResponseStatus != 0 && req.Response.Status != cfg.ResponseStatus {
		return &AssertionResult{
			Success:  false,
			Expected: fmt.Sprintf("status=%d", cfg.ResponseStatus),
			Actual:   fmt.Sprintf("status=%d", req.Response.Status),
			ErrorMsg: "响应状态码不匹配",
		}
	}
	if cfg.ResponseContains != "" {
		body := req.Response.Body
		if cfg.IgnoreCase {
			if !strings.Contains(strings.ToLower(body), strings.ToLower(cfg.ResponseContains)) {
				return &AssertionResult{
					Success:  false,
					Expected: fmt.Sprintf("contains: %s", cfg.ResponseContains),
					Actual:   truncate(body, 500),
					ErrorMsg: "响应体不包含预期文本",
				}
			}
		} else if !strings.Contains(body, cfg.ResponseContains) {
			return &AssertionResult{
				Success:  false,
				Expected: fmt.Sprintf("contains: %s", cfg.ResponseContains),
				Actual:   truncate(body, 500),
				ErrorMsg: "响应体不包含预期文本",
			}
		}
	}
	if cfg.ResponseJSONPath != "" && cfg.ExpectedValue != nil {
		var data any
		if err := json.Unmarshal([]byte(req.Response.Body), &data); err != nil {
			return &AssertionResult{
				Success:  false,
				Expected: fmt.Sprintf("JSON path %s = %v", cfg.ResponseJSONPath, cfg.ExpectedValue),
				Actual:   "响应不是合法 JSON",
				ErrorMsg: "JSON 解析失败",
			}
		}
		got := interceptor.ExtractJSONPath(data, cfg.ResponseJSONPath)
		want := fmt.Sprintf("%v", cfg.ExpectedValue)
		if got != want {
			// 数字比较兼容
			if gf, gerr := strconv.ParseFloat(got, 64); gerr == nil {
				if wf, werr := strconv.ParseFloat(want, 64); werr == nil {
					if gf == wf {
						return &AssertionResult{Success: true, Expected: want, Actual: got}
					}
				}
			}
			return &AssertionResult{
				Success:  false,
				Expected: fmt.Sprintf("JSON path %s = %s", cfg.ResponseJSONPath, want),
				Actual:   fmt.Sprintf("JSON path %s = %s", cfg.ResponseJSONPath, got),
				ErrorMsg: "JSON 路径值不匹配",
			}
		}
	}
	return &AssertionResult{Success: true}
}

// APIRequestV1Asserter assert_api_request_v1：检查发出请求的请求体 / 请求头。
type APIRequestV1Asserter struct{}

func (a *APIRequestV1Asserter) Type() define.E2EAssertionType { return define.E2EAssertAPIRequestV1 }

func (a *APIRequestV1Asserter) Validate(assertion *define.E2EAssertion) error {
	var cfg define.APIRequestAssertionV1Config
	if err := json.Unmarshal(assertion.Config, &cfg); err != nil {
		return err
	}
	if cfg.FindByURL == "" && cfg.FindByPattern == "" && cfg.FindByMethod == "" {
		return errors.New("assert_api_request_v1: 至少需要 1 个查找条件")
	}
	return nil
}

func (a *APIRequestV1Asserter) Assert(ctx *AssertionContext, assertion *define.E2EAssertion) *AssertionResult {
	var cfg define.APIRequestAssertionV1Config
	_ = json.Unmarshal(assertion.Config, &cfg)
	matcher := interceptor.NewMatcher()
	matchCfg := &interceptor.MatchConfig{
		URL:      cfg.FindByURL,
		Contains: cfg.FindByPattern,
		Method:   cfg.FindByMethod,
	}
	if matchCfg.Contains != "" && strings.Contains(matchCfg.Contains, "*") {
		matchCfg.Regex = matchCfg.Contains
		matchCfg.Contains = ""
	}
	matched := matcher.Match(ctx.RequestRepo, matchCfg)
	if len(matched) == 0 {
		return &AssertionResult{
			Success:  false,
			Expected: fmt.Sprintf("URL=%s Method=%s", cfg.FindByURL, cfg.FindByMethod),
			ErrorMsg: "未找到匹配的请求",
		}
	}
	targetList := matched
	if cfg.MatchIndex > 0 && cfg.MatchIndex <= len(matched) {
		targetList = []*interceptor.CapturedRequest{matched[cfg.MatchIndex-1]}
	}
	for _, req := range targetList {
		if cfg.RequestHeaderName != "" {
			value, ok := lookupHeader(req.Headers, cfg.RequestHeaderName)
			if !ok || !strings.EqualFold(value, cfg.RequestHeaderValue) {
				continue
			}
		}
		if cfg.RequestBodyJSONPath != "" {
			var data any
			if err := json.Unmarshal([]byte(req.PostData), &data); err != nil {
				continue
			}
			got := interceptor.ExtractJSONPath(data, cfg.RequestBodyJSONPath)
			want := fmt.Sprintf("%v", cfg.RequestBodyExpected)
			if got != want {
				continue
			}
		}
		ctx.RequestRepo.MarkMatched(req.ID, assertion.ID)
		return &AssertionResult{Success: true, MatchedURL: req.URL, MatchedReqID: req.ID}
	}
	return &AssertionResult{
		Success:  false,
		Expected: fmt.Sprintf("%+v", cfg),
		Actual:   fmt.Sprintf("%d 个候选请求均不匹配", len(targetList)),
		ErrorMsg: "请求断言失败",
	}
}

func lookupHeader(headers map[string]string, name string) (string, bool) {
	for k, v := range headers {
		if strings.EqualFold(k, name) {
			return v, true
		}
	}
	return "", false
}
