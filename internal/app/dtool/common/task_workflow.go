package common

import (
	"dev_tool/internal/app/dtool/memory"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

const (
	taskWorkflowStatusInit              = `init`
	taskWorkflowStageIdle               = `idle`
	taskWorkflowStageRequirementFetch   = `requirement_fetch`
	taskWorkflowRequirementFetchIdle    = `idle`
	taskWorkflowRequirementFetchRunning = `running`
	taskWorkflowRequirementFetchSuccess = `success`
	taskWorkflowRequirementFetchFailed  = `failed`
)

const (
	taskWorkflowChatStatusRunning     = `running`
	taskWorkflowChatStatusCompleted   = `completed`
	taskWorkflowChatStatusError       = `error`
	taskWorkflowChatStatusInterrupted = `interrupted`
)

const (
	agentChatReadYes = 1
	agentChatReadNo  = 0
)

const (
	TaskWorkflowChatStatusRunning     = taskWorkflowChatStatusRunning
	TaskWorkflowChatStatusCompleted   = taskWorkflowChatStatusCompleted
	TaskWorkflowChatStatusError       = taskWorkflowChatStatusError
	TaskWorkflowChatStatusInterrupted = taskWorkflowChatStatusInterrupted
)

const (
	// agentChatTableName 通用 Agent 对话表名。 // Shared table name for all agent chat records.
	agentChatTableName = `agent_chat`
	// AgentChatSourceTypeWorkflow 表示记录来自工作流。 // Source type used when the chat is created from a workflow.
	AgentChatSourceTypeWorkflow = `work_flow`
	// AgentChatSourceTypeAgentCli 表示记录来自 AgentCli 独立执行。 // Source type used when the chat is created from standalone AgentCli execution.
	AgentChatSourceTypeAgentCli = `agent_cli`
)

// TaskWorkflowFragmentRef 描述工作流中一个知识片段引用。
type TaskWorkflowFragmentRef struct {
	Raw        string
	FolderName string
	FileID     string
	FullRef    string
	IsLegacy   bool
}

// TaskWorkflowNormalizeFolderName 规范化工作流默认知识片段文件夹。
func TaskWorkflowNormalizeFolderName(folderName string) string {
	return memory.NormalizeFolderName(folderName)
}

// TaskWorkflowFragmentFolderName 读取工作流配置的默认知识片段文件夹。
func TaskWorkflowFragmentFolderName(workflowInfo map[string]any) string {
	if len(workflowInfo) == 0 {
		return memory.DefaultFolderName
	}
	return TaskWorkflowNormalizeFolderName(cast.ToString(workflowInfo[`fragment_folder_name`]))
}

// TaskWorkflowBuildFragmentRef 使用“文件夹/片段ID”格式构建工作流知识片段引用。
func TaskWorkflowBuildFragmentRef(folderName, fileID string) string {
	fileID = strings.TrimSpace(fileID)
	if fileID == `` {
		return ``
	}
	return TaskWorkflowNormalizeFolderName(folderName) + `/` + fileID
}

// TaskWorkflowParseFragmentRef 解析工作流知识片段引用，兼容旧的纯 file_id 数据。
func TaskWorkflowParseFragmentRef(raw, fallbackFolderName string) TaskWorkflowFragmentRef {
	raw = strings.TrimSpace(raw)
	fallbackFolderName = TaskWorkflowNormalizeFolderName(fallbackFolderName)
	ref := TaskWorkflowFragmentRef{
		Raw:        raw,
		FolderName: fallbackFolderName,
		FileID:     ``,
		FullRef:    ``,
		IsLegacy:   false,
	}
	if raw == `` {
		return ref
	}
	if idx := strings.LastIndex(raw, `/`); idx > 0 && idx < len(raw)-1 {
		ref.FolderName = TaskWorkflowNormalizeFolderName(raw[:idx])
		ref.FileID = strings.TrimSpace(raw[idx+1:])
		ref.FullRef = TaskWorkflowBuildFragmentRef(ref.FolderName, ref.FileID)
		return ref
	}
	ref.FileID = raw
	ref.FullRef = TaskWorkflowBuildFragmentRef(ref.FolderName, ref.FileID)
	ref.IsLegacy = true
	return ref
}

// TaskWorkflowNormalizeFragmentRef 统一输出工作流知识片段引用格式。
func TaskWorkflowNormalizeFragmentRef(raw, fallbackFolderName string) string {
	return TaskWorkflowParseFragmentRef(raw, fallbackFolderName).FullRef
}

// TaskWorkflowContainsFragmentID 检查指定知识片段是否属于指定工作流。
// 查询 tbl_task_workflow_document 表，检查是否存在 workflowID 与 fragmentID 匹配的记录。
func (h *CSqlite) TaskWorkflowContainsFragmentID(workflowID int, fragmentID string) (bool, error) {
	if workflowID <= 0 || strings.TrimSpace(fragmentID) == `` {
		return false, nil
	}
	fragmentID = strings.TrimSpace(fragmentID)

	docRecords, err := h.Client.QuickQuery(taskWorkflowDocumentTable, `id`, map[string]any{
		`workflow_id`: workflowID,
		`file_id`:     fragmentID,
	}).All()
	if err != nil {
		return false, err
	}
	return len(docRecords) > 0, nil
}

// TaskWorkflowCreateOrGetByHomeTaskID 查询或创建任务工作流主记录。
func (h *CSqlite) TaskWorkflowCreateOrGetByHomeTaskID(homeTaskID int) (map[string]any, error) {
	if homeTaskID <= 0 {
		return nil, errors.New(`任务id不能为空`)
	}
	existing, err := h.Client.QuickQuery(`tbl_task_workflow`, `*`, map[string]any{
		`home_task_id`: homeTaskID,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return existing, nil
	}
	homeTaskInfo, err := h.HomeTaskRow(homeTaskID)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	fetchType := strings.TrimSpace(strings.ToLower(cast.ToString(homeTaskInfo[`fetch_type`])))
	if fetchType == `` {
		fetchType = `tapd`
	}
	requirementSourceURL := strings.TrimSpace(cast.ToString(homeTaskInfo[`tapd_url`]))
	if fetchType == `zentao` {
		requirementSourceURL = strings.TrimSpace(cast.ToString(homeTaskInfo[`zentao_url`]))
	}
	newID, err := h.Client.QuickCreate(`tbl_task_workflow`, map[string]any{
		`home_task_id`:                  homeTaskID,
		`status`:                        taskWorkflowStatusInit,
		`current_stage`:                 taskWorkflowStageIdle,
		`requirement_fragment_id`:       strings.TrimSpace(cast.ToString(homeTaskInfo[`memory_fragment_id`])),
		`fragment_folder_name`:          memory.DefaultFolderName,
		`requirement_fetch_status`:      taskWorkflowRequirementFetchIdle,
		`requirement_fetch_started_at`:  0,
		`requirement_fetch_finished_at`: 0,
		`requirement_fetch_error`:       ``,
		`requirement_source_url`:        requirementSourceURL,
		`base_branch`:                   ``,
		`feature_branch`:                ``,
		`last_error`:                    ``,
		`create_time`:                   now,
		`update_time`:                   now,
	}).Exec()
	if err != nil {
		return nil, err
	}
	return h.TaskWorkflowInfo(cast.ToInt(newID))
}

// TaskWorkflowInfo 查询单条任务工作流主记录。
func (h *CSqlite) TaskWorkflowInfo(id int) (map[string]any, error) {
	if id <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	info, err := h.Client.QuickQuery(`tbl_task_workflow`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`工作流不存在`)
	}
	return info, nil
}

// TaskWorkflowBindDevPlanFragment 绑定开发执行片段引用（写入文档表）。
func (h *CSqlite) TaskWorkflowBindDevPlanFragment(workflowID int, fragmentRef string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	fragmentRef = strings.TrimSpace(fragmentRef)
	ref := TaskWorkflowParseFragmentRef(fragmentRef, memory.DefaultFolderName)
	if ref.FileID == `` {
		return errors.New(`开发执行片段id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	// 更新状态
	_, _ = h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{`id`: workflowID}, map[string]any{
		`status`:      `dev_plan_ready`,
		`update_time`: time.Now().Unix(),
	}).Exec()
	// 写入文档表
	return h.TaskWorkflowDocumentUpsert(workflowID, ``, TaskWorkflowDocumentTypeName(TaskWorkflowDocTypeDevPlan), TaskWorkflowDocTypeDevPlan, 0, 0, ref.FileID, ref.FolderName, ``)
}

// TaskWorkflowCreateRun 创建一条工作流测试/计划运行记录。
func (h *CSqlite) TaskWorkflowCreateRun(workflowID int, runType, triggerSource, requirementSnapshotMD, devPlanSnapshotMD string, coverageReport, testPlan, testReport map[string]any, summaryMD string) (map[string]any, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	runType = strings.TrimSpace(runType)
	if runType == `` {
		return nil, errors.New(`运行类型不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	runNo := `tw-` + cast.ToString(workflowID) + `-` + cast.ToString(now)
	newID, err := h.Client.QuickCreate(`tbl_task_test_run`, map[string]any{
		`workflow_id`:             workflowID,
		`run_no`:                  runNo,
		`run_type`:                runType,
		`status`:                  `success`,
		`trigger_source`:          strings.TrimSpace(triggerSource),
		`requirement_snapshot_md`: requirementSnapshotMD,
		`dev_plan_snapshot_md`:    devPlanSnapshotMD,
		`diff_snapshot_text`:      ``,
		`coverage_report_json`:    gstool.JsonEncode(coverageReport),
		`test_plan_json`:          gstool.JsonEncode(testPlan),
		`test_report_json`:        gstool.JsonEncode(testReport),
		`summary_md`:              summaryMD,
		`started_at`:              now,
		`finished_at`:             now,
		`create_time`:             now,
	}).Exec()
	if err != nil {
		return nil, err
	}
	updateData := map[string]any{
		`status`:        `test_plan_ready`,
		`current_stage`: runType,
		`update_time`:   now,
	}
	if runType == `api_test_execute` {
		updateData[`status`] = `test_run_ready`
	}
	_, err = h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, updateData).Exec()
	if err != nil {
		return nil, err
	}
	return h.TaskWorkflowRunInfo(cast.ToInt(newID))
}

// TaskWorkflowRunInfo 查询单条运行记录。
func (h *CSqlite) TaskWorkflowRunInfo(id int) (map[string]any, error) {
	if id <= 0 {
		return nil, errors.New(`运行记录id不能为空`)
	}
	info, err := h.Client.QuickQuery(`tbl_task_test_run`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`运行记录不存在`)
	}
	return info, nil
}

// TaskWorkflowLatestRunByType 查询某类最新运行记录。
func (h *CSqlite) TaskWorkflowLatestRunByType(workflowID int, runType string) (map[string]any, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	runType = strings.TrimSpace(runType)
	if runType == `` {
		return nil, errors.New(`运行类型不能为空`)
	}
	list, err := h.Client.QueryBySql(`select * from tbl_task_test_run where workflow_id = ? and run_type = ? order by id desc limit 1`, workflowID, runType).All()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return map[string]any{}, nil
	}
	return list[0], nil
}

// TaskWorkflowRunList 查询工作流运行历史。
func (h *CSqlite) TaskWorkflowRunList(workflowID int) ([]map[string]any, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	return h.Client.QueryBySql(`select * from tbl_task_test_run where workflow_id = ? order by id desc`, workflowID).All()
}

// TaskWorkflowMarkRequirementFetchRunning 标记需求抓取节点开始执行。
func (h *CSqlite) TaskWorkflowMarkRequirementFetchRunning(workflowID int, sourceURL string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`status`:                        taskWorkflowStageRequirementFetch,
		`current_stage`:                 taskWorkflowStageRequirementFetch,
		`requirement_fetch_status`:      taskWorkflowRequirementFetchRunning,
		`requirement_fetch_started_at`:  now,
		`requirement_fetch_finished_at`: 0,
		`requirement_fetch_error`:       ``,
		`requirement_source_url`:        strings.TrimSpace(sourceURL),
		`last_error`:                    ``,
		`update_time`:                   now,
	}).Exec()
	return err
}

// TaskWorkflowMarkRequirementFetchSuccess 标记需求抓取节点成功完成。
func (h *CSqlite) TaskWorkflowMarkRequirementFetchSuccess(workflowID int, sourceURL string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`status`:                        taskWorkflowStatusInit,
		`current_stage`:                 taskWorkflowStageRequirementFetch,
		`requirement_fetch_status`:      taskWorkflowRequirementFetchSuccess,
		`requirement_fetch_finished_at`: now,
		`requirement_fetch_error`:       ``,
		`requirement_source_url`:        strings.TrimSpace(sourceURL),
		`last_error`:                    ``,
		`update_time`:                   now,
	}).Exec()
	return err
}

// TaskWorkflowMarkRequirementFetchFailed 标记需求抓取节点执行失败。
func (h *CSqlite) TaskWorkflowMarkRequirementFetchFailed(workflowID int, sourceURL, errMsg string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	now := time.Now().Unix()
	errMsg = strings.TrimSpace(errMsg)
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`status`:                        taskWorkflowStatusInit,
		`current_stage`:                 taskWorkflowStageRequirementFetch,
		`requirement_fetch_status`:      taskWorkflowRequirementFetchFailed,
		`requirement_fetch_finished_at`: now,
		`requirement_fetch_error`:       errMsg,
		`requirement_source_url`:        strings.TrimSpace(sourceURL),
		`last_error`:                    errMsg,
		`update_time`:                   now,
	}).Exec()
	return err
}

// TaskWorkflowUpdateNodeStatuses 更新工作流节点状态。
func (h *CSqlite) TaskWorkflowUpdateNodeStatuses(workflowID int, nodeStatuses string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`node_statuses`: strings.TrimSpace(nodeStatuses),
		`update_time`:   time.Now().Unix(),
	}).Exec()
	return err
}

// TaskWorkflowBatchWorkflowSummaryByHomeTaskIDs 根据 home_task_id 列表批量查询工作流节点状态和未读数。
func (h *CSqlite) TaskWorkflowBatchWorkflowSummaryByHomeTaskIDs(homeTaskIDs []int) (map[int]string, map[int]int, error) {
	nodeStatusesMap := map[int]string{}
	unreadCountMap := map[int]int{}
	if len(homeTaskIDs) == 0 {
		return nodeStatusesMap, unreadCountMap, nil
	}
	placeholders := make([]string, 0, len(homeTaskIDs))
	args := make([]any, 0, len(homeTaskIDs)+3)
	args = append(args, AgentChatSourceTypeWorkflow, agentChatReadNo, taskWorkflowChatStatusRunning)
	for _, id := range homeTaskIDs {
		if id <= 0 {
			continue
		}
		placeholders = append(placeholders, `?`)
		args = append(args, id)
	}
	if len(placeholders) == 0 {
		return nodeStatusesMap, unreadCountMap, nil
	}
	list, err := h.Client.QueryBySql(`SELECT tw.home_task_id, tw.node_statuses, COALESCE(ac.unread_total, 0) AS unread_total
FROM tbl_task_workflow tw
LEFT JOIN (
	SELECT from_id AS workflow_id, COUNT(1) AS unread_total
	FROM agent_chat
	WHERE from_type = ? AND is_read = ? AND status <> ?
	GROUP BY from_id
) ac ON ac.workflow_id = tw.id
WHERE tw.home_task_id IN (`+strings.Join(placeholders, `,`)+`)`, args...).All()
	if err != nil {
		return nil, nil, err
	}
	for _, row := range list {
		homeTaskID := cast.ToInt(row[`home_task_id`])
		nodeStatusesMap[homeTaskID] = strings.TrimSpace(cast.ToString(row[`node_statuses`]))
		unreadCountMap[homeTaskID] = cast.ToInt(row[`unread_total`])
	}
	return nodeStatusesMap, unreadCountMap, nil
}

// TaskWorkflowBatchNodeStatusesByHomeTaskIDs 根据 home_task_id 列表批量查询工作流 node_statuses。
func (h *CSqlite) TaskWorkflowBatchNodeStatusesByHomeTaskIDs(homeTaskIDs []int) (map[int]string, error) {
	nodeStatusesMap, _, err := h.TaskWorkflowBatchWorkflowSummaryByHomeTaskIDs(homeTaskIDs)
	if err != nil {
		return nil, err
	}
	return nodeStatusesMap, nil
}

// WorkflowUnreadSnapshotItem describes one active workflow unread snapshot used by SSE badge updates.
type WorkflowUnreadSnapshotItem struct {
	HomeTaskID       int
	WorkflowID       int
	WorkflowUnread   int
	PromptTypeUnread map[string]int
}

// TaskWorkflowActiveUnreadSnapshots queries unread badge snapshot for every non-archived workflow task.
func (h *CSqlite) TaskWorkflowActiveUnreadSnapshots() ([]WorkflowUnreadSnapshotItem, error) {
	rows, err := h.Client.QueryBySql(`SELECT tw.id AS workflow_id, tw.home_task_id
FROM tbl_task_workflow tw
INNER JOIN tbl_home_task ht ON ht.id = tw.home_task_id
WHERE IFNULL(ht.is_archived, 0) = 0 AND IFNULL(ht.use_workflow, 0) <> 0
ORDER BY tw.id ASC`).All()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []WorkflowUnreadSnapshotItem{}, nil
	}

	workflowIDList := make([]int, 0, len(rows))
	workflowMap := make(map[int]*WorkflowUnreadSnapshotItem, len(rows))
	for _, row := range rows {
		workflowID := cast.ToInt(row[`workflow_id`])
		homeTaskID := cast.ToInt(row[`home_task_id`])
		if workflowID <= 0 || homeTaskID <= 0 {
			continue
		}
		workflowIDList = append(workflowIDList, workflowID)
		workflowMap[workflowID] = &WorkflowUnreadSnapshotItem{
			HomeTaskID:       homeTaskID,
			WorkflowID:       workflowID,
			WorkflowUnread:   0,
			PromptTypeUnread: map[string]int{},
		}
	}
	if len(workflowIDList) == 0 {
		return []WorkflowUnreadSnapshotItem{}, nil
	}

	placeholders := make([]string, 0, len(workflowIDList))
	args := make([]any, 0, len(workflowIDList)+3)
	args = append(args, AgentChatSourceTypeWorkflow, agentChatReadNo, taskWorkflowChatStatusRunning)
	for _, workflowID := range workflowIDList {
		placeholders = append(placeholders, `?`)
		args = append(args, workflowID)
	}
	query := fmt.Sprintf(`SELECT from_id AS workflow_id, prompt_type, COUNT(1) AS unread_total
FROM agent_chat
WHERE from_type = ? AND is_read = ? AND status <> ? AND from_id IN (%s)
GROUP BY from_id, prompt_type`, strings.Join(placeholders, `,`))
	unreadRows, err := h.Client.QueryBySql(query, args...).All()
	if err != nil {
		return nil, err
	}
	for _, row := range unreadRows {
		workflowID := cast.ToInt(row[`workflow_id`])
		item := workflowMap[workflowID]
		if item == nil {
			continue
		}
		unreadTotal := cast.ToInt(row[`unread_total`])
		promptType := strings.TrimSpace(cast.ToString(row[`prompt_type`]))
		item.WorkflowUnread += unreadTotal
		if promptType != `` {
			item.PromptTypeUnread[promptType] = unreadTotal
		}
	}

	list := make([]WorkflowUnreadSnapshotItem, 0, len(workflowIDList))
	for _, workflowID := range workflowIDList {
		if item := workflowMap[workflowID]; item != nil {
			list = append(list, *item)
		}
	}
	return list, nil
}

// TaskWorkflowUpdateFragmentFolderName 更新工作流默认知识片段文件夹。
func (h *CSqlite) TaskWorkflowUpdateFragmentFolderName(workflowID int, folderName string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`fragment_folder_name`: TaskWorkflowNormalizeFolderName(folderName),
		`update_time`:          time.Now().Unix(),
	}).Exec()
	return err
}

// TaskWorkflowRefreshFragmentRefsByFileID 根据片段 file_id 统一刷新工作流文档表中的文件夹名称。
func (h *CSqlite) TaskWorkflowRefreshFragmentRefsByFileID(fileID, newFolderName string) error {
	fileID = strings.TrimSpace(fileID)
	if fileID == `` {
		return nil
	}
	newFolderName = TaskWorkflowNormalizeFolderName(newFolderName)
	// 刷新文档表中的记录
	docRows, err := h.Client.QuickQuery(taskWorkflowDocumentTable, `*`, map[string]any{
		`file_id`: fileID,
	}).All()
	if err != nil {
		return err
	}
	for _, doc := range docRows {
		docID := cast.ToInt(doc[`id`])
		if docID <= 0 {
			continue
		}
		_, _ = h.Client.QuickUpdate(taskWorkflowDocumentTable, map[string]any{`id`: docID}, map[string]any{
			`folder_name`: newFolderName,
			`update_time`: time.Now().Unix(),
		}).Exec()
	}
	return nil
}

// AgentChatCreateBySource 创建通用对话记录。
// AgentChatCreateBySource creates a chat row for either workflow-driven or standalone AgentCli execution.
func (h *CSqlite) AgentChatCreateBySource(fromType string, fromID int, prompt, promptType, cliType string, agentCliID int, localDir, settingsPath, modelName string, thinkingCollapsed int, thinkingIntensity string) (int64, error) {
	// 仅允许受控来源类型，避免把查询语义写坏。 // Restrict source types so downstream filtering stays deterministic.
	if fromType != AgentChatSourceTypeWorkflow && fromType != AgentChatSourceTypeAgentCli {
		return 0, errors.New(`对话来源类型无效`)
	}
	if fromID <= 0 {
		return 0, errors.New(`对话来源id不能为空`)
	}
	if strings.TrimSpace(cliType) == `` {
		cliType = `claude`
	}
	now := time.Now().Format(`2006-01-02 15:04:05`)
	id, err := h.Client.QuickCreate(agentChatTableName, map[string]any{
		`from_type`:          fromType,
		`from_id`:            fromID,
		`prompt`:             prompt,
		`prompt_type`:        promptType,
		`cli_type`:           cliType,
		`agent_cli_id`:       agentCliID,
		`local_dir`:          localDir,
		`settings_path`:      settingsPath,
		`model_name`:         modelName,
		`thinking_collapsed`: thinkingCollapsed,
		`thinking_intensity`: thinkingIntensity,
		`is_read`:            agentChatReadYes,
		`status`:             taskWorkflowChatStatusRunning,
		`raw_output`:         ``,
		`created_at`:         now,
		`updated_at`:         now,
	}).Exec()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// TaskWorkflowChatCreate 创建工作流对话记录。
// TaskWorkflowChatCreate keeps the workflow-specific call site stable while writing into the generic agent chat table.
func (h *CSqlite) TaskWorkflowChatCreate(workflowID int, prompt, promptType, cliType string, agentCliID int, localDir, settingsPath, modelName string, thinkingCollapsed int, thinkingIntensity string) (int64, error) {
	return h.AgentChatCreateBySource(AgentChatSourceTypeWorkflow, workflowID, prompt, promptType, cliType, agentCliID, localDir, settingsPath, modelName, thinkingCollapsed, thinkingIntensity)
}

// AgentChatCreate 创建 AgentCli 独立执行对话记录。
// AgentChatCreate creates a standalone AgentCli chat while still using the shared chat table.
func (h *CSqlite) AgentChatCreate(agentCliID int, prompt, promptType, cliType string, localDir, settingsPath, modelName string, thinkingCollapsed int, thinkingIntensity string) (int64, error) {
	return h.AgentChatCreateBySource(AgentChatSourceTypeAgentCli, agentCliID, prompt, promptType, cliType, agentCliID, localDir, settingsPath, modelName, thinkingCollapsed, thinkingIntensity)
}

// TaskWorkflowChatUpdateSessionID 更新 session_id。
func (h *CSqlite) TaskWorkflowChatUpdateSessionID(chatID int64, sessionID string) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(agentChatTableName, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`session_id`: sessionID,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatUpdatePID 更新进程 PID。
func (h *CSqlite) TaskWorkflowChatUpdatePID(chatID int64, pid int) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(agentChatTableName, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`pid`:        pid,
		`updated_at`: now,
	}).Exec()
	return err
}

const (
	// ChatOutputFlushBatchSize SSE 对话输出批量写 DB 的行数阈值
	ChatOutputFlushBatchSize = 200
)

// TaskWorkflowChatAppendOutput 追加一行 raw_output。
func (h *CSqlite) TaskWorkflowChatAppendOutput(chatID int64, line string) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	info, err := h.TaskWorkflowChatInfo(chatID)
	if err != nil {
		return err
	}
	current := cast.ToString(info[`raw_output`])
	newOutput := current
	if current != `` {
		newOutput += "\n"
	}
	newOutput += line
	_, err = h.Client.QuickUpdate(agentChatTableName, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`raw_output`: newOutput,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatAppendOutputBatch 批量追加多行 raw_output，一次 DB 读写完成。
func (h *CSqlite) TaskWorkflowChatAppendOutputBatch(chatID int64, lines []string) error {
	if len(lines) == 0 {
		return nil
	}
	now := time.Now().Format(`2006-01-02 15:04:05`)
	info, err := h.TaskWorkflowChatInfo(chatID)
	if err != nil {
		return err
	}
	current := cast.ToString(info[`raw_output`])
	newOutput := current
	if current != `` {
		newOutput += "\n"
	}
	newOutput += strings.Join(lines, "\n")
	_, err = h.Client.QuickUpdate(agentChatTableName, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`raw_output`: newOutput,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatMarkCompleted 标记对话完成。
func (h *CSqlite) TaskWorkflowChatMarkCompleted(chatID int64) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(agentChatTableName, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`is_read`:    agentChatReadNo,
		`status`:     taskWorkflowChatStatusCompleted,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatMarkError 标记对话异常终止。
func (h *CSqlite) TaskWorkflowChatMarkError(chatID int64) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(agentChatTableName, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`is_read`:    agentChatReadNo,
		`status`:     taskWorkflowChatStatusError,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatRecoverInterrupted 启动时将所有 running 状态的记录标记为 interrupted（进程已随上次进程退出而终止）。
func (h *CSqlite) TaskWorkflowChatRecoverInterrupted() {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	upNumber, err := h.Client.ExecBySql(
		`update agent_chat set status = ?, updated_at = ? where status = ?`,
		taskWorkflowChatStatusInterrupted, now, taskWorkflowChatStatusRunning,
	).Exec()
	if err != nil {
		gstool.FmtPrintlnLogTime(`TaskWorkflowChatRecoverInterrupted 失败: %v`, err)
	} else {
		gstool.FmtPrintlnLogTime(`[agent cli] 更新状态进行中的为异常中断，数量%d`, upNumber)
	}
}

// TaskWorkflowChatInfo 获取单条对话记录。
func (h *CSqlite) TaskWorkflowChatInfo(chatID int64) (map[string]any, error) {
	return h.Client.QuickQuery(agentChatTableName, `*`, map[string]any{
		`id`: chatID,
	}).One()
}

// TaskWorkflowChatList 获取 workflow 下所有对话记录。
func (h *CSqlite) TaskWorkflowChatList(workflowID int) ([]map[string]any, error) {
	rows, err := h.Client.QuickQuery(agentChatTableName, `*`, map[string]any{
		`from_type`: AgentChatSourceTypeWorkflow,
		`from_id`:   workflowID,
	}).Order(`id DESC`).All()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// TaskWorkflowChatMarkRunning 标记对话为运行中（用于继续对话）。
// continuePrompt 非空时同步更新 prompt 字段，用于继续对话时保存用户新输入。
func (h *CSqlite) TaskWorkflowChatMarkRunning(chatID int64, continuePrompt string) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	gstool.FmtPrintlnLogTime("[db-mark-running] chat_id=%d 准备将状态更新为running，时间=%s", chatID, now)
	updateFields := map[string]any{
		`is_read`:    agentChatReadYes,
		`status`:     taskWorkflowChatStatusRunning,
		`updated_at`: now,
	}
	if continuePrompt != `` {
		updateFields[`prompt`] = continuePrompt
	}
	_, err := h.Client.QuickUpdate(agentChatTableName, map[string]any{
		`id`: chatID,
	}, updateFields).Exec()
	if err != nil {
		gstool.FmtPrintlnLogTime("[db-mark-running] chat_id=%d 更新失败: %v", chatID, err)
		return err
	}
	gstool.FmtPrintlnLogTime("[db-mark-running] chat_id=%d 更新成功", chatID)

	// 立即验证写入是否生效
	verifyInfo, verifyErr := h.TaskWorkflowChatInfo(chatID)
	if verifyErr != nil {
		gstool.FmtPrintlnLogTime("[db-mark-running] chat_id=%d 验证查询失败: %v", chatID, verifyErr)
	} else {
		verifyStatus := cast.ToString(verifyInfo[`status`])
		gstool.FmtPrintlnLogTime("[db-mark-running] chat_id=%d 验证查询状态=%s", chatID, verifyStatus)
	}
	return nil
}

// TaskWorkflowChatMarkInterrupted 标记对话为用户主动中断。
func (h *CSqlite) TaskWorkflowChatMarkInterrupted(chatID int64) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(agentChatTableName, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`is_read`:    agentChatReadNo,
		`status`:     taskWorkflowChatStatusInterrupted,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatListByPromptType 按提示词类型查询对话历史。
func (h *CSqlite) TaskWorkflowChatListByPromptType(workflowID int, promptType string) ([]map[string]any, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	promptType = strings.TrimSpace(promptType)
	if promptType == `` {
		return nil, errors.New(`提示词类型不能为空`)
	}
	rows, err := h.Client.QuickQuery(agentChatTableName, `*`, map[string]any{
		`from_type`:   AgentChatSourceTypeWorkflow,
		`from_id`:     workflowID,
		`prompt_type`: promptType,
	}).Order(`id DESC`).All()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// AgentChatListByAgentCli 按 Agent CLI 查询独立执行对话历史。
// AgentChatListByAgentCli only returns standalone AgentCli runs so workflow chats do not pollute the history card.
func (h *CSqlite) AgentChatListByAgentCli(agentCliID int) ([]map[string]any, error) {
	if agentCliID <= 0 {
		return nil, errors.New(`agent_cli_id不能为空`)
	}
	rows, err := h.Client.QuickQuery(agentChatTableName, `*`, map[string]any{
		`from_type`:    AgentChatSourceTypeAgentCli,
		`agent_cli_id`: agentCliID,
	}).Order(`id DESC`).All()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// AgentChatListAll 返回所有 AgentCli 独立执行历史，用于提取共享工作目录。
func (h *CSqlite) AgentChatListAll() ([]map[string]any, error) {
	rows, err := h.Client.QuickQuery(agentChatTableName, `*`, map[string]any{
		`from_type`: AgentChatSourceTypeAgentCli,
	}).Order(`id DESC`).All()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// AgentChatMarkRead 将一个已结束对话标记为已读。
func (h *CSqlite) AgentChatMarkRead(chatID int64) error {
	if chatID <= 0 {
		return errors.New(`chat_id不能为空`)
	}
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(agentChatTableName, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`is_read`:    agentChatReadYes,
		`updated_at`: now,
	}).Exec()
	return err
}

// AgentChatUnreadCountByWorkflow 汇总一个工作流下所有未读历史对话数。

// AgentChatUnreadCountByAgentCli 汇总一个 Agent CLI 下所有未读历史对话数。
func (h *CSqlite) AgentChatUnreadCountByAgentCli(agentCliID int) (int, error) {
	if agentCliID <= 0 {
		return 0, errors.New(`agent_cli_id不能为空`)
	}
	row, err := h.Client.QueryBySql(
		`SELECT COUNT(1) AS total FROM agent_chat WHERE from_type = ? AND agent_cli_id = ? AND is_read = ? AND status <> ?`,
		AgentChatSourceTypeAgentCli, agentCliID, agentChatReadNo, taskWorkflowChatStatusRunning,
	).One()
	if err != nil {
		return 0, err
	}
	return cast.ToInt(row[`total`]), nil
}

// TaskWorkflowClearChatSessionIDs 删除指定 prompt_type 的所有对话记录。
func (h *CSqlite) TaskWorkflowClearChatSessionIDs(workflowID int, promptType string) error {
	promptType = strings.TrimSpace(promptType)
	if promptType == `` {
		return nil
	}
	_, err := h.Client.QuickDelete(agentChatTableName, map[string]any{
		`from_type`:   AgentChatSourceTypeWorkflow,
		`from_id`:     workflowID,
		`prompt_type`: promptType,
	}).Exec()
	return err
}

// ===================== 模板关联查询 =====================

// HomeTaskWorkflowTemplateID 查询任务关联的模板ID。
func (h *CSqlite) HomeTaskWorkflowTemplateID(homeTaskID int) (int, error) {
	if homeTaskID <= 0 {
		return 0, nil
	}
	info, err := h.HomeTaskRow(homeTaskID)
	if err != nil {
		return 0, err
	}
	if len(info) == 0 {
		return 0, nil
	}
	return cast.ToInt(info[`workflow_template_id`]), nil
}

// HomeTaskWorkflowTemplateSteps 根据 homeTaskID 获取关联模板的步骤列表。
// 如果任务未关联模板，则返回默认模板的步骤。
func (h *CSqlite) HomeTaskWorkflowTemplateSteps(homeTaskID int) (map[string]any, []map[string]any, error) {
	templateID, err := h.HomeTaskWorkflowTemplateID(homeTaskID)
	if err != nil {
		return nil, nil, err
	}

	var template map[string]any
	if templateID <= 0 {
		// 未关联模板，使用默认模板
		template, err = h.WorkflowTemplateDefaultInfo()
	} else {
		template, err = h.WorkflowTemplateInfo(templateID)
	}
	if err != nil {
		return nil, nil, err
	}

	// 从 template map 中提取 steps
	steps := make([]map[string]any, 0)
	if rawSteps, ok := template[`steps`]; ok {
		switch v := rawSteps.(type) {
		case []map[string]any:
			steps = v
		}
	}

	return template, steps, nil
}

// ===================== 工作流文档独立表操作 =====================

const (
	taskWorkflowDocumentTable = `tbl_task_workflow_document`

	// 文档类型常量
	TaskWorkflowDocTypeRequirement           = `requirement`
	TaskWorkflowDocTypeDevPlan               = `dev_plan`
	TaskWorkflowDocTypePlainTextRequirement  = `plain_text_requirement`
	TaskWorkflowDocTypeDesignPlanRequirement = `design_plan_requirement`
	TaskWorkflowDocTypeApiDoc                = `api_doc`
	TaskWorkflowDocTypeDesign                = `design`
	TaskWorkflowDocTypeStepDocument          = `step_document`
)

// taskWorkflowDocTypeName 文档类型到可读名称的映射。
var taskWorkflowDocTypeName = map[string]string{
	TaskWorkflowDocTypeRequirement:           `需求文档`,
	TaskWorkflowDocTypeDevPlan:               `开发执行计划`,
	TaskWorkflowDocTypePlainTextRequirement:  `纯文本需求文档`,
	TaskWorkflowDocTypeDesignPlanRequirement: `设计方案需求文档`,
	TaskWorkflowDocTypeApiDoc:                `接口文档`,
	TaskWorkflowDocTypeDesign:                `设计文档`,
	TaskWorkflowDocTypeStepDocument:          `步骤文档`,
}

// TaskWorkflowDocumentTypeName 返回文档类型的可读名称。
func TaskWorkflowDocumentTypeName(docType string) string {
	if name, ok := taskWorkflowDocTypeName[docType]; ok {
		return name
	}
	return docType
}

// TaskWorkflowDocumentUpsert 创建或更新工作流文档记录。
// 按 (workflow_id, document_type, template_step_id, document_id) 联合唯一键 upsert，
// document_id 用于区分同一步骤下的多个文档，避免互相覆盖。
func (h *CSqlite) TaskWorkflowDocumentUpsert(workflowID int, docID, docName, docType string, templateID, templateStepID int, fileID, folderName, placeholder string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	docType = strings.TrimSpace(docType)
	if docType == `` {
		return errors.New(`文档类型不能为空`)
	}
	now := time.Now().Unix()
	// 查询是否已存在
	existing, err := h.Client.QuickQuery(taskWorkflowDocumentTable, `id`, map[string]any{
		`workflow_id`:      workflowID,
		`document_type`:    docType,
		`template_step_id`: templateStepID,
		`document_id`:      docID,
	}).One()
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		_, err = h.Client.QuickUpdate(taskWorkflowDocumentTable, map[string]any{
			`id`: existing[`id`],
		}, map[string]any{
			`document_name`: docName,
			`file_id`:       fileID,
			`folder_name`:   folderName,
			`placeholder`:   placeholder,
			`update_time`:   now,
		}).Exec()
		return err
	}
	_, err = h.Client.QuickCreate(taskWorkflowDocumentTable, map[string]any{
		`workflow_id`:      workflowID,
		`document_id`:      docID,
		`document_name`:    docName,
		`document_type`:    docType,
		`template_id`:      templateID,
		`template_step_id`: templateStepID,
		`file_id`:          fileID,
		`folder_name`:      folderName,
		`placeholder`:      placeholder,
		`create_time`:      now,
		`update_time`:      now,
	}).Exec()
	return err
}

// TaskWorkflowDocumentList 查询工作流的所有文档记录。
func (h *CSqlite) TaskWorkflowDocumentList(workflowID int) ([]map[string]any, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	return h.Client.QuickQuery(taskWorkflowDocumentTable, `*`, map[string]any{
		`workflow_id`: workflowID,
	}).Order(`id ASC`).All()
}

// TaskWorkflowDocumentDeleteByWorkflow 删除工作流的所有文档记录。
func (h *CSqlite) TaskWorkflowDocumentDeleteByWorkflow(workflowID int) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	_, err := h.Client.QuickDelete(taskWorkflowDocumentTable, map[string]any{
		`workflow_id`: workflowID,
	}).Exec()
	return err
}

// TaskWorkflowDocumentHasRecords 判断工作流在新表中是否有文档记录。
func (h *CSqlite) TaskWorkflowDocumentHasRecords(workflowID int) bool {
	if workflowID <= 0 {
		return false
	}
	cnt, err := h.Client.QuickQuery(taskWorkflowDocumentTable, `COUNT(1) AS cnt`, map[string]any{
		`workflow_id`: workflowID,
	}).One()
	if err != nil {
		return false
	}
	return cast.ToInt(cnt[`cnt`]) > 0
}
