package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// AgentCliGroupList 返回 AgentCli 专用分组列表（含每个分组关联的 AgentCli 数量）
func AgentCliGroupList(c *gin.Context) {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_cli_group ORDER BY sort_order ASC, id ASC`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 统计每个分组关联的 AgentCli 数量
	countRows, _ := common.DbMain.Client.QueryBySql(
		`SELECT group_id, COUNT(*) as cnt FROM tbl_agent_cli_group_rel GROUP BY group_id`,
	).All()
	countMap := make(map[int]int)
	for _, cr := range countRows {
		countMap[cast.ToInt(cr["group_id"])] = cast.ToInt(cr["cnt"])
	}

	type groupItemWithCount struct {
		define.AgentCliGroupItem
		CliCount int `json:"cli_count"`
	}

	items := make([]groupItemWithCount, 0, len(rows))
	for _, row := range rows {
		gid := cast.ToInt(row["id"])
		items = append(items, groupItemWithCount{
			AgentCliGroupItem: define.AgentCliGroupItem{
				Id:        gid,
				Name:      cast.ToString(row["name"]),
				SortOrder: cast.ToInt(row["sort_order"]),
				CreatedAt: cast.ToInt64(row["created_at"]),
				UpdatedAt: cast.ToInt64(row["updated_at"]),
			},
			CliCount: countMap[gid],
		})
	}

	gsgin.GinResponseSuccess(c, "", gin.H{"list": items})
}

// AgentCliGroupSave 新增/编辑 AgentCli 专用分组
func AgentCliGroupSave(c *gin.Context) {
	var req define.AgentCliGroupSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if req.Name == "" {
		gsgin.GinResponseError(c, "分组名称不能为空", nil)
		return
	}

	now := time.Now().Unix()

	if req.Id > 0 {
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_cli_group SET name = ?, sort_order = ?, updated_at = ? WHERE id = ?`,
			req.Name, req.SortOrder, now, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, "", define.AgentCliGroupItem{
			Id:        req.Id,
			Name:      req.Name,
			SortOrder: req.SortOrder,
			UpdatedAt: now,
		})
		return
	}

	lastId, err := common.DbMain.Client.InsertBySql(
		`INSERT INTO tbl_agent_cli_group (name, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		req.Name, req.SortOrder, now, now,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", define.AgentCliGroupItem{
		Id:        int(lastId),
		Name:      req.Name,
		SortOrder: req.SortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// AgentCliGroupDelete 删除 AgentCli 专用分组（同时清理关联表）
func AgentCliGroupDelete(c *gin.Context) {
	var req define.AgentCliGroupDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if req.Id <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}

	// 删除分组记录
	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_cli_group WHERE id = ?`, req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 清理关联表中该分组的所有关联
	common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_cli_group_rel WHERE group_id = ?`, req.Id,
	).Exec()

	gsgin.GinResponseSuccess(c, "", nil)
}

// AgentCliGroupRelSave 保存某个 AgentCli 的分组关联（先删旧关联再批量插入）
func AgentCliGroupRelSave(c *gin.Context) {
	var req define.AgentCliGroupRelSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}

	if req.AgentCliId <= 0 {
		gsgin.GinResponseError(c, "agent_cli_id 不能为空", nil)
		return
	}

	// 删除该 AgentCli 的旧关联
	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_cli_group_rel WHERE agent_cli_id = ?`, req.AgentCliId,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 批量插入新关联
	for _, groupId := range req.GroupIds {
		if groupId <= 0 {
			continue
		}
		common.DbMain.Client.InsertBySql(
			`INSERT INTO tbl_agent_cli_group_rel (agent_cli_id, group_id) VALUES (?, ?)`,
			req.AgentCliId, groupId,
		).Exec()
	}

	gsgin.GinResponseSuccess(c, "", nil)
}
