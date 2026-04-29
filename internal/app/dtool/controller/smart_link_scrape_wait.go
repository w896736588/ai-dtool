package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
)

const (
	// smartLinkScrapeTaskPollInterval 抓取任务结果轮询间隔，避免数据库空转过快。
	smartLinkScrapeTaskPollInterval = 300 * time.Millisecond
)

// querySmartLinkTaskByID 查询抓取任务当前状态，供同步等待结果复用。
func querySmartLinkTaskByID(taskID string) (map[string]any, error) {
	return common.DbMain.Client.QuickQuery("tbl_smart_link_task", "*", map[string]any{
		"task_id": taskID,
	}).One()
}

// parseSmartLinkTaskResultPayload 解析任务结果 JSON，提取下载地址等最终返回字段。
func parseSmartLinkTaskResultPayload(resultPayload string) (define.AgentTaskResultFileUploadResponse, error) {
	result := define.AgentTaskResultFileUploadResponse{}
	resultPayload = strings.TrimSpace(resultPayload)
	if resultPayload == "" || resultPayload == "{}" {
		return result, nil
	}
	payload := struct {
		DownloadURL string `json:"download_url"`
		FileName    string `json:"file_name"`
	}{}
	if err := json.Unmarshal([]byte(resultPayload), &payload); err != nil {
		return result, err
	}
	result.DownloadURL = strings.TrimSpace(payload.DownloadURL)
	result.FileName = strings.TrimSpace(payload.FileName)
	return result, nil
}

// isSmartLinkTaskSuccessStatus 统一兼容任务阶段状态与最终结果状态的成功标记。
func isSmartLinkTaskSuccessStatus(status string) bool {
	status = strings.TrimSpace(strings.ToLower(status))
	return status == string(define.SmartLinkTaskStatusSuccess) || status == "succeeded"
}

// waitForSmartLinkTaskResult 同步等待抓取任务结束，并返回 ZIP 下载地址。
func waitForSmartLinkTaskResult(taskID string, timeout time.Duration, queryFunc func(taskID string) (map[string]any, error)) (define.AgentTaskResultFileUploadResponse, error) {
	if strings.TrimSpace(taskID) == "" {
		return define.AgentTaskResultFileUploadResponse{}, errors.New("task_id不能为空")
	}
	if queryFunc == nil {
		return define.AgentTaskResultFileUploadResponse{}, errors.New("queryFunc不能为空")
	}
	deadline := time.Now().Add(timeout)
	for {
		info, err := queryFunc(taskID)
		if err != nil {
			return define.AgentTaskResultFileUploadResponse{}, err
		}
		status := strings.TrimSpace(cast.ToString(info["status"]))
		if isSmartLinkTaskSuccessStatus(status) {
			result, parseErr := parseSmartLinkTaskResultPayload(cast.ToString(info["result_payload"]))
			if parseErr != nil {
				return define.AgentTaskResultFileUploadResponse{}, parseErr
			}
			if result.DownloadURL == "" {
				return define.AgentTaskResultFileUploadResponse{}, errors.New("抓取任务已完成，但未生成zip下载地址")
			}
			return result, nil
		}
		if status == string(define.SmartLinkTaskStatusFailed) || status == string(define.SmartLinkTaskStatusCancelled) {
			errMsg := strings.TrimSpace(cast.ToString(info["error_message"]))
			if errMsg == "" {
				errMsg = "抓取任务执行失败"
			}
			return define.AgentTaskResultFileUploadResponse{}, errors.New(errMsg)
		}
		if time.Now().After(deadline) {
			return define.AgentTaskResultFileUploadResponse{}, fmt.Errorf("等待抓取任务结果超时 task_id=%s", taskID)
		}
		time.Sleep(smartLinkScrapeTaskPollInterval)
	}
}
