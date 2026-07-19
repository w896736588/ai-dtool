// Package business 提供按 ws_token 上报步骤与提交的内部入口（任务 6）。
package business

import (
	"dev_tool/internal/app/dtool/component/e2e/store"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// E2ERecordStepAddByToken recorder.js 收到一个步骤后，通过 iframe 同源 fetch 调用。
// ws_token 鉴权已由 middleware 验证，本函数只负责幂等校验与持久化。
func E2ERecordStepAddByToken(token string, step *define.RecordedStep) (*define.E2ERecordStepAddResponse, error) {
	if token == "" {
		return nil, errors.New("ws_token 不能为空")
	}
	rs := store.NewRecordSessionStore()
	row, err := rs.FindByToken(token)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("会话不存在或 token 已失效")
	}
	status := cast.ToString(row["status"])
	if status == "committed" || status == "discarded" {
		return nil, errors.New("会话已关闭")
	}
	if step == nil {
		return nil, errors.New("step 不能为空")
	}
	if step.ID == "" {
		step.ID = "stp_" + cast.ToString(time.Now().UnixNano())
	}
	if step.Version == "" {
		step.Version = "1.0"
	}
	if step.WaitAfterMs <= 0 {
		step.WaitAfterMs = 200
	}
	step.RecordedAt = time.Now().UnixMilli()
	payload, err := json.Marshal(step)
	if err != nil {
		return nil, fmt.Errorf("序列化步骤失败: %w", err)
	}
	sessionUUID := cast.ToString(row["session_id"])
	if err := rs.AppendStep(sessionUUID, string(payload)); err != nil {
		return nil, err
	}
	if err := rs.MarkRecording(cast.ToInt64(row["row_id"])); err != nil {
		return nil, err
	}
	return &define.E2ERecordStepAddResponse{
		StepID:    step.ID,
		SessionID: sessionUUID,
	}, nil
}

// E2ERecordCommitByToken recorder.js 收尾时通过 ws_token 提交：把所有步骤落库为用例。
// 若 req.GroupID > 0 则创建新用例；否则只更新 status=committed。
func E2ERecordCommitByToken(token string, req *define.E2ERecordCommitByTokenRequest) (*define.E2ERecordCommitResponse, error) {
	if token == "" {
		return nil, errors.New("ws_token 不能为空")
	}
	rs := store.NewRecordSessionStore()
	row, err := rs.FindByToken(token)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("会话不存在或 token 已失效")
	}

	rowID := cast.ToInt64(row["row_id"])
	sessionUUID := cast.ToString(row["session_id"])
	envURL := cast.ToString(row["env_url"])
	envBaseURL := cast.ToString(row["env_base_url"])
	steps := parseRecordedSteps(row["steps"])

	e2eSteps := make([]define.E2EStep, 0, len(steps))
	var allAsserts []define.E2EAssertion
	for _, s := range steps {
		e2eSteps = append(e2eSteps, define.E2EStep{
			ID:          s.ID,
			Type:        s.Type,
			Version:     s.Version,
			Description: s.Description,
			WaitAfterMs: s.WaitAfterMs,
			Config:      s.Config,
		})
		if len(s.Assertions) > 0 {
			var arr []define.E2EAssertion
			if json.Unmarshal(s.Assertions, &arr) == nil {
				allAsserts = append(allAsserts, arr...)
			}
		}
	}

	name := strings.TrimSpace(cast.ToString(row["name"]))
	groupID := 0
	tags := ""
	if req != nil {
		if strings.TrimSpace(req.Name) != "" {
			name = strings.TrimSpace(req.Name)
		}
		groupID = req.GroupID
		tags = strings.TrimSpace(req.Tags)
	}

	var caseID int64
	if groupID > 0 {
		stepsJSON, _ := json.Marshal(e2eSteps)
		assertsJSON, _ := json.Marshal(allAsserts)
		cs := store.NewCaseStore()
		createReq := &define.E2ECaseSaveRequest{
			Name:           name,
			GroupID:        groupID,
			EnvURL:         envURL,
			EnvBaseURL:     envBaseURL,
			Steps:          stepsJSON,
			Assertions:     assertsJSON,
			Tags:           tags,
			TimeoutSeconds: 600,
		}
		caseID, err = cs.Create(createReq)
		if err != nil {
			return nil, err
		}
	}
	_ = rowID // 保留 row_id 字段语义，便于后续扩展
	if err := rs.UpdateStatus(sessionUUID, "committed"); err != nil {
		return nil, err
	}
	return &define.E2ERecordCommitResponse{
		CaseID:  caseID,
		Steps:   len(e2eSteps),
		GroupID: groupID,
	}, nil
}