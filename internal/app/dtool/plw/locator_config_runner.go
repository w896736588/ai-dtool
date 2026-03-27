package plw

import (
	"encoding/json"
	"errors"
	"sync"
)

var errLocatorConfigNotFound = errors.New(`locator config not found`)

// LocatorConfig 描述新版 locator 执行配置。
// LocatorConfig describes the new locator execution config.
type LocatorConfig struct {
	Version  int                   `json:"version"`
	Mode     string                `json:"mode"`
	Strategy string                `json:"strategy"`
	Locators []LocatorConfigItem   `json:"locators"`
	Options  *LocatorConfigOptions `json:"options,omitempty"`
}

// LocatorConfigItem 描述一个基础 locator 规则。
// LocatorConfigItem describes one base locator rule.
type LocatorConfigItem struct {
	ID      string        `json:"id"`
	Role    string        `json:"role,omitempty"`
	Query   *LocatorInput `json:"query,omitempty"`
	OnFound any           `json:"on_found,omitempty"`
	Summary string        `json:"summary,omitempty"`
}

// LocatorConfigOptions 描述附加动作与提取配置。
// LocatorConfigOptions stores extra action and extraction config.
type LocatorConfigOptions struct {
	ActionType  string `json:"action_type,omitempty"`
	ExtractType string `json:"extract_type,omitempty"`
}

// LocatorConfigRunResult 描述新版 locator config 的运行结果。
// LocatorConfigRunResult describes the run result of the locator config.
type LocatorConfigRunResult struct {
	BoolValue *bool
	TextValue string
}

// LocatorConfigRunner 使用既有 RunOne 能力运行新版 locator config。
// LocatorConfigRunner runs the new locator config via existing RunOne-like capability.
type LocatorConfigRunner struct {
	runQuery func(item LocatorConfigItem, action *ElementAction) (*ElementResult, error)
}

func decodeLocatorConfig(raw string) (*LocatorConfig, error) {
	config := &LocatorConfig{}
	if err := json.Unmarshal([]byte(raw), config); err != nil {
		return nil, err
	}
	if config.Version != 2 || config.Mode == `` {
		return nil, errors.New(`invalid locator config`)
	}
	return config, nil
}

// Run 根据 mode 与 strategy 执行新版 locator config。
// Run executes the new locator config by mode and strategy.
func (h *LocatorConfigRunner) Run(config *LocatorConfig, value string) (*LocatorConfigRunResult, error) {
	if config == nil {
		return nil, errors.New(`locator config is nil`)
	}
	switch config.Mode {
	case `bool_result`:
		return h.runBoolResult(config)
	case `text_content`:
		return h.runTextContent(config)
	case `click`, `input`:
		return h.runAction(config, value)
	default:
		return nil, errors.New(`unsupported locator config mode`)
	}
}

func (h *LocatorConfigRunner) runBoolResult(config *LocatorConfig) (*LocatorConfigRunResult, error) {
	type boolResult struct {
		value bool
		ok    bool
	}
	resultCh := make(chan boolResult, len(config.Locators))
	doneCh := make(chan struct{})
	var wg sync.WaitGroup

	for _, item := range config.Locators {
		current := item
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := h.runOne(current, &ElementAction{Type: `exist`})
			if err != nil || result == nil {
				return
			}
			boolValue, ok := readBoolOnFound(current.OnFound)
			if !ok {
				return
			}
			resultCh <- boolResult{value: boolValue, ok: true}
		}()
	}

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for {
		select {
		case result := <-resultCh:
			if result.ok {
				boolValue := result.value
				return &LocatorConfigRunResult{BoolValue: &boolValue}, nil
			}
		case <-doneCh:
			return nil, errLocatorConfigNotFound
		}
	}
}

func (h *LocatorConfigRunner) runTextContent(config *LocatorConfig) (*LocatorConfigRunResult, error) {
	if len(config.Locators) == 0 {
		return &LocatorConfigRunResult{TextValue: ``}, nil
	}

	type queryResult struct {
		action string
		result *ElementResult
		err    error
	}
	resultCh := make(chan queryResult, len(config.Locators))
	doneCh := make(chan struct{})
	var wg sync.WaitGroup
	for _, item := range config.Locators {
		current := item
		actionType, ok := readTextOnFound(current.OnFound)
		if !ok {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			action := &ElementAction{Type: `exist`}
			if actionType == `extract_text` {
				action = &ElementAction{Type: `text_content`}
			}
			result, err := h.runOne(current, action)
			resultCh <- queryResult{action: actionType, result: result, err: err}
		}()
	}

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for {
		select {
		case item := <-resultCh:
			// extract_text 一旦命中，就直接返回提取结果。
			// Once extract_text matches, return extracted text immediately.
			if item.action == `extract_text` && item.err == nil && item.result != nil {
				return &LocatorConfigRunResult{TextValue: item.result.TextContent}, nil
			}
			// return_empty 一旦命中，就直接返回空字符串。
			// Once return_empty matches, return an empty string immediately.
			if item.action == `return_empty` && item.err == nil && item.result != nil {
				return &LocatorConfigRunResult{TextValue: ``}, nil
			}
		case <-doneCh:
			return &LocatorConfigRunResult{TextValue: ``}, nil
		}
	}
}

func (h *LocatorConfigRunner) runAction(config *LocatorConfig, value string) (*LocatorConfigRunResult, error) {
	actionType := config.Mode
	if config.Options != nil && config.Options.ActionType != `` {
		actionType = config.Options.ActionType
	}
	type actionResult struct {
		ok bool
	}
	resultCh := make(chan actionResult, len(config.Locators))
	doneCh := make(chan struct{})
	var wg sync.WaitGroup
	for _, item := range config.Locators {
		current := item
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.runOne(current, &ElementAction{Type: actionType, Value: value})
			if err != nil {
				return
			}
			resultCh <- actionResult{ok: true}
		}()
	}

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for {
		select {
		case result := <-resultCh:
			if result.ok {
				return &LocatorConfigRunResult{}, nil
			}
		case <-doneCh:
			return nil, errLocatorConfigNotFound
		}
	}
}

func (h *LocatorConfigRunner) runOne(item LocatorConfigItem, action *ElementAction) (*ElementResult, error) {
	if h != nil && h.runQuery != nil {
		return h.runQuery(item, action)
	}
	return nil, errLocatorConfigNotFound
}

// readBoolOnFound 用于读取 bool_result 的布尔返回配置。
// readBoolOnFound reads the bool_result on_found value.
func readBoolOnFound(value any) (bool, bool) {
	boolValue, ok := value.(bool)
	return boolValue, ok
}

// readTextOnFound 用于读取 text_content 的命中动作配置。
// readTextOnFound reads the text_content on_found action.
func readTextOnFound(value any) (string, bool) {
	action, ok := value.(string)
	if !ok || (action != `extract_text` && action != `return_empty`) {
		return ``, false
	}
	return action, true
}
