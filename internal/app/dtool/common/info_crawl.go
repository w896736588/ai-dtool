package common

import (
	"dev_tool/internal/app/dtool/define"
	"errors"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// InfoCrawlTaskList 查询任务列表。
func (h *CSqlite) InfoCrawlTaskList() ([]map[string]any, error) {
	list, err := h.Client.QueryBySql(`
select id,name,prompt,ai_model_id,status,create_time,update_time
from tbl_info_crawl_task
where status = ?
order by update_time desc, id desc`, define.InfoCrawlTaskStatusNormal).All()
	if err != nil {
		return nil, err
	}
	h.infoCrawlFillTimeDesc(list)
	return list, nil
}

// InfoCrawlTaskInfo 查询任务详情。
func (h *CSqlite) InfoCrawlTaskInfo(id int) (map[string]any, error) {
	task, err := h.InfoCrawlTaskRow(id)
	if err != nil {
		return nil, err
	}
	runList, err := h.InfoCrawlRunList(id, 20)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		`task`:     task,
		`run_list`: runList,
	}, nil
}

// InfoCrawlTaskRow 查询单个任务。
func (h *CSqlite) InfoCrawlTaskRow(id int) (map[string]any, error) {
	task, err := h.Client.QuickQuery(`tbl_info_crawl_task`, `*`, map[string]any{
		`id`:     id,
		`status`: define.InfoCrawlTaskStatusNormal,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(task) == 0 {
		return nil, errors.New(`任务不存在`)
	}
	h.infoCrawlFillRowTimeDesc(task)
	return task, nil
}

// InfoCrawlTaskSave 保存任务。
func (h *CSqlite) InfoCrawlTaskSave(id int, name, prompt string, aiModelID int) (map[string]any, error) {
	now := time.Now().Unix()
	name = strings.TrimSpace(name)
	prompt = strings.TrimSpace(prompt)
	if name == `` {
		return nil, errors.New(`任务名称不能为空`)
	}
	if prompt == `` {
		return nil, errors.New(`任务提示词不能为空`)
	}
	if aiModelID <= 0 {
		return nil, errors.New(`请选择AI模型`)
	}
	if _, err := h.InfoCrawlAiModelInfo(aiModelID); err != nil {
		return nil, err
	}
	if id <= 0 {
		newID, err := h.Client.QuickCreate(`tbl_info_crawl_task`, map[string]any{
			`name`:        name,
			`prompt`:      prompt,
			`ai_model_id`: aiModelID,
			`status`:      define.InfoCrawlTaskStatusNormal,
			`create_time`: now,
			`update_time`: now,
		}).Exec()
		if err != nil {
			return nil, err
		}
		id = cast.ToInt(newID)
	} else {
		_, err := h.Client.QuickUpdate(`tbl_info_crawl_task`, map[string]any{
			`id`:     id,
			`status`: define.InfoCrawlTaskStatusNormal,
		}, map[string]any{
			`name`:        name,
			`prompt`:      prompt,
			`ai_model_id`: aiModelID,
			`update_time`: now,
		}).Exec()
		if err != nil {
			return nil, err
		}
	}
	return h.InfoCrawlTaskRow(id)
}

// InfoCrawlTaskDelete 软删除任务。
func (h *CSqlite) InfoCrawlTaskDelete(id int) error {
	if id <= 0 {
		return errors.New(`任务id不能为空`)
	}
	_, err := h.Client.QuickUpdate(`tbl_info_crawl_task`, map[string]any{
		`id`:     id,
		`status`: define.InfoCrawlTaskStatusNormal,
	}, map[string]any{
		`status`:      define.InfoCrawlTaskStatusDelete,
		`update_time`: time.Now().Unix(),
	}).Exec()
	return err
}

// InfoCrawlRunCreate 创建执行记录。
func (h *CSqlite) InfoCrawlRunCreate(taskID int, taskInfo map[string]any, aiModelInfo map[string]any) (int, error) {
	now := time.Now().Unix()
	newID, err := h.Client.QuickCreate(`tbl_info_crawl_run`, map[string]any{
		`task_id`:           taskID,
		`status`:            define.InfoCrawlRunStatusRunning,
		`run_message`:       `任务已提交，准备执行`,
		`prompt_snapshot`:   cast.ToString(taskInfo[`prompt`]),
		`ai_model_snapshot`: gstool.JsonEncode(aiModelInfo),
		`output_content`:    ``,
		`error_message`:     ``,
		`create_time`:       now,
		`update_time`:       now,
	}).Exec()
	return cast.ToInt(newID), err
}

// InfoCrawlRunUpdate 更新执行记录。
func (h *CSqlite) InfoCrawlRunUpdate(id int, updateData map[string]any) error {
	updateData[`update_time`] = time.Now().Unix()
	_, err := h.Client.QuickUpdate(`tbl_info_crawl_run`, map[string]any{
		`id`: id,
	}, updateData).Exec()
	return err
}

// InfoCrawlRunList 查询执行历史。
func (h *CSqlite) InfoCrawlRunList(taskID, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 20
	}
	list, err := h.Client.QueryBySql(`
select id,task_id,status,run_message,prompt_snapshot,ai_model_snapshot,output_content,error_message,create_time,update_time
from tbl_info_crawl_run
where task_id = ?
order by id desc
limit ?`, taskID, limit).All()
	if err != nil {
		return nil, err
	}
	h.infoCrawlFillTimeDesc(list)
	return list, nil
}

// InfoCrawlRunInfo 查询执行详情。
func (h *CSqlite) InfoCrawlRunInfo(id int) (map[string]any, error) {
	runInfo, err := h.Client.QuickQuery(`tbl_info_crawl_run`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(runInfo) == 0 {
		return nil, errors.New(`执行记录不存在`)
	}
	h.infoCrawlFillTimeDesc([]map[string]any{runInfo})
	return map[string]any{
		`run_info`: runInfo,
	}, nil
}

// InfoCrawlAiModelInfo 查询 AI 模型配置。
func (h *CSqlite) InfoCrawlAiModelInfo(id int) (map[string]any, error) {
	info, err := h.Client.QueryBySql(`
select m.*,p.name as provider_name,p.provider_type,p.base_url,p.api_key
from tbl_ai_model m
left join tbl_ai_provider p on p.id = m.provider_id
where m.id = ? and m.status = 1 and p.status = 1`, id).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`AI模型不存在或已停用`)
	}
	return info, nil
}

// infoCrawlFillTimeDesc 填充时间描述字段。
func (h *CSqlite) infoCrawlFillTimeDesc(list []map[string]any) {
	for _, item := range list {
		h.infoCrawlFillRowTimeDesc(item)
	}
}

// infoCrawlFillRowTimeDesc 填充单行时间描述字段。
func (h *CSqlite) infoCrawlFillRowTimeDesc(row map[string]any) {
	row[`create_time_desc`] = h.infoCrawlFormatTime(cast.ToInt64(row[`create_time`]))
	row[`update_time_desc`] = h.infoCrawlFormatTime(cast.ToInt64(row[`update_time`]))
}

// infoCrawlFormatTime 格式化时间。
func (h *CSqlite) infoCrawlFormatTime(unixTime int64) string {
	if unixTime <= 0 {
		return ``
	}
	return gstool.TimeUnixToString(time.Unix(unixTime, 0), `Y-m-d H:i:s`)
}
