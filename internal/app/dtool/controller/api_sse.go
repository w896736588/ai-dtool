package controller

import (
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_define"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
)

const apiDataChangeSseStatusPrefix = `ClientId:`

// BroadcastApiChange 将 API 数据变更广播到所有已连接的 SSE 客户端。
// sourceClientId 是触发变更的前端 SSE 客户端 ID，前端据此跳过自身。
// changeType 参见 plan 中的 change_type 枚举。
// ids 携带受影响的 collection_id / folder_id / api_id / old_folder_id 等字段。
func BroadcastApiChange(sourceClientId, changeType string, ids map[string]any) {
	if ids == nil {
		ids = make(map[string]any)
	}
	ids[`source_client_id`] = sourceClientId
	ids[`change_type`] = changeType

	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseApiDataChange,
		Data:            ids,
		Type:            p_define.SseContentTypeMsg,
	})

	for _, item := range gsgin.SseStatus() {
		clientID := strings.TrimSpace(strings.TrimPrefix(item, apiDataChangeSseStatusPrefix))
		if clientID == `` || clientID == item {
			continue
		}
		sse := gsgin.SseGetByClientId(clientID)
		if sse == nil {
			continue
		}
		_ = sse.SendToChan(msg)
	}
}
