package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/define"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// E2EGroupList 列出分组。
func E2EGroupList(c *gin.Context) {
	var req define.E2EGroupListRequest
	_ = gsgin.GinPostBody(c, &req)
	resp, err := business.E2EGroupList(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// E2EGroupCreate 创建分组。
func E2EGroupCreate(c *gin.Context) {
	var req define.E2EGroupCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		gsgin.GinResponseError(c, "名称不能为空", nil)
		return
	}
	id, err := business.E2EGroupCreate(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", gin.H{"id": id})
}

// E2EGroupUpdate 更新分组。
func E2EGroupUpdate(c *gin.Context) {
	var req define.E2EGroupUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.ID <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}
	if err := business.E2EGroupUpdate(&req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// E2EGroupDelete 删除分组。
func E2EGroupDelete(c *gin.Context) {
	var req define.E2EGroupDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.ID <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}
	if err := business.E2EGroupDelete(req.ID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// ---- 用例 ----

// E2ECaseList 列出用例。
func E2ECaseList(c *gin.Context) {
	var req define.E2ECaseListRequest
	_ = gsgin.GinPostBody(c, &req)
	resp, err := business.E2ECaseList(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// E2ECaseDetail 获取用例详情。
func E2ECaseDetail(c *gin.Context) {
	var req define.E2ECaseDetailRequest
	_ = gsgin.GinPostBody(c, &req)
	if req.ID <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}
	data, err := business.E2ECaseDetail(req.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if data == nil {
		gsgin.GinResponseError(c, "用例不存在", nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", data)
}

// E2ECaseSave 创建或更新用例。
func E2ECaseSave(c *gin.Context) {
	var req define.E2ECaseSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		gsgin.GinResponseError(c, "名称不能为空", nil)
		return
	}
	if req.ID <= 0 && req.GroupID <= 0 {
		gsgin.GinResponseError(c, "group_id 不能为空", nil)
		return
	}
	id, err := business.E2ECaseSave(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", gin.H{"id": id})
}

// E2ECaseDelete 删除用例。
func E2ECaseDelete(c *gin.Context) {
	var req define.E2ECaseDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.ID <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}
	if err := business.E2ECaseDelete(req.ID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// E2ECaseCreate 创建用例。
func E2ECaseCreate(c *gin.Context) {
	var req define.E2ECaseSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		gsgin.GinResponseError(c, "名称不能为空", nil)
		return
	}
	if req.GroupID <= 0 {
		gsgin.GinResponseError(c, "group_id 不能为空", nil)
		return
	}
	id, err := business.E2ECaseSave(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", gin.H{"id": id})
}

// E2ECaseUpdate 更新用例。
func E2ECaseUpdate(c *gin.Context) {
	var req define.E2ECaseSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.ID <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}
	id, err := business.E2ECaseSave(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", gin.H{"id": id})
}

// ---- 执行 ----

// E2ERunExecute 触发执行（异步）。
func E2ERunExecute(c *gin.Context) {
	var req define.E2ERunExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.CaseID <= 0 {
		gsgin.GinResponseError(c, "case_id 不能为空", nil)
		return
	}
	runID, err := business.E2ERunExecute(req.CaseID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", define.E2ERunExecuteResponse{RunID: runID})
}

// E2ERunBatch 批量执行（按 group）。
func E2ERunBatch(c *gin.Context) {
	var req define.E2ERunBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.GroupID <= 0 {
		gsgin.GinResponseError(c, "group_id 不能为空", nil)
		return
	}
	ids, err := business.E2ERunExecuteBatch(req.GroupID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", define.E2ERunBatchResponse{RunIDs: ids})
}

// E2ERunStop 停止执行。
func E2ERunStop(c *gin.Context) {
	var req define.E2ERunStopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.RunID <= 0 {
		gsgin.GinResponseError(c, "run_id 不能为空", nil)
		return
	}
	if err := business.E2ERunStop(req.RunID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// E2ERunList 列出执行。
func E2ERunList(c *gin.Context) {
	var req define.E2ERunListRequest
	_ = gsgin.GinPostBody(c, &req)
	resp, err := business.E2ERunList(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// E2ERunDetail 执行详情。
func E2ERunDetail(c *gin.Context) {
	var req define.E2ERunDetailRequest
	_ = gsgin.GinPostBody(c, &req)
	if req.RunID <= 0 {
		gsgin.GinResponseError(c, "run_id 不能为空", nil)
		return
	}
	data, err := business.E2ERunDetail(req.RunID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", data)
}

// E2ERunRequests 请求追踪。
func E2ERunRequests(c *gin.Context) {
	// 支持 GET 请求的 RESTful 风格路由
	runID := c.Param("runId")
	if runID != "" {
		// RESTful 风格：GET /api/e2e/run/:runId/requests
		req := define.E2ERunRequestsRequest{
			RunID: cast.ToInt64(runID),
		}
		rows, err := business.E2ERunRequests(&req)
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", gin.H{"requests": rows, "total": len(rows)})
		return
	}

	// POST 请求风格
	var req define.E2ERunRequestsRequest
	_ = gsgin.GinPostBody(c, &req)
	if req.RunID <= 0 {
		gsgin.GinResponseError(c, "run_id 不能为空", nil)
		return
	}
	rows, err := business.E2ERunRequests(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", gin.H{"list": rows, "total": len(rows)})
}

// E2ERunRequestDetail 单个请求详情。
func E2ERunRequestDetail(c *gin.Context) {
	runID := c.Param("runId")
	requestID := c.Param("requestId")
	if runID == "" || requestID == "" {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	detail, err := business.E2ERunRequestDetail(cast.ToInt64(runID), requestID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", detail)
}

// ---- 类型清单 ----

// E2EStepTypeList 步骤类型清单。
func E2EStepTypeList(c *gin.Context) {
	gsgin.GinResponseSuccess(c, "", business.E2EStepTypeList())
}

// E2EAssertionTypeList 断言类型清单。
func E2EAssertionTypeList(c *gin.Context) {
	gsgin.GinResponseSuccess(c, "", business.E2EAssertionTypeList())
}

// E2EHealth 健康检查。
func E2EHealth(c *gin.Context) {
	gsgin.GinResponseSuccess(c, "", gin.H{
		"status":  "ok",
		"steps":   len(business.E2EStepTypeList().Items),
		"asserts": len(business.E2EAssertionTypeList().Items),
	})
}

// =============== 录制功能（v5.0 新增） ===============

// E2ERecordSessionCreate 创建录制会话。
func E2ERecordSessionCreate(c *gin.Context) {
	var req define.E2ERecordSessionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if strings.TrimSpace(req.SessionName) == "" {
		gsgin.GinResponseError(c, "session_name 不能为空", nil)
		return
	}
	resp, err := business.E2ERecordSessionCreate(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// E2ERecordSessionGet 获取单个录制会话详情。
func E2ERecordSessionGet(c *gin.Context) {
	var req define.E2ERecordSessionGetRequest
	_ = gsgin.GinPostBody(c, &req)
	if req.ID <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}
	data, err := business.E2ERecordSessionGet(req.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if data == nil {
		gsgin.GinResponseError(c, "录制会话不存在", nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", data)
}

// E2ERecordSessionList 列出录制会话。
func E2ERecordSessionList(c *gin.Context) {
	var req define.E2ERecordSessionListRequest
	_ = gsgin.GinPostBody(c, &req)
	resp, err := business.E2ERecordSessionList(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// E2ERecordSessionDelete 删除录制会话。
func E2ERecordSessionDelete(c *gin.Context) {
	var req define.E2ERecordSessionDeleteRequest
	_ = gsgin.GinPostBody(c, &req)
	if req.ID <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}
	if err := business.E2ERecordSessionDelete(req.ID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// E2ERecordStepAdd 录制过程中追加一步。
func E2ERecordStepAdd(c *gin.Context) {
	var req define.E2ERecordStepAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.SessionID <= 0 {
		gsgin.GinResponseError(c, "session_id 不能为空", nil)
		return
	}
	if strings.TrimSpace(string(req.Step.Type)) == "" {
		gsgin.GinResponseError(c, "step.type 不能为空", nil)
		return
	}
	resp, err := business.E2ERecordStepAdd(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// E2ERecordStepUpdate 更新录制会话中的某一步。
func E2ERecordStepUpdate(c *gin.Context) {
	var req define.E2ERecordStepUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.SessionID <= 0 || req.StepID == "" {
		gsgin.GinResponseError(c, "session_id / step_id 不能为空", nil)
		return
	}
	if err := business.E2ERecordStepUpdate(&req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// E2ERecordStepDelete 删除录制会话中的某一步。
func E2ERecordStepDelete(c *gin.Context) {
	var req define.E2ERecordStepDeleteRequest
	_ = gsgin.GinPostBody(c, &req)
	if req.SessionID <= 0 || req.StepID == "" {
		gsgin.GinResponseError(c, "session_id / step_id 不能为空", nil)
		return
	}
	if err := business.E2ERecordStepDelete(&req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

// E2ERecordCommit 将录制会话落库为用例。
func E2ERecordCommit(c *gin.Context) {
	var req define.E2ERecordCommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.SessionID <= 0 {
		gsgin.GinResponseError(c, "session_id 不能为空", nil)
		return
	}
	if req.GroupID <= 0 {
		gsgin.GinResponseError(c, "group_id 不能为空", nil)
		return
	}
	resp, err := business.E2ERecordCommit(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// =============== 录制功能 v6（基于 smart_link + ws_token） ===============

// E2ERecordOpen 开启一次 smart_link 录制会话：开浏览器 + 写 DB + 注入 init script。
func E2ERecordOpen(c *gin.Context) {
	var req define.E2ERecordOpenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	resp, err := business.E2ERecordOpen(&req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if !resp.OK {
		gsgin.GinResponseError(c, resp.Error, resp)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// E2ERecordResume 按 session 行续录：清掉旧 ws_token 后重开。
func E2ERecordResume(c *gin.Context) {
	var req define.E2ERecordResumeRequest
	_ = gsgin.GinPostBody(c, &req)
	if req.SessionID <= 0 {
		gsgin.GinResponseError(c, "session_id 必须为正数", nil)
		return
	}
	resp, err := business.E2ERecordResume(req.SessionID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// E2ERecordStepAddByToken recorder.js 通过 ws_token 上报单步。
// ws_token 鉴权已由 RecorderTokenAuthMiddleware 完成。
func E2ERecordStepAddByToken(c *gin.Context) {
	tokenRaw, _ := c.Get("ws_token")
	token, _ := tokenRaw.(string)
	var req define.E2ERecordStepByTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	resp, err := business.E2ERecordStepAddByToken(token, &req.Step)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// E2ERecordCommitByToken recorder.js 通过 ws_token 提交录制到用例。
func E2ERecordCommitByToken(c *gin.Context) {
	tokenRaw, _ := c.Get("ws_token")
	token, _ := tokenRaw.(string)
	var req define.E2ERecordCommitByTokenRequest
	_ = c.ShouldBindJSON(&req)
	resp, err := business.E2ERecordCommitByToken(token, &req)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", resp)
}

// 避免部分导入未引用告警
var _ = cast.ToInt64
