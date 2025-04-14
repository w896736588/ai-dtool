package p_variable

import "dev_tool/base"

func (h *VariableRun) CmdList(variableId any) ([]map[string]any, error) {
	return base.Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`variable_id`: variableId,
		`status`:      1,
	}).Order(`weight asc`).All()
}

func (h *VariableRun) Variable(variableId any) map[string]any {
	variableInfo, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable`, `*`, map[string]interface{}{
		`id`: variableId,
	}).One()
	return variableInfo
}
