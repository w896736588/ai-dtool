package controller

import (
	"strings"
	"time"

	"dev_tool/internal/app/dtool/component"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// MemoryFragmentShareCreate 创建一个 24 小时有效的知识片段只读分享 token。
func MemoryFragmentShareCreate(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	fragmentID := strings.TrimSpace(cast.ToString(dataMap[`id`]))
	if fragmentID == `` || fragmentID == `0` {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	if _, err := memoryDB.MemoryFragmentInfo(fragmentID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, err := shareStore.Create(fragmentID, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, `创建分享链接失败：`+err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, ``, memoryFragmentShareResponse(share))
}

// MemoryFragmentShareInfo 通过分享 token 读取知识片段详情，只返回查看页所需数据。
func MemoryFragmentShareInfo(c *gin.Context) {
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		gsgin.GinResponseError(c, err.Error(), map[string]any{
			`configured`: false,
		})
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	token := strings.TrimSpace(cast.ToString(dataMap[`token`]))
	if token == `` {
		gsgin.GinResponseError(c, `分享链接不能为空`, nil)
		return
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, ok, err := shareStore.Resolve(token, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, `读取分享链接失败：`+err.Error(), nil)
		return
	}
	if !ok {
		gsgin.GinResponseError(c, `分享链接不存在或已过期`, nil)
		return
	}
	info, err := component.MemoryRuntime.DB().MemoryFragmentInfo(share.FragmentID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`fragment`: info,
		`share`:    memoryFragmentShareResponse(share),
	})
}

func memoryFragmentShareResponse(share memoryFragmentShare) map[string]any {
	return map[string]any{
		`token`:          share.Token,
		`fragment_id`:    share.FragmentID,
		`expire_at`:      share.ExpireAt.Unix(),
		`expire_at_desc`: gstool.TimeUnixToString(share.ExpireAt, `Y-m-d H:i:s`),
	}
}
