package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_define"
	"time"

	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

func buildWorkflowUnreadSnapshotData() (map[string]any, error) {
	items, err := common.DbMain.TaskWorkflowActiveUnreadSnapshots()
	if err != nil {
		return nil, err
	}
	taskUnreadMap := make(map[int]int, len(items))
	workflowItems := make([]map[string]any, 0, len(items))
	workflowUnreadTotal := 0
	for _, item := range items {
		if item.HomeTaskID > 0 {
			taskUnreadMap[item.HomeTaskID] = item.WorkflowUnread
		}
		workflowUnreadTotal += item.WorkflowUnread
		promptTypeUnread := make(map[string]int, len(item.PromptTypeUnread))
		for promptType, unread := range item.PromptTypeUnread {
			promptTypeUnread[promptType] = unread
		}
		workflowItems = append(workflowItems, map[string]any{
			`home_task_id`:       item.HomeTaskID,
			`workflow_id`:        item.WorkflowID,
			`workflow_unread`:    item.WorkflowUnread,
			`has_unread`:         item.WorkflowUnread > 0,
			`top_history_unread`: item.WorkflowUnread > 0,
			`prompt_type_unread`: promptTypeUnread,
			`type`:               `workflow_detail_badge`,
		})
	}
	return map[string]any{
		`type`:                   `workflow_unread_snapshot`,
		`workflow_menu_badge`:    map[string]any{`type`: `workflow_menu_badge`, `has_unread`: workflowUnreadTotal > 0, `unread_total`: workflowUnreadTotal},
		`workflow_task_badges`:   taskUnreadMap,
		`workflow_detail_badges`: workflowItems,
	}, nil
}

func sendWorkflowUnreadSnapshot(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	data, err := buildWorkflowUnreadSnapshotData()
	if err != nil {
		gstool.FmtPrintlnLogTime(`WorkflowUnreadSnapshot 查询错误 %s`, err.Error())
		return
	}
	distributeIDs := []string{
		define.SseWorkflowUnreadHomeMenu,
		define.SseWorkflowUnreadHomeTask,
		define.SseWorkflowUnreadDetail,
	}
	for _, distributeID := range distributeIDs {
		err = sse.SendToChan(gstool.JsonEncode(p_define.SseData{
			SseDistributeId: distributeID,
			Data:            data,
			Type:            p_define.SseContentTypeMsg,
		}))
		if err != nil {
			gstool.FmtPrintlnLogTime(`WorkflowUnreadSnapshot 推送错误 %s distribute_id=%s`, err.Error(), distributeID)
		}
	}
}

// BindWorkflowUnreadSnapshotSSE attaches workflow unread badge snapshot events to a normal SSE stream.
func BindWorkflowUnreadSnapshotSSE(sse *gsgin.Sse, stopC chan int, interval time.Duration) {
	if sse == nil {
		return
	}
	if interval <= 0 {
		interval = 3 * time.Second
	}
	sendWorkflowUnreadSnapshot(sse)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sendWorkflowUnreadSnapshot(sse)
			case <-stopC:
				return
			}
		}
	}()
}
