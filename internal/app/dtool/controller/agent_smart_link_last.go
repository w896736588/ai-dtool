package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// SmartLinkLastForAgent 给本地 agent 提供智能网页数据目录历史的代理接口。
// 该接口把 DB 访问留在服务端，避免 agent 进程直接读取用户配置库。
func SmartLinkLastForAgent(c *gin.Context) {
	var req define.AgentSmartLinkLastRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	switch req.Action {
	case define.AgentSmartLinkLastActionGetLast:
		smartLinkLastForAgentGetLast(c, req)
	case define.AgentSmartLinkLastActionExists:
		smartLinkLastForAgentExists(c, req)
	case define.AgentSmartLinkLastActionUpsert:
		smartLinkLastForAgentUpsert(c, req)
	default:
		gsgin.GinResponseError(c, "action不支持", nil)
	}
}

// smartLinkLastForAgentGetLast 查询某个用户在指定域名上次使用的数据目录索引。
func smartLinkLastForAgentGetLast(c *gin.Context, req define.AgentSmartLinkLastRequest) {
	if strings.TrimSpace(req.UserName) == "" || strings.TrimSpace(req.Domain) == "" {
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkLastResponse{})
		return
	}
	row, err := common.DbLog.Client.QueryBySql(
		`select * from tbl_smart_link_last where user_name = ? and domain = ? `,
		req.UserName,
		req.Domain,
	).One()
	if err != nil {
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkLastResponse{})
		return
	}
	gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkLastResponse{
		UserDataIndex: cast.ToInt(row[`user_data_index`]),
	})
}

// smartLinkLastForAgentExists 判断指定域名和目录索引是否已有历史占用记录。
func smartLinkLastForAgentExists(c *gin.Context, req define.AgentSmartLinkLastRequest) {
	if strings.TrimSpace(req.Domain) == "" || req.UserDataIndex <= 0 {
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkLastResponse{})
		return
	}
	row, err := common.DbLog.Client.QueryBySql(
		`select * from tbl_smart_link_last where domain = ? and user_data_index = ? `,
		req.Domain,
		req.UserDataIndex,
	).One()
	if err != nil {
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkLastResponse{})
		return
	}
	gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkLastResponse{
		Exists: len(row) > 0,
	})
}

// smartLinkLastForAgentUpsert 写入或更新本次任务实际使用的数据目录索引。
func smartLinkLastForAgentUpsert(c *gin.Context, req define.AgentSmartLinkLastRequest) {
	if req.SmartLinkID <= 0 || strings.TrimSpace(req.UserName) == "" || strings.TrimSpace(req.Domain) == "" || req.UserDataIndex <= 0 {
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkLastResponse{})
		return
	}
	now := time.Now().Unix()
	row, err := common.DbLog.Client.QueryBySql(
		`select * from tbl_smart_link_last where  smart_link_id = ? and user_name = ? and domain = ?`,
		req.SmartLinkID,
		req.UserName,
		req.Domain,
	).One()
	if err == nil && len(row) > 0 {
		_, err = common.DbLog.Client.QuickUpdate(`tbl_smart_link_last`, map[string]any{
			`smart_link_id`: req.SmartLinkID,
			`user_name`:     req.UserName,
			`domain`:        req.Domain,
		}, map[string]any{
			`user_data_index`: req.UserDataIndex,
			`update_time`:     now,
		}).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkLastResponse{})
		return
	}
	_, err = common.DbLog.Client.QuickCreate(`tbl_smart_link_last`, map[string]any{
		`smart_link_id`:   req.SmartLinkID,
		`user_name`:       req.UserName,
		`user_data_index`: req.UserDataIndex,
		`domain`:          req.Domain,
		`create_time`:     now,
		`update_time`:     now,
	}).Exec()
	if err != nil && strings.Contains(err.Error(), `UNIQUE constraint failed: tbl_smart_link_last.domain, tbl_smart_link_last.user_data_index`) {
		// 同域名同目录唯一索引冲突时，说明已有占用记录，回退为按占用键更新。
		_, err = common.DbLog.Client.QuickUpdate(`tbl_smart_link_last`, map[string]any{
			`domain`:          req.Domain,
			`user_data_index`: req.UserDataIndex,
		}, map[string]any{
			`smart_link_id`: req.SmartLinkID,
			`user_name`:     req.UserName,
			`update_time`:   now,
		}).Exec()
	}
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", define.AgentSmartLinkLastResponse{})
}
