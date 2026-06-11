package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// AgentCliPromptTemplateList 返回提示词模板列表以及共享工作目录候选。
func AgentCliPromptTemplateList(c *gin.Context) {
	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_cli_prompt_template ORDER BY sort_order ASC, id ASC`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	relRows, err := common.DbMain.Client.QueryBySql(
		`SELECT template_id, local_dir FROM tbl_agent_cli_prompt_template_dir_rel ORDER BY id ASC`,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dirMap := make(map[int][]string)
	for _, row := range relRows {
		templateID := cast.ToInt(row["template_id"])
		localDir := strings.TrimSpace(cast.ToString(row["local_dir"]))
		if templateID <= 0 || localDir == "" {
			continue
		}
		dirMap[templateID] = append(dirMap[templateID], localDir)
	}

	items := make([]define.AgentCliPromptTemplateItem, 0, len(rows))
	for _, row := range rows {
		id := cast.ToInt(row["id"])
		items = append(items, define.AgentCliPromptTemplateItem{
			Id:           id,
			Name:         cast.ToString(row["name"]),
			Content:      cast.ToString(row["content"]),
			ApplyAllDirs: cast.ToInt(row["apply_all_dirs"]) == 1,
			SortOrder:    cast.ToInt(row["sort_order"]),
			LocalDirs:    dirMap[id],
			CreatedAt:    cast.ToInt64(row["created_at"]),
			UpdatedAt:    cast.ToInt64(row["updated_at"]),
		})
	}

	historyRows, err := common.DbMain.AgentChatListAll()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", gin.H{
		"list":       items,
		"local_dirs": extractAgentCliPromptTemplateDirs(historyRows),
	})
}

// AgentCliPromptTemplateSave 新增或编辑提示词模板。
func AgentCliPromptTemplateSave(c *gin.Context) {
	var req define.AgentCliPromptTemplateSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Content = strings.TrimSpace(req.Content)
	if req.Name == "" {
		gsgin.GinResponseError(c, "模板名称不能为空", nil)
		return
	}
	if req.Content == "" {
		gsgin.GinResponseError(c, "模板内容不能为空", nil)
		return
	}

	normalizedDirs := normalizeAgentCliPromptTemplateDirs(req.LocalDirs)
	now := time.Now().Unix()
	templateID := req.Id
	if req.Id > 0 {
		_, err := common.DbMain.Client.ExecBySql(
			`UPDATE tbl_agent_cli_prompt_template SET name = ?, content = ?, apply_all_dirs = ?, sort_order = ?, updated_at = ? WHERE id = ?`,
			req.Name, req.Content, boolToInt(req.ApplyAllDirs), req.SortOrder, now, req.Id,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	} else {
		lastID, err := common.DbMain.Client.InsertBySql(
			`INSERT INTO tbl_agent_cli_prompt_template (name, content, apply_all_dirs, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
			req.Name, req.Content, boolToInt(req.ApplyAllDirs), req.SortOrder, now, now,
		).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		templateID = int(lastID)
	}

	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_cli_prompt_template_dir_rel WHERE template_id = ?`,
		templateID,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if !req.ApplyAllDirs {
		for _, localDir := range normalizedDirs {
			_, err = common.DbMain.Client.InsertBySql(
				`INSERT INTO tbl_agent_cli_prompt_template_dir_rel (template_id, local_dir) VALUES (?, ?)`,
				templateID, localDir,
			).Exec()
			if err != nil {
				gsgin.GinResponseError(c, err.Error(), nil)
				return
			}
		}
	}

	gsgin.GinResponseSuccess(c, "", define.AgentCliPromptTemplateItem{
		Id:           templateID,
		Name:         req.Name,
		Content:      req.Content,
		ApplyAllDirs: req.ApplyAllDirs,
		SortOrder:    req.SortOrder,
		LocalDirs:    normalizedDirs,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}

// AgentCliPromptTemplateDelete 删除提示词模板及其目录关联。
func AgentCliPromptTemplateDelete(c *gin.Context) {
	var req define.AgentCliPromptTemplateDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gsgin.GinResponseError(c, "参数错误", nil)
		return
	}
	if req.Id <= 0 {
		gsgin.GinResponseError(c, "id 不能为空", nil)
		return
	}

	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_cli_prompt_template WHERE id = ?`,
		req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	_, err = common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_agent_cli_prompt_template_dir_rel WHERE template_id = ?`,
		req.Id,
	).Exec()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, "", nil)
}

func extractAgentCliPromptTemplateDirs(rows []map[string]any) []string {
	dirSet := make(map[string]struct{})
	dirList := make([]string, 0)
	for _, row := range rows {
		localDir := strings.TrimSpace(cast.ToString(row["local_dir"]))
		if localDir == "" {
			continue
		}
		if _, exists := dirSet[localDir]; exists {
			continue
		}
		dirSet[localDir] = struct{}{}
		dirList = append(dirList, localDir)
	}
	return dirList
}

func normalizeAgentCliPromptTemplateDirs(values []string) []string {
	seen := make(map[string]struct{})
	list := make([]string, 0, len(values))
	for _, value := range values {
		dir := strings.TrimSpace(value)
		if dir == "" {
			continue
		}
		if _, exists := seen[dir]; exists {
			continue
		}
		seen[dir] = struct{}{}
		list = append(list, dir)
	}
	return list
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
