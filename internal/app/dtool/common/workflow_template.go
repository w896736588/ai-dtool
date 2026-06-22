package common

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

const (
	// 工作流固定步骤的 step_key 常量
	WorkflowFixedStepTaskConfig       = `task-config`
	WorkflowFixedStepRequirementFetch = `requirement-fetch`
	WorkflowFixedStepIssueFix         = `issue_fix`

	// 自定义步骤 step_key 前缀，格式 custom_{id}
	workflowCustomStepKeyPrefix = `custom_`

	// 表名常量
	workflowTemplateTable     = `tbl_workflow_template`
	workflowTemplateStepTable = `tbl_workflow_template_step`
)

// WorkflowTemplateStepDocument 描述模板步骤中预生成的知识片段文档配置。
type WorkflowTemplateStepDocument struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Placeholder string `json:"placeholder"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	IsApiDoc    bool   `json:"is_api_doc"`
}

// WorkflowTemplateStepDocumentsParse 解析步骤文档配置 JSON。
func WorkflowTemplateStepDocumentsParse(raw string) []WorkflowTemplateStepDocument {
	raw = strings.TrimSpace(raw)
	if raw == `` || raw == `null` || raw == `[]` {
		return nil
	}
	var docs []WorkflowTemplateStepDocument
	if err := json.Unmarshal([]byte(raw), &docs); err != nil {
		return nil
	}
	result := make([]WorkflowTemplateStepDocument, 0, len(docs))
	for _, doc := range docs {
		doc.ID = strings.TrimSpace(doc.ID)
		doc.Name = strings.TrimSpace(doc.Name)
		doc.Placeholder = strings.TrimSpace(doc.Placeholder)
		doc.Title = strings.TrimSpace(doc.Title)
		if doc.Name == `` {
			continue
		}
		if doc.ID == `` {
			doc.ID = workflowTemplateStepDocumentGenerateID()
		}
		result = append(result, doc)
	}
	return result
}

// workflowTemplateStepDocumentGenerateID 生成一个简短的文档唯一标识。
func workflowTemplateStepDocumentGenerateID() string {
	input := fmt.Sprintf(`%d-%d`, time.Now().UnixNano(), rand.Intn(900000)+100000)
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])[:8]
}

// WorkflowTemplateStepDocumentsToRelativePlaceholder 根据文档占位符生成对应的文件相对路径占位符。
// 例如：{接口文档地址} -> {接口文档地址文件相对地址}
func WorkflowTemplateStepDocumentsToRelativePlaceholder(placeholder string) string {
	placeholder = strings.TrimSpace(placeholder)
	if len(placeholder) < 2 {
		return ``
	}
	if !strings.HasPrefix(placeholder, `{`) || !strings.HasSuffix(placeholder, `}`) {
		return ``
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(placeholder, `{`), `}`)
	return `{` + inner + `文件相对地址}`
}

// WorkflowTemplateStepDocumentsToIDPlaceholder 根据文档占位符生成对应的文档ID占位符。
// 例如：{纯文本需求文档} -> {纯文本需求文档ID}
func WorkflowTemplateStepDocumentsToIDPlaceholder(placeholder string) string {
	placeholder = strings.TrimSpace(placeholder)
	if len(placeholder) < 2 {
		return ``
	}
	if !strings.HasPrefix(placeholder, `{`) || !strings.HasSuffix(placeholder, `}`) {
		return ``
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(placeholder, `{`), `}`)
	return `{` + inner + `ID}`
}

// WorkflowTemplateFixedStepKeys 返回所有固定步骤的 step_key 列表。
func WorkflowTemplateFixedStepKeys() []string {
	return []string{
		WorkflowFixedStepTaskConfig,
		WorkflowFixedStepRequirementFetch,
		WorkflowFixedStepIssueFix,
	}
}

// WorkflowStepKeyIsFixed 判断 step_key 是否为固定步骤。
func WorkflowStepKeyIsFixed(stepKey string) bool {
	for _, fixed := range WorkflowTemplateFixedStepKeys() {
		if stepKey == fixed {
			return true
		}
	}
	return false
}

// WorkflowCustomStepKey 生成自定义步骤的 step_key（格式：custom_{id}）。
func WorkflowCustomStepKey(stepID int) string {
	return workflowCustomStepKeyPrefix + cast.ToString(stepID)
}

// ===================== 模板 CRUD =====================

// WorkflowTemplateList 获取所有模板列表（含步骤）。
func (h *CSqlite) WorkflowTemplateList() ([]map[string]any, error) {
	templates, err := h.Client.QuickQuery(workflowTemplateTable, `*`, nil).Order(`sort_order ASC, id ASC`).All()
	if err != nil {
		return nil, err
	}
	for i, template := range templates {
		templateID := cast.ToInt(template[`id`])
		steps, stepErr := h.WorkflowTemplateStepsByTemplateID(templateID)
		if stepErr != nil {
			return nil, stepErr
		}
		templates[i][`steps`] = steps
	}
	return templates, nil
}

// WorkflowTemplateInfo 获取单个模板详情（含步骤）。
func (h *CSqlite) WorkflowTemplateInfo(templateID int) (map[string]any, error) {
	if templateID <= 0 {
		return nil, errors.New(`模板id不能为空`)
	}
	template, err := h.Client.QuickQuery(workflowTemplateTable, `*`, map[string]any{
		`id`: templateID,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(template) == 0 {
		return nil, errors.New(`模板不存在`)
	}
	steps, stepErr := h.WorkflowTemplateStepsByTemplateID(templateID)
	if stepErr != nil {
		return nil, stepErr
	}
	template[`steps`] = steps
	return template, nil
}

// WorkflowTemplateDefaultInfo 获取默认模板详情（含步骤）。
func (h *CSqlite) WorkflowTemplateDefaultInfo() (map[string]any, error) {
	template, err := h.Client.QuickQuery(workflowTemplateTable, `*`, map[string]any{
		`is_default`: 1,
	}).Order(`id ASC`).One()
	if err != nil {
		return nil, err
	}
	if len(template) == 0 {
		return nil, errors.New(`默认模板不存在`)
	}
	templateID := cast.ToInt(template[`id`])
	steps, stepErr := h.WorkflowTemplateStepsByTemplateID(templateID)
	if stepErr != nil {
		return nil, stepErr
	}
	template[`steps`] = steps
	return template, nil
}

// WorkflowTemplateSave 创建或更新模板（upsert 模式）。
func (h *CSqlite) WorkflowTemplateSave(id int, name, description string) (int64, error) {
	name = trimSpace(name)
	if name == `` {
		return 0, errors.New(`模板名称不能为空`)
	}
	now := time.Now().Unix()
	if id > 0 {
		// 更新现有模板
		_, err := h.Client.QuickUpdate(workflowTemplateTable, map[string]any{
			`id`: id,
		}, map[string]any{
			`name`:        name,
			`description`: description,
			`update_time`: now,
		}).Exec()
		return int64(id), err
	}
	// 创建新模板
	newID, err := h.Client.QuickCreate(workflowTemplateTable, map[string]any{
		`name`:        name,
		`description`: description,
		`is_default`:  0,
		`sort_order`:  0,
		`create_time`: now,
		`update_time`: now,
	}).Exec()
	if err != nil {
		return 0, err
	}
	// 创建新模板后，自动添加固定步骤
	templateID := cast.ToInt(newID)
	if err := h.workflowTemplateCreateFixedSteps(templateID); err != nil {
		return 0, err
	}
	return newID, nil
}

// WorkflowTemplateImport 导入模板（创建模板 + 批量创建步骤，跳过自动添加固定步骤）。
func (h *CSqlite) WorkflowTemplateImport(name, description string, steps []WorkflowTemplateImportStepData) (int64, error) {
	name = trimSpace(name)
	if name == `` {
		return 0, errors.New(`模板名称不能为空`)
	}
	now := time.Now().Unix()
	// 创建模板（不添加固定步骤）
	newID, err := h.Client.QuickCreate(workflowTemplateTable, map[string]any{
		`name`:        name,
		`description`: description,
		`is_default`:  0,
		`sort_order`:  0,
		`create_time`: now,
		`update_time`: now,
	}).Exec()
	if err != nil {
		return 0, err
	}
	templateID := cast.ToInt(newID)
	// 批量创建步骤
	for i, step := range steps {
		stepName := trimSpace(step.Name)
		if stepName == `` {
			continue
		}
		isFixed := step.IsFixed
		stepKey := step.StepKey
		if isFixed == 0 && stepKey == `` {
			// 自定义步骤在创建后再设置 step_key
		}
		stepNewID, createErr := h.Client.QuickCreate(workflowTemplateStepTable, map[string]any{
			`template_id`:    templateID,
			`name`:           stepName,
			`step_key`:       stepKey,
			`prompt_content`: step.PromptContent,
			`step_documents`: step.StepDocuments,
			`remark`:         step.Remark,
			`sort_order`:     step.SortOrder,
			`is_fixed`:       isFixed,
			`create_time`:    now,
			`update_time`:    now,
		}).Exec()
		if createErr != nil {
			return 0, createErr
		}
		// 自定义步骤需要更新 step_key 为 custom_{id}
		if isFixed == 0 && stepKey == `` {
			actualStepKey := workflowCustomStepKeyPrefix + cast.ToString(stepNewID)
			_, _ = h.Client.QuickUpdate(workflowTemplateStepTable, map[string]any{
				`id`: stepNewID,
			}, map[string]any{
				`step_key`: actualStepKey,
			}).Exec()
		}
		_ = i
	}
	return newID, nil
}

// WorkflowTemplateImportStepData 导入模板步骤数据。
type WorkflowTemplateImportStepData struct {
	Name          string
	StepKey       string
	PromptContent string
	StepDocuments string
	Remark        string
	IsFixed       int
	SortOrder     int
}

// WorkflowTemplateDelete 删除模板（检查是否有关联任务）。
func (h *CSqlite) WorkflowTemplateDelete(templateID int) error {
	if templateID <= 0 {
		return errors.New(`模板id不能为空`)
	}
	// 检查是否为默认模板
	template, err := h.Client.QuickQuery(workflowTemplateTable, `*`, map[string]any{
		`id`: templateID,
	}).One()
	if err != nil {
		return err
	}
	if cast.ToInt(template[`is_default`]) == 1 {
		return errors.New(`默认模板不能删除`)
	}
	// 检查是否有关联任务
	rows, err := h.Client.QuickQuery(`tbl_home_task`, `id`, map[string]any{
		`workflow_template_id`: templateID,
	}).All()
	if err != nil {
		return err
	}
	if len(rows) > 0 {
		return errors.New(`该模板有关联的任务，无法删除`)
	}
	// 删除模板步骤
	_, _ = h.Client.QuickDelete(workflowTemplateStepTable, map[string]any{
		`template_id`: templateID,
	}).Exec()
	// 删除模板
	_, err = h.Client.QuickDelete(workflowTemplateTable, map[string]any{
		`id`: templateID,
	}).Exec()
	return err
}

// WorkflowTemplateSetDefault 设置默认模板（先取消其他默认，再设置当前）。
func (h *CSqlite) WorkflowTemplateSetDefault(templateID int) error {
	if templateID <= 0 {
		return errors.New(`模板id不能为空`)
	}
	// 取消其他默认
	_, err := h.dbExec(`UPDATE "` + workflowTemplateTable + `" SET "is_default" = 0 WHERE "is_default" = 1`)
	if err != nil {
		return err
	}
	// 设置当前为默认
	_, err = h.Client.QuickUpdate(workflowTemplateTable, map[string]any{
		`id`: templateID,
	}, map[string]any{
		`is_default`:  1,
		`update_time`: time.Now().Unix(),
	}).Exec()
	return err
}

// ===================== 步骤 CRUD =====================

// WorkflowTemplateStepsByTemplateID 根据模板ID获取所有步骤（按 sort_order 排序）。
func (h *CSqlite) WorkflowTemplateStepsByTemplateID(templateID int) ([]map[string]any, error) {
	if templateID <= 0 {
		return nil, errors.New(`模板id不能为空`)
	}
	steps, err := h.Client.QuickQuery(workflowTemplateStepTable, `*`, map[string]any{
		`template_id`: templateID,
	}).Order(`sort_order ASC, id ASC`).All()
	if err != nil {
		return nil, err
	}
	return steps, nil
}

// WorkflowTemplateStepInfo 获取单个步骤详情。
func (h *CSqlite) WorkflowTemplateStepInfo(stepID int) (map[string]any, error) {
	if stepID <= 0 {
		return nil, errors.New(`步骤id不能为空`)
	}
	return h.Client.QuickQuery(workflowTemplateStepTable, `*`, map[string]any{
		`id`: stepID,
	}).One()
}

// workflowTemplateStepKeyExists 检查同一模板中是否已存在相同的 step_key（排除指定步骤自身）。
func (h *CSqlite) workflowTemplateStepKeyExists(templateID int, stepKey string, excludeStepID int) (bool, error) {
	if stepKey == `` {
		return false, nil
	}
	steps, err := h.WorkflowTemplateStepsByTemplateID(templateID)
	if err != nil {
		return false, err
	}
	for _, step := range steps {
		sID := cast.ToInt(step[`id`])
		if sID == excludeStepID {
			continue
		}
		if cast.ToString(step[`step_key`]) == stepKey {
			return true, nil
		}
	}
	return false, nil
}

// WorkflowTemplateStepSave 创建或更新步骤。
func (h *CSqlite) WorkflowTemplateStepSave(id, templateID int, name, stepKey, promptContent, stepDocuments, remark string, sortOrder int) (int64, error) {
	name = trimSpace(name)
	if name == `` {
		return 0, errors.New(`步骤名称不能为空`)
	}
	if templateID <= 0 {
		return 0, errors.New(`模板id不能为空`)
	}
	now := time.Now().Unix()
	if id > 0 {
		// 更新现有步骤
		updateData := map[string]any{
			`name`:           name,
			`prompt_content`: promptContent,
			`step_documents`: stepDocuments,
			`remark`:         remark,
			`update_time`:    now,
		}
		// 非固定步骤可以修改 step_key（用于自定义步骤）
		existingStep, err := h.WorkflowTemplateStepInfo(id)
		if err != nil {
			return 0, err
		}
		if len(existingStep) > 0 {
			isFixed := cast.ToInt(existingStep[`is_fixed`])
			existingStepKey := cast.ToString(existingStep[`step_key`])
			if isFixed == 1 && existingStepKey != stepKey {
				return 0, errors.New(`固定步骤不能修改 step_key`)
			}
			if stepKey != `` {
				// 检查 step_key 唯一性（排除自身）
				exists, dupErr := h.workflowTemplateStepKeyExists(templateID, stepKey, id)
				if dupErr != nil {
					return 0, dupErr
				}
				if exists {
					return 0, errors.New(`步骤 key 已存在，不能重复`)
				}
				updateData[`step_key`] = stepKey
			}
		}
		_, err = h.Client.QuickUpdate(workflowTemplateStepTable, map[string]any{
			`id`: id,
		}, updateData).Exec()
		return int64(id), err
	}
	// 创建新步骤
	// 自定义步骤使用 custom_{新id} 作为 step_key
	// 先插入获取 id，再更新 step_key
	newID, err := h.Client.QuickCreate(workflowTemplateStepTable, map[string]any{
		`template_id`:    templateID,
		`name`:           name,
		`step_key`:       ``,
		`prompt_content`: promptContent,
		`step_documents`: stepDocuments,
		`remark`:         remark,
		`sort_order`:     sortOrder,
		`is_fixed`:       0,
		`create_time`:    now,
		`update_time`:    now,
	}).Exec()
	if err != nil {
		return 0, err
	}
	stepID := cast.ToInt(newID)
	// 使用生成的 ID 设置 step_key
	actualStepKey := workflowCustomStepKeyPrefix + cast.ToString(stepID)
	if stepKey != `` {
		// 检查用户指定 step_key 的唯一性（排除自身，id 还未赋值用 0）
		exists, dupErr := h.workflowTemplateStepKeyExists(templateID, stepKey, 0)
		if dupErr != nil {
			return 0, dupErr
		}
		if exists {
			return 0, errors.New(`步骤 key 已存在，不能重复`)
		}
		actualStepKey = stepKey
	}
	_, err = h.Client.QuickUpdate(workflowTemplateStepTable, map[string]any{
		`id`: stepID,
	}, map[string]any{
		`step_key`: actualStepKey,
	}).Exec()
	return int64(stepID), err
}

// WorkflowTemplateStepDelete 删除步骤（固定步骤不可删除）。
func (h *CSqlite) WorkflowTemplateStepDelete(stepID int) error {
	if stepID <= 0 {
		return errors.New(`步骤id不能为空`)
	}
	step, err := h.WorkflowTemplateStepInfo(stepID)
	if err != nil {
		return err
	}
	if len(step) == 0 {
		return errors.New(`步骤不存在`)
	}
	if cast.ToInt(step[`is_fixed`]) == 1 {
		return errors.New(`固定步骤不能删除`)
	}
	_, err = h.Client.QuickDelete(workflowTemplateStepTable, map[string]any{
		`id`: stepID,
	}).Exec()
	return err
}

// WorkflowTemplateStepSort 更新步骤排序（传入排序后的步骤ID列表）。
func (h *CSqlite) WorkflowTemplateStepSort(templateID int, stepIDs []int) error {
	if templateID <= 0 {
		return errors.New(`模板id不能为空`)
	}
	if len(stepIDs) == 0 {
		return nil
	}
	now := time.Now().Unix()
	for i, stepID := range stepIDs {
		_, err := h.Client.QuickUpdate(workflowTemplateStepTable, map[string]any{
			`id`: stepID,
		}, map[string]any{
			`sort_order`:  i,
			`update_time`: now,
		}).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// ===================== 工作流实例提示词相关 =====================

// WorkflowStepPromptsRead 读取工作流实例的 step_prompts JSON，返回 map。
func (h *CSqlite) WorkflowStepPromptsRead(workflowID int) (map[string]string, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	info, err := h.TaskWorkflowInfo(workflowID)
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`工作流不存在`)
	}

	result := make(map[string]string)
	stepPromptsRaw := cast.ToString(info[`step_prompts`])
	if stepPromptsRaw != `` {
		if err := json.Unmarshal([]byte(stepPromptsRaw), &result); err != nil {
			gstool.FmtPrintlnLogTime(`[workflow] step_prompts JSON 解析失败 workflowID=%d err=%v`, workflowID, err)
		}
	}
	return result, nil
}

// WorkflowStepPromptsSave 保存工作流实例的单个步骤提示词。
// 同时写入新字段 step_prompts JSON 和旧字段 prompt_xxx（向后兼容）。
func (h *CSqlite) WorkflowStepPromptsSave(workflowID int, stepKey, stepPrompt string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	stepKey = trimSpace(stepKey)
	if stepKey == `` {
		return errors.New(`步骤key不能为空`)
	}

	// 读取现有 step_prompts
	existing, err := h.WorkflowStepPromptsRead(workflowID)
	if err != nil {
		return err
	}
	existing[stepKey] = stepPrompt

	// 序列化为 JSON
	jsonBytes, err := json.Marshal(existing)
	if err != nil {
		return err
	}

	now := time.Now().Unix()

	// 写入新字段 step_prompts
	_, err = h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`step_prompts`: string(jsonBytes),
		`update_time`:  now,
	}).Exec()
	if err != nil {
		return err
	}

	return nil
}

// WorkflowStepPromptsRestore 从模板步骤还原提示词。
// 对每个模板步骤，读取其 prompt_content 作为默认值写入 step_prompts。
func (h *CSqlite) WorkflowStepPromptsRestore(workflowID int, templateSteps []map[string]any) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}

	prompts := make(map[string]string)
	for _, step := range templateSteps {
		stepKey := cast.ToString(step[`step_key`])
		promptContent := cast.ToString(step[`prompt_content`])
		prompts[stepKey] = promptContent
	}

	jsonBytes, err := json.Marshal(prompts)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	_, err = h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`step_prompts`: string(jsonBytes),
		`update_time`:  now,
	}).Exec()

	return nil
}

// ===================== 辅助方法 =====================

// workflowTemplateCreateFixedSteps 为新建模板自动添加固定步骤。
func (h *CSqlite) workflowTemplateCreateFixedSteps(templateID int) error {
	now := time.Now().Unix()
	fixedSteps := []struct {
		name    string
		stepKey string
		order   int
	}{
		{`任务配置`, WorkflowFixedStepTaskConfig, 0},
		{`抓取需求`, WorkflowFixedStepRequirementFetch, 1},
		{`问题修改`, WorkflowFixedStepIssueFix, 99}, // 問題修改放在最后
	}
	for _, fs := range fixedSteps {
		_, err := h.Client.QuickCreate(workflowTemplateStepTable, map[string]any{
			`template_id`:    templateID,
			`name`:           fs.name,
			`step_key`:       fs.stepKey,
			`prompt_content`: ``,
			`sort_order`:     fs.order,
			`is_fixed`:       1,
			`create_time`:    now,
			`update_time`:    now,
		}).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// dbExec 执行原生 SQL（忽略影响行数，仅关心是否有错误）。
func (h *CSqlite) dbExec(sql string) (int64, error) {
	return h.Client.ExecBySql(sql).Exec()
}

// WorkflowMigrateLegacyStepKeys 将已有工作流的 step_prompts 和 node_statuses 中的旧 step_key 迁移到 custom_xx 格式。
// 应在 SQL 迁移（将模板步骤 step_key 更新为 custom_xx）之后调用。
// 该操作是幂等的：已使用新 key 的记录不会被重复修改。
func (h *CSqlite) WorkflowMigrateLegacyStepKeys() {
	// 旧 step_name → old step_key 的固定映射（来自旧版默认模板定义）
	nameToOldKey := map[string]string{
		`需求分析`:      `requirement`,
		`开发执行`:      `design`,
		`接口生成`:      `api-dev`,
		`自动化测试+修复`:  `api-test-fix`,
		`代码检查`:      `code-review`,
		`需求核对浏览器测试`: `browser-test`,
	}

	// 读取默认模板步骤，构建 name → new step_key 映射
	steps, err := h.WorkflowTemplateStepsByTemplateID(1)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[workflow] 迁移旧step_key失败: 读取模板步骤 err=%v`, err)
		return
	}

	oldToNew := map[string]string{}
	for _, step := range steps {
		name := cast.ToString(step[`name`])
		stepKey := cast.ToString(step[`step_key`])
		if oldKey, ok := nameToOldKey[name]; ok {
			if stepKey != `` && stepKey != oldKey {
				oldToNew[oldKey] = stepKey
			}
		}
	}

	if len(oldToNew) == 0 {
		return // 无需迁移
	}

	// 遍历所有工作流
	workflows, err := h.Client.QuickQuery(`tbl_task_workflow`, `*`, nil).All()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[workflow] 迁移旧step_key失败: 查询工作流 err=%v`, err)
		return
	}

	now := time.Now().Unix()
	migratedCount := 0
	for _, wf := range workflows {
		workflowID := cast.ToInt(wf[`id`])
		needUpdate := false
		updateData := map[string]any{`update_time`: now}

		// 迁移 step_prompts：重命名 key + 替换值中的旧 step_key 引用
		stepPromptsRaw := cast.ToString(wf[`step_prompts`])
		if stepPromptsRaw != `` {
			prompts := map[string]string{}
			if json.Unmarshal([]byte(stepPromptsRaw), &prompts) == nil {
				keyChanged := migrateMapKeys(prompts, oldToNew)
				valChanged := migratePromptValues(prompts, oldToNew)
				if keyChanged || valChanged {
					newJSON, _ := json.Marshal(prompts)
					updateData[`step_prompts`] = string(newJSON)
					needUpdate = true
				}
			}
		}

		// 迁移 node_statuses
		nodeStatusesRaw := cast.ToString(wf[`node_statuses`])
		if nodeStatusesRaw != `` {
			statuses := map[string]string{}
			if json.Unmarshal([]byte(nodeStatusesRaw), &statuses) == nil {
				if migrateMapKeys(statuses, oldToNew) {
					newJSON, _ := json.Marshal(statuses)
					updateData[`node_statuses`] = string(newJSON)
					needUpdate = true
				}
			}
		}

		if needUpdate {
			_, _ = h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{`id`: workflowID}, updateData).Exec()
			migratedCount++
		}
	}

	if migratedCount > 0 {
		gstool.FmtPrintlnLogTime(`[workflow] 迁移旧step_key完成: 更新了 %d 个工作流`, migratedCount)
	}
}

// migrateMapKeys 将 map 中的旧 key 替换为新 key。
// 若新 key 已有值则保留（不覆盖），返回是否发生了任何变更。
func migrateMapKeys(m map[string]string, oldToNew map[string]string) bool {
	changed := false
	for oldKey, newKey := range oldToNew {
		if val, ok := m[oldKey]; ok {
			delete(m, oldKey)
			// 若新 key 已有值则保留（不覆盖），优先保留用户可能已用新 key 保存的值
			if _, exists := m[newKey]; !exists {
				m[newKey] = val
			}
			changed = true
		}
	}
	return changed
}

// migratePromptValues 将 prompt 值中的旧 step_key 文本替换为新 step_key。
// 例如将 "当前步骤：requirement" 替换为 "当前步骤：custom_3"。
// 返回是否发生了任何变更。
func migratePromptValues(prompts map[string]string, oldToNew map[string]string) bool {
	// 构建反向映射：newKey → oldKey
	newToOld := map[string]string{}
	for oldK, newK := range oldToNew {
		newToOld[newK] = oldK
	}
	changed := false
	for newKey, oldKey := range newToOld {
		if val, ok := prompts[newKey]; ok && strings.Contains(val, oldKey) {
			prompts[newKey] = strings.ReplaceAll(val, oldKey, newKey)
			changed = true
		}
	}
	return changed
}

// trimSpace 封装 strings.TrimSpace，方便本文件内使用。
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
