package plw

import (
	"testing"
	"time"
)

// TestDecodeLocatorConfig 验证新版 locator config 能被正确解析。
// TestDecodeLocatorConfig verifies new locator configs can be decoded.
func TestDecodeLocatorConfig(t *testing.T) {
	raw := `{"version":2,"mode":"bool_result","strategy":"first_match_return","locators":[{"id":"rule_1","query":{"spec":{"method":"locator","value":".user-info"}},"on_found":true}]}`

	config, err := decodeLocatorConfig(raw)
	if err != nil {
		t.Fatalf("decodeLocatorConfig() error = %v", err)
	}
	if config.Mode != `bool_result` || config.Strategy != `first_match_return` {
		t.Fatalf("config = %#v, want bool_result/first_match_return", config)
	}
	boolValue, ok := readBoolOnFound(config.Locators[0].OnFound)
	if len(config.Locators) != 1 || !ok || !boolValue {
		t.Fatalf("config.Locators = %#v, want first locator on_found=true", config.Locators)
	}
}

// TestRunLocatorConfigBoolResult 验证 bool_result 会在第一条命中规则时返回对应布尔值。
// TestRunLocatorConfigBoolResult verifies bool_result returns the configured boolean on first match.
func TestRunLocatorConfigBoolResult(t *testing.T) {
	runner := &LocatorConfigRunner{
		runQuery: func(item LocatorConfigItem, action *ElementAction) (*ElementResult, error) {
			if item.Query != nil && item.Query.Spec != nil && item.Query.Spec.Value == `.need-login` {
				return &ElementResult{Exists: true}, nil
			}
			return nil, errLocatorConfigNotFound
		},
	}

	result, err := runner.Run(&LocatorConfig{
		Version:  2,
		Mode:     `bool_result`,
		Strategy: `first_match_return`,
		Locators: []LocatorConfigItem{
			{
				ID:      `rule_1`,
				Query:   &LocatorInput{Spec: &LocatorSpec{Method: `locator`, Value: `.need-login`}},
				OnFound: true,
			},
		},
	}, ``)
	if err != nil {
		t.Fatalf("Run(bool_result) error = %v", err)
	}
	if result.BoolValue == nil || !*result.BoolValue {
		t.Fatalf("result.BoolValue = %#v, want true", result.BoolValue)
	}
}

// TestRunLocatorConfigTextContent 验证文本提取规则可返回提取文本。
// TestRunLocatorConfigTextContent verifies text rules can return extracted text.
func TestRunLocatorConfigTextContent(t *testing.T) {
	runner := &LocatorConfigRunner{
		runQuery: func(item LocatorConfigItem, action *ElementAction) (*ElementResult, error) {
			if action.Type == `text_content` {
				return &ElementResult{TextContent: `欢迎回来`}, nil
			}
			return nil, errLocatorConfigNotFound
		},
	}

	result, err := runner.Run(&LocatorConfig{
		Version:  2,
		Mode:     `text_content`,
		Strategy: `first_match_return`,
		Locators: []LocatorConfigItem{
			{ID: `a`, OnFound: `extract_text`, Query: &LocatorInput{Spec: &LocatorSpec{Method: `locator`, Value: `.content`}}},
			{ID: `b`, OnFound: `return_empty`, Query: &LocatorInput{Spec: &LocatorSpec{Method: `locator`, Value: `.empty`}}},
		},
	}, ``)
	if err != nil {
		t.Fatalf("Run(text_content) error = %v", err)
	}
	if result.TextValue != `欢迎回来` {
		t.Fatalf("result.TextValue = %q, want 欢迎回来", result.TextValue)
	}
}

// TestRunLocatorConfigTextContentReturnsSourceImmediately 验证 extract_text 先返回时会直接结束。
// TestRunLocatorConfigTextContentReturnsSourceImmediately verifies extract_text can finish immediately.
func TestRunLocatorConfigTextContentReturnsSourceImmediately(t *testing.T) {
	emptyStartedCh := make(chan struct{}, 1)
	emptyReleaseCh := make(chan struct{})
	runner := &LocatorConfigRunner{
		runQuery: func(item LocatorConfigItem, action *ElementAction) (*ElementResult, error) {
			if action.Type == `exist` {
				emptyStartedCh <- struct{}{}
				<-emptyReleaseCh
				return &ElementResult{Exists: true}, nil
			}
			return &ElementResult{TextContent: `立即返回文本`}, nil
		},
	}

	doneCh := make(chan *LocatorConfigRunResult, 1)
	errCh := make(chan error, 1)
	go func() {
		result, err := runner.Run(&LocatorConfig{
			Version:  2,
			Mode:     `text_content`,
			Strategy: `first_match_return`,
			Locators: []LocatorConfigItem{
				{ID: `a`, OnFound: `extract_text`, Query: &LocatorInput{Spec: &LocatorSpec{Method: `locator`, Value: `.content`}}},
				{ID: `b`, OnFound: `return_empty`, Query: &LocatorInput{Spec: &LocatorSpec{Method: `locator`, Value: `.empty`}}},
			},
		}, ``)
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- result
	}()

	<-emptyStartedCh

	select {
	case err := <-errCh:
		close(emptyReleaseCh)
		t.Fatalf("Run(text_content) error = %v", err)
	case result := <-doneCh:
		close(emptyReleaseCh)
		if result == nil || result.TextValue != `立即返回文本` {
			t.Fatalf("result = %#v, want text value 立即返回文本", result)
		}
	case <-time.After(200 * time.Millisecond):
		close(emptyReleaseCh)
		t.Fatal("Run(text_content) should return after extract_text result without waiting for return_empty")
	}
}

// TestRunLocatorConfigTextContentReturnsEmptyImmediately 验证 return_empty 先返回时会直接返回空字符串。
// TestRunLocatorConfigTextContentReturnsEmptyImmediately verifies return_empty can finish immediately.
func TestRunLocatorConfigTextContentReturnsEmptyImmediately(t *testing.T) {
	textStartedCh := make(chan struct{}, 1)
	textReleaseCh := make(chan struct{})
	runner := &LocatorConfigRunner{
		runQuery: func(item LocatorConfigItem, action *ElementAction) (*ElementResult, error) {
			if action.Type == `text_content` {
				textStartedCh <- struct{}{}
				<-textReleaseCh
				return &ElementResult{TextContent: `不应返回`}, nil
			}
			return &ElementResult{Exists: true}, nil
		},
	}

	doneCh := make(chan *LocatorConfigRunResult, 1)
	errCh := make(chan error, 1)
	go func() {
		result, err := runner.Run(&LocatorConfig{
			Version:  2,
			Mode:     `text_content`,
			Strategy: `first_match_return`,
			Locators: []LocatorConfigItem{
				{ID: `a`, OnFound: `extract_text`, Query: &LocatorInput{Spec: &LocatorSpec{Method: `locator`, Value: `.content`}}},
				{ID: `b`, OnFound: `return_empty`, Query: &LocatorInput{Spec: &LocatorSpec{Method: `locator`, Value: `.empty`}}},
			},
		}, ``)
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- result
	}()

	<-textStartedCh

	select {
	case err := <-errCh:
		close(textReleaseCh)
		t.Fatalf("Run(text_content) error = %v", err)
	case result := <-doneCh:
		close(textReleaseCh)
		if result == nil || result.TextValue != `` {
			t.Fatalf("result = %#v, want empty text value", result)
		}
	case <-time.After(200 * time.Millisecond):
		close(textReleaseCh)
		t.Fatal("Run(text_content) should return empty string after return_empty result without waiting for extract_text")
	}
}

// TestRunLocatorConfigAction 验证 click/input 会命中任意一个后执行动作。
// TestRunLocatorConfigAction verifies click/input act on the first matched locator.
func TestRunLocatorConfigAction(t *testing.T) {
	var calledAction string
	var calledValue string
	runner := &LocatorConfigRunner{
		runQuery: func(item LocatorConfigItem, action *ElementAction) (*ElementResult, error) {
			if item.Query != nil && item.Query.Spec != nil && item.Query.Spec.Value == `.confirm-btn` {
				calledAction = action.Type
				calledValue = action.Value
				return &ElementResult{}, nil
			}
			return nil, errLocatorConfigNotFound
		},
	}

	_, err := runner.Run(&LocatorConfig{
		Version:  2,
		Mode:     `input`,
		Strategy: `first_found_do_action`,
		Locators: []LocatorConfigItem{
			{ID: `a`, Query: &LocatorInput{Spec: &LocatorSpec{Method: `locator`, Value: `.confirm-btn`}}},
		},
		Options: &LocatorConfigOptions{ActionType: `input`},
	}, `frog`)
	if err != nil {
		t.Fatalf("Run(input) error = %v", err)
	}
	if calledAction != `input` || calledValue != `frog` {
		t.Fatalf("calledAction/value = %q/%q, want input/frog", calledAction, calledValue)
	}
}
