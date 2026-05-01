package common

import (
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

var DbMain = &CSqlite{}
var DbLog = &CSqlite{}

type CSqlite struct {
	Client *gsdb.GsSqlite
	Env    *define.Env
}

func (h *CSqlite) InitTable() {
	//TODO 初始化表机构和变更
}

func (h *CSqlite) Login(username, password string) (int, error) {
	one, err := h.Client.QuickQuery(`tbl_user`, `*`, map[string]interface{}{
		`username`: username,
		`password`: h.GetSaltPassword(password),
	}).One()
	if err != nil {
		return 0, err
	}
	if len(one) > 0 {
		return cast.ToInt(one[`id`]), nil
	} else {
		return 0, nil
	}
}

func (h *CSqlite) CreateUser(username, password string) (int, error) {
	one, err := h.Client.QuickQuery(`tbl_user`, `*`, map[string]interface{}{
		`username`: username,
	}).One()
	if err != nil {
		return 0, err
	}
	if len(one) > 0 {
		return 0, errors.New(`已存在用户`)
	}
	newId, newError := h.Client.QuickCreate(`tbl_user`, map[string]interface{}{
		`username`: username,
		`password`: h.GetSaltPassword(password),
	}).Exec()
	if newError != nil {
		return 0, newError
	}
	return cast.ToInt(newId), nil
}

func (h *CSqlite) GetSaltPassword(password string) string {
	return gstool.Md5(password + h.Env.AppName)
}

func (h *CSqlite) CreateSsh(name, host, port string, userid, isPublic int) (int, error) {
	newId, newError := h.Client.QuickCreate(`tbl_ssh`, map[string]interface{}{
		`name`:      name,
		`host`:      host,
		`port`:      port,
		`userid`:    userid,
		`is_public`: isPublic,
	}).Exec()
	if newError != nil {
		return 0, newError
	}
	return cast.ToInt(newId), nil
}

func (h *CSqlite) CreateSshUser(sshid, userid int, username, password string) (int, error) {
	one, err := h.Client.QuickQuery(`tbl_user_ssh`, `*`, map[string]interface{}{
		`ssh_id`:  sshid,
		`user_id`: userid,
	}).One()
	if err != nil {
		return 0, err
	}
	if len(one) > 0 {
		_, editError := h.Client.QuickUpdate(`tbl_user_ssh`, map[string]interface{}{
			`ssh_id`:  sshid,
			`user_id`: userid,
		}, map[string]interface{}{
			`username`: username,
			`password`: password,
		}).Exec()
		if editError != nil {
			return 0, editError
		}
		return cast.ToInt(one[`id`]), nil
	} else {
		newId, newError := h.Client.QuickCreate(`tbl_user_ssh`, map[string]interface{}{
			`ssh_id`:   sshid,
			`user_id`:  userid,
			`username`: username,
			`password`: password,
		}).Exec()
		if newError != nil {
			return 0, newError
		}
		return cast.ToInt(newId), nil
	}
}

func (h *CSqlite) CreateCode(name, path string, sshid, codeGroupId int) (int, error) {
	one, err := h.Client.QuickQuery(`tbl_user_ssh`, `*`, map[string]interface{}{
		`ssh_id`: sshid,
		`path`:   path,
	}).One()
	if err != nil {
		return 0, err
	}
	if len(one) > 0 {
		_, editError := h.Client.QuickUpdate(`tbl_user_ssh`, map[string]interface{}{
			`ssh_id`: sshid,
			`path`:   path,
		}, map[string]interface{}{
			`name`:          name,
			`code_group_id`: codeGroupId,
		}).Exec()
		if editError != nil {
			return 0, editError
		}
		return cast.ToInt(one[`id`]), nil
	} else {
		newId, newError := h.Client.QuickCreate(`tbl_user_ssh`, map[string]interface{}{
			`name`:          name,
			`path`:          path,
			`code_group_id`: codeGroupId,
			`ssh_id`:        sshid,
		}).Exec()
		if newError != nil {
			return 0, newError
		}
		return cast.ToInt(newId), nil
	}
}

func (h *CSqlite) GetSshConfigByUserSshId(userSshId int) (map[string]any, error) {
	sql := `select us.username,us.password,s.host,s.port from tbl_user_ssh us 
			left join tbl_ssh s on s.id = us.ssh_id 
			where us.id = ?`
	return h.Client.QueryBySql(sql, userSshId).One()
}

func (h *CSqlite) GetSshConfigByUserId(userId int) ([]map[string]any, error) {
	sql := `select us.username,us.password,s.host,s.port,us.id from tbl_user_ssh us 
			left join tbl_ssh s on s.id = us.ssh_id 
			where us.user_id = ?`
	return h.Client.QueryBySql(sql, userId).All()
}

func (h *CSqlite) GetAllSshConfig() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
}

func (h *CSqlite) GetSshConfig(sshId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_ssh`, `*`, map[string]any{
		`id`: sshId,
	}).One()
}

func (h *CSqlite) GetRedisConfig(redisId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_redis`, `*`, map[string]any{
		`id`: redisId,
	}).One()
}

func (h *CSqlite) GetAllRedisConfig() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_redis`, `*`, nil).All()
}

func (h *CSqlite) GetMysqlConfig(mysqlId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_mysql`, `*`, map[string]any{
		`id`: mysqlId,
	}).One()
}

func (h *CSqlite) StarAdd(id, name, key, value, _type any) (int64, error) {
	if cast.ToInt(id) == 0 {
		return h.Client.QuickCreate(`tbl_star`, map[string]any{
			`name`:        name,
			`key`:         key,
			`value`:       value,
			`type`:        _type,
			`create_time`: time.Now().Unix(),
			`update_time`: time.Now().Unix(),
		}).Exec()
	} else {
		return h.Client.QuickUpdate(`tbl_star`, map[string]any{
			`id`: id,
		}, map[string]any{
			`name`:        name,
			`key`:         key,
			`value`:       value,
			`type`:        _type,
			`update_time`: time.Now().Unix(),
		}).Exec()
	}
}

func (h *CSqlite) StarDel(id any) (int64, error) {
	return h.Client.QuickDelete(`tbl_star`, map[string]any{
		`id`: id,
	}).Exec()
}

func (h *CSqlite) StarList(_type any) ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_star`, `*`, map[string]any{
		`type`: _type,
	}).All()
}

func (h *CSqlite) MarkdownAdd(id, name, markdownType, value any) (int64, error) {
	if cast.ToInt(id) == 0 {
		return h.Client.QuickCreate(`tbl_markdown`, map[string]any{
			`name`:          name,
			`content`:       value,
			`markdown_type`: markdownType,
			`create_time`:   time.Now().Unix(),
			`update_time`:   time.Now().Unix(),
		}).Exec()
	} else {
		//记录变更记录
		oldInfo, _ := h.Client.QuickQuery(`tbl_markdown`, `content`, map[string]any{
			`id`: id,
		}).One()

		upNum, upErr := h.Client.QuickUpdate(`tbl_markdown`, map[string]any{
			`id`: id,
		}, map[string]any{
			`name`:        name,
			`content`:     value,
			`update_time`: time.Now().Unix(),
		}).Exec()
		if upErr == nil && upNum > 0 && cast.ToString(oldInfo[`content`]) != value {

			_, _ = h.Client.QuickCreate(`tbl_markdown_history`, map[string]any{
				`markdown_id`: id,
				`old_content`: oldInfo[`content`],
				`new_content`: value,
				`change_desc`: p_common.TBaseClient.DiffText(cast.ToString(oldInfo[`content`]), cast.ToString(value)),
				`create_time`: time.Now().Unix(),
				`update_time`: time.Now().Unix(),
			}).Exec()
		}
		return upNum, upErr
	}
}

func (h *CSqlite) MarkdownDel(id any) (int64, error) {
	return h.Client.QuickDelete(`tbl_markdown`, map[string]any{
		`id`: id,
	}).Exec()
}

func (h *CSqlite) MarkdownList(markdownType string) ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_markdown`, `*`, map[string]any{
		`markdown_type`: markdownType,
	}).Order("weight asc").All()
}

func (h *CSqlite) MarkdownHistoryList(id int) ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_markdown_history`, `*`, map[string]any{
		`markdown_id`: id,
	}).Order("id desc").All()
}

func (h *CSqlite) MarkdownHistoryDel(historyId any) (int64, error) {
	return h.Client.QuickDelete(`tbl_markdown_history`, map[string]any{
		`id`: historyId,
	}).Exec()
}

func (h *CSqlite) AllGlobal() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_global`, `*`, map[string]any{}).All()
}

func (h *CSqlite) AllGlobalMap() (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_global`, `*`, map[string]any{}).ToMap(`key`, `value`)
}

func (h *CSqlite) GlobalValue(key string) (string, error) {
	one, err := h.Client.QuickQuery(`tbl_global`, `*`, map[string]any{
		`key`: key,
	}).Order(`id asc`).One()
	if err != nil {
		return ``, err
	}
	return cast.ToString(one[`value`]), nil
}

func (h *CSqlite) SetGlobalValue(name, key, value, desc string) error {
	now := time.Now().Unix()
	one, err := h.Client.QuickQuery(`tbl_global`, `*`, map[string]any{
		`key`: key,
	}).Order(`id asc`).One()
	if err != nil {
		return err
	}
	updateData := map[string]any{
		`name`:        name,
		`key`:         key,
		`value`:       value,
		`desc`:        desc,
		`update_time`: now,
	}
	if cast.ToInt(one[`id`]) > 0 {
		_, err = h.Client.QuickUpdate(`tbl_global`, map[string]any{
			`id`: one[`id`],
		}, updateData).Exec()
		return err
	}
	updateData[`create_time`] = now
	_, err = h.Client.QuickCreate(`tbl_global`, updateData).Exec()
	return err
}

// CronTaskByType 按 type 查询单条定时任务。
func (h *CSqlite) CronTaskByType(taskType string) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_cron_task`, `*`, map[string]any{
		`type`: taskType,
	}).Order(`id asc`).One()
}

// CronTaskList 查询所有定时任务。
func (h *CSqlite) CronTaskList() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_cron_task`, `*`, nil).Order(`id asc`).All()
}

// CronTaskSave 保存定时任务配置（按 type upsert）。
func (h *CSqlite) CronTaskSave(taskType, name string, enabled int, triggerTime string) error {
	now := time.Now().Unix()
	one, err := h.Client.QuickQuery(`tbl_cron_task`, `*`, map[string]any{
		`type`: taskType,
	}).Order(`id asc`).One()
	if err != nil && !DbRowMissing(err) {
		return err
	}
	updateData := map[string]any{
		`name`:         name,
		`type`:         taskType,
		`enabled`:      enabled,
		`trigger_time`: triggerTime,
		`update_time`:  now,
	}
	if cast.ToInt(one[`id`]) > 0 {
		_, err = h.Client.QuickUpdate(`tbl_cron_task`, map[string]any{
			`id`: one[`id`],
		}, updateData).Exec()
		return err
	}
	updateData[`create_time`] = now
	_, err = h.Client.QuickCreate(`tbl_cron_task`, updateData).Exec()
	return err
}

// CronTaskUpdateLastTriggerTime 更新定时任务最后触发时间。
func (h *CSqlite) CronTaskUpdateLastTriggerTime(taskType string) error {
	now := time.Now().Unix()
	_, err := h.Client.QuickUpdate(`tbl_cron_task`, map[string]any{
		`type`: taskType,
	}, map[string]any{
		`last_trigger_time`: now,
		`update_time`:       now,
	}).Exec()
	return err
}

// DbRowMissing 判断数据库查询是否因行不存在而报错。
func DbRowMissing(err error) bool {
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, `not found`) || strings.Contains(errText, `no rows`)
}

// HomeTaskConfigValue 按 key 从 tbl_home_task_config 读取配置值。
func (h *CSqlite) HomeTaskConfigValue(key string) (string, error) {
	one, err := h.Client.QuickQuery(`tbl_home_task_config`, `*`, map[string]any{
		`key`: key,
	}).Order(`id asc`).One()
	if err != nil {
		return ``, err
	}
	return cast.ToString(one[`value`]), nil
}

// HomeTaskConfigSave 按 key 保存首页任务配置（upsert）。
func (h *CSqlite) HomeTaskConfigSave(name, key, value, desc string) error {
	now := time.Now().Unix()
	one, err := h.Client.QuickQuery(`tbl_home_task_config`, `*`, map[string]any{
		`key`: key,
	}).Order(`id asc`).One()
	if err != nil && !DbRowMissing(err) {
		return err
	}
	updateData := map[string]any{
		`name`:        name,
		`key`:         key,
		`value`:       value,
		`desc`:        desc,
		`update_time`: now,
	}
	if cast.ToInt(one[`id`]) > 0 {
		_, err = h.Client.QuickUpdate(`tbl_home_task_config`, map[string]any{
			`id`: one[`id`],
		}, updateData).Exec()
		return err
	}
	updateData[`create_time`] = now
	_, err = h.Client.QuickCreate(`tbl_home_task_config`, updateData).Exec()
	return err
}

// MemoryConfigValue 按 key 从 tbl_memory_config 读取配置值。
func (h *CSqlite) MemoryConfigValue(key string) (string, error) {
	one, err := h.Client.QuickQuery(`tbl_memory_config`, `*`, map[string]any{
		`key`: key,
	}).Order(`id asc`).One()
	if err != nil {
		return ``, err
	}
	return cast.ToString(one[`value`]), nil
}

// MemoryConfigSave 按 key 保存记忆配置（upsert）。
func (h *CSqlite) MemoryConfigSave(name, key, value, desc string) error {
	now := time.Now().Unix()
	one, err := h.Client.QuickQuery(`tbl_memory_config`, `*`, map[string]any{
		`key`: key,
	}).Order(`id asc`).One()
	if err != nil && !DbRowMissing(err) {
		return err
	}
	updateData := map[string]any{
		`name`:        name,
		`key`:         key,
		`value`:       value,
		`desc`:        desc,
		`update_time`: now,
	}
	if cast.ToInt(one[`id`]) > 0 {
		_, err = h.Client.QuickUpdate(`tbl_memory_config`, map[string]any{
			`id`: one[`id`],
		}, updateData).Exec()
		return err
	}
	updateData[`create_time`] = now
	_, err = h.Client.QuickCreate(`tbl_memory_config`, updateData).Exec()
	return err
}

func (h *CSqlite) CmdList(variableId any) ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`variable_id`: variableId,
		`status`:      1,
	}).Order(`weight asc`).All()
}

func (h *CSqlite) CmdInfo(cmdId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`id`:     cmdId,
		`status`: 1,
	}).One()
}

func (h *CSqlite) Variable(variableId any) map[string]any {
	variableInfo, _ := h.Client.QuickQuery(`tbl_variable`, `*`, map[string]interface{}{
		`id`: variableId,
	}).One()
	return variableInfo
}

func (h *CSqlite) GetApiInfo(id int) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_api`, `*`, map[string]interface{}{
		`id`: id,
	}).One()
}

// MemoryFragmentList 查询知识片段列表。
func (h *CSqlite) MemoryFragmentList(limit int) ([]map[string]any, error) {
	list, err := h.memoryFragmentListByDeleted(0, limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// MemoryFragmentTrashList 查询回收站中的知识片段列表。
func (h *CSqlite) MemoryFragmentTrashList(limit int) ([]map[string]any, error) {
	list, err := h.memoryFragmentListByDeleted(1, limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// MemoryFragmentInfo 查询单个知识片段详情。
func (h *CSqlite) MemoryFragmentInfo(id int) (map[string]any, error) {
	one, err := h.Client.QuickQuery(`tbl_memory_fragment`, `*`, map[string]any{
		`id`:         id,
		`is_deleted`: 0,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(one) == 0 {
		return nil, errors.New(`片段不存在`)
	}
	tags, tagErr := h.memoryFragmentLoadTags(id)
	if tagErr != nil {
		return nil, tagErr
	}
	one[`tags`] = tags
	one[`create_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(one[`create_time`]))
	one[`update_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(one[`update_time`]))
	delete(one, `index_status`)
	delete(one, `index_version`)
	return one, nil
}

// MemoryFragmentSave 保存知识片段。
func (h *CSqlite) MemoryFragmentSave(id int, title, content string, tags []string) (map[string]any, error) {
	now := time.Now().Unix()
	title = strings.TrimSpace(title)
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	if title == `` {
		title = `未命名片段`
	}
	tags = h.memoryFragmentNormalizeTags(tags)
	contentText := h.memoryFragmentToPlainText(content)
	indexStatus := `success`
	indexVersion := 1
	changeDesc := ``

	if id > 0 {
		oldInfo, err := h.Client.QuickQuery(`tbl_memory_fragment`, `*`, map[string]any{
			`id`:         id,
			`is_deleted`: 0,
		}).One()
		if err != nil {
			return nil, err
		}
		if len(oldInfo) == 0 {
			return nil, errors.New(`片段不存在`)
		}
		oldTags, tagErr := h.memoryFragmentLoadTags(id)
		if tagErr != nil {
			return nil, tagErr
		}
		indexVersion = cast.ToInt(oldInfo[`index_version`]) + 1
		_, err = h.Client.QuickUpdate(`tbl_memory_fragment`, map[string]any{
			`id`: id,
		}, map[string]any{
			`title`:         title,
			`content`:       content,
			`content_text`:  contentText,
			`index_status`:  `pending`,
			`index_version`: indexVersion,
			`update_time`:   now,
		}).Exec()
		if err != nil {
			return nil, err
		}
		if err = h.memoryFragmentReplaceTags(id, tags, now); err != nil {
			return nil, err
		}
		changeDesc = h.memoryFragmentBuildChangeDesc(
			cast.ToString(oldInfo[`title`]),
			title,
			cast.ToString(oldInfo[`content`]),
			content,
			oldTags,
			tags,
		)
		if changeDesc != `` {
			_, err = h.Client.QuickCreate(`tbl_memory_fragment_history`, map[string]any{
				`fragment_id`: id,
				`title_old`:   cast.ToString(oldInfo[`title`]),
				`title_new`:   title,
				`content_old`: cast.ToString(oldInfo[`content`]),
				`content_new`: content,
				`tags_old`:    gstool.JsonEncode(oldTags),
				`tags_new`:    gstool.JsonEncode(tags),
				`change_desc`: changeDesc,
				`create_time`: now,
				`update_time`: now,
			}).Exec()
			if err != nil {
				return nil, err
			}
		}
	} else {
		newID, err := h.Client.QuickCreate(`tbl_memory_fragment`, map[string]any{
			`title`:         title,
			`content`:       content,
			`content_text`:  contentText,
			`is_deleted`:    0,
			`index_status`:  `pending`,
			`index_version`: indexVersion,
			`create_time`:   now,
			`update_time`:   now,
		}).Exec()
		if err != nil {
			return nil, err
		}
		id = cast.ToInt(newID)
		if err = h.memoryFragmentReplaceTags(id, tags, now); err != nil {
			return nil, err
		}
	}

	if err := h.memoryFragmentSyncSearchIndex(id, title, contentText, tags, now); err != nil {
		indexStatus = `failed`
	} else {
		indexStatus = `success`
	}
	_, _ = h.Client.QuickUpdate(`tbl_memory_fragment`, map[string]any{
		`id`: id,
	}, map[string]any{
		`index_status`: indexStatus,
		`update_time`:  now,
	}).Exec()
	return h.MemoryFragmentInfo(id)
}

// MemoryFragmentSoftDelete 软删除知识片段。
func (h *CSqlite) MemoryFragmentSoftDelete(id int) (int64, error) {
	now := time.Now().Unix()
	_, _ = h.Client.ExecBySql(`delete from tbl_memory_fragment_fts where fragment_id = ?`, id).Exec()
	return h.Client.QuickUpdate(`tbl_memory_fragment`, map[string]any{
		`id`:         id,
		`is_deleted`: 0,
	}, map[string]any{
		`is_deleted`:   1,
		`index_status`: `pending`,
		`update_time`:  now,
	}).Exec()
}

// MemoryFragmentRestore 从回收站恢复知识片段。
func (h *CSqlite) MemoryFragmentRestore(id int) (int64, error) {
	now := time.Now().Unix()
	affected, err := h.Client.QuickUpdate(`tbl_memory_fragment`, map[string]any{
		`id`:         id,
		`is_deleted`: 1,
	}, map[string]any{
		`is_deleted`:   0,
		`index_status`: `pending`,
		`update_time`:  now,
	}).Exec()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		return 0, errors.New(`片段不存在`)
	}
	info, infoErr := h.MemoryFragmentInfo(id)
	if infoErr != nil {
		return affected, infoErr
	}
	if syncErr := h.memoryFragmentSyncSearchIndex(
		id,
		cast.ToString(info[`title`]),
		h.memoryFragmentToPlainText(cast.ToString(info[`content`])),
		h.memoryFragmentNormalizeTags(cast.ToStringSlice(info[`tags`])),
		now,
	); syncErr != nil {
		return affected, syncErr
	}
	_, _ = h.Client.QuickUpdate(`tbl_memory_fragment`, map[string]any{
		`id`: id,
	}, map[string]any{
		`index_status`: `success`,
		`update_time`:  now,
	}).Exec()
	return affected, nil
}

// MemoryFragmentHardDelete 彻底删除回收站中的知识片段及关联数据。
func (h *CSqlite) MemoryFragmentHardDelete(id int) error {
	trashInfo, err := h.Client.QuickQuery(`tbl_memory_fragment`, `id`, map[string]any{
		`id`:         id,
		`is_deleted`: 1,
	}).One()
	if err != nil {
		return err
	}
	if len(trashInfo) == 0 {
		return errors.New(`片段不存在`)
	}
	if _, err = h.Client.QuickDelete(`tbl_memory_fragment_tag`, map[string]any{
		`fragment_id`: id,
	}).Exec(); err != nil {
		return err
	}
	if _, err = h.Client.QuickDelete(`tbl_memory_fragment_history`, map[string]any{
		`fragment_id`: id,
	}).Exec(); err != nil {
		return err
	}
	if _, err = h.Client.QuickDelete(`tbl_memory_fragment_fts`, map[string]any{
		`fragment_id`: id,
	}).Exec(); err != nil {
		return err
	}
	if _, err = h.Client.QuickDelete(`tbl_memory_fragment`, map[string]any{
		`id`:         id,
		`is_deleted`: 1,
	}).Exec(); err != nil {
		return err
	}
	return nil
}

// MemoryFragmentHistoryList 查询知识片段历史记录。
func (h *CSqlite) MemoryFragmentHistoryList(fragmentID int) ([]map[string]any, error) {
	list, err := h.Client.QuickQuery(`tbl_memory_fragment_history`, `*`, map[string]any{
		`fragment_id`: fragmentID,
	}).Order("id desc").All()
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i][`create_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(list[i][`create_time`]))
		tagsOld := make([]string, 0)
		tagsNew := make([]string, 0)
		_ = gstool.JsonDecode(cast.ToString(list[i][`tags_old`]), &tagsOld)
		_ = gstool.JsonDecode(cast.ToString(list[i][`tags_new`]), &tagsNew)
		list[i][`tags_old_list`] = tagsOld
		list[i][`tags_new_list`] = tagsNew
	}
	return list, nil
}

// memoryFragmentListByDeleted 按删除状态查询知识片段列表。
func (h *CSqlite) memoryFragmentListByDeleted(isDeleted int, limit int) ([]map[string]any, error) {
	sql := `
select
	f.*,
	ifnull(group_concat(t.tag_name, ','), '') as tags_text
from tbl_memory_fragment f
left join tbl_memory_fragment_tag t on t.fragment_id = f.id
where f.is_deleted = ?
group by f.id
order by f.update_time desc, f.id desc`
	if limit > 0 {
		sql += fmt.Sprintf(" limit %d", limit)
	}
	list, err := h.Client.QueryBySql(sql, isDeleted).All()
	if err != nil {
		return nil, err
	}
	h.memoryFragmentFillDisplayFields(list)
	return list, nil
}

// MemoryFragmentTagList 查询知识片段标签列表。
func (h *CSqlite) MemoryFragmentTagList() ([]map[string]any, error) {
	sql := `
select
	t.tag_name,
	count(1) as use_count
from tbl_memory_fragment_tag t
inner join tbl_memory_fragment f on f.id = t.fragment_id
where f.is_deleted = 0
group by t.tag_name
order by use_count desc, t.tag_name asc`
	return h.Client.QueryBySql(sql).All()
}

// MemoryFragmentSearch 搜索知识片段。
func (h *CSqlite) MemoryFragmentSearch(mode, query string, selectedTags []string, limit int) ([]map[string]any, error) {
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode != `keyword` {
		mode = `keyword`
	}
	query = h.memoryFragmentNormalizeSearchQuery(query)
	selectedTags = h.memoryFragmentNormalizeTags(selectedTags)
	sql := `
select
	f.*,
	ifnull(group_concat(distinct t.tag_name), '') as tags_text,
	ifnull(max(s.search_text), '') as search_text
from tbl_memory_fragment f
left join tbl_memory_fragment_tag t on t.fragment_id = f.id
left join tbl_memory_fragment_fts s on s.fragment_id = f.id
where f.is_deleted = 0
group by f.id`
	rows, err := h.Client.QueryBySql(sql).All()
	if err != nil {
		return nil, err
	}
	type searchRow struct {
		row   map[string]any
		score int
	}
	resultRows := make([]searchRow, 0)
	tokens := h.memoryFragmentSearchTokens(query)
	for _, row := range rows {
		tags := h.memoryFragmentSplitTagText(cast.ToString(row[`tags_text`]))
		if !h.memoryFragmentMatchSelectedTags(selectedTags, tags) {
			continue
		}
		searchText := strings.ToLower(cast.ToString(row[`search_text`]))
		match, score := h.memoryFragmentSearchScore(mode, query, tokens, searchText, tags, cast.ToString(row[`title`]))
		if query == `` && len(selectedTags) > 0 {
			match = true
			score = 1
		}
		if query == `` && len(selectedTags) == 0 {
			match = true
			score = 1
		}
		if !match {
			continue
		}
		rowCopy := row
		rowCopy[`tags`] = tags
		rowCopy[`create_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(row[`create_time`]))
		rowCopy[`update_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(row[`update_time`]))
		delete(rowCopy, `index_status`)
		delete(rowCopy, `index_version`)
		rowCopy[`score`] = score
		resultRows = append(resultRows, searchRow{
			row:   rowCopy,
			score: score,
		})
	}
	sort.SliceStable(resultRows, func(i, j int) bool {
		if resultRows[i].score != resultRows[j].score {
			return resultRows[i].score > resultRows[j].score
		}
		updateLeft := cast.ToInt64(resultRows[i].row[`update_time`])
		updateRight := cast.ToInt64(resultRows[j].row[`update_time`])
		if updateLeft != updateRight {
			return updateLeft > updateRight
		}
		return cast.ToInt(resultRows[i].row[`id`]) > cast.ToInt(resultRows[j].row[`id`])
	})
	if limit <= 0 {
		limit = 50
	}
	result := make([]map[string]any, 0)
	for index, item := range resultRows {
		if index >= limit {
			break
		}
		result = append(result, item.row)
	}
	return result, nil
}

// memoryFragmentNormalizeSearchQuery 规范化搜索文本，便于多关键词搜索。
func (h *CSqlite) memoryFragmentNormalizeSearchQuery(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(query)), ` `)
}

// memoryFragmentReplaceTags 重建片段标签。
func (h *CSqlite) memoryFragmentReplaceTags(fragmentID int, tags []string, now int64) error {
	if _, err := h.Client.ExecBySql(`delete from tbl_memory_fragment_tag where fragment_id = ?`, fragmentID).Exec(); err != nil {
		return err
	}
	for _, tag := range tags {
		if _, err := h.Client.QuickCreate(`tbl_memory_fragment_tag`, map[string]any{
			`fragment_id`: fragmentID,
			`tag_name`:    tag,
			`create_time`: now,
			`update_time`: now,
		}).Exec(); err != nil {
			return err
		}
	}
	return nil
}

// memoryFragmentLoadTags 查询片段标签。
func (h *CSqlite) memoryFragmentLoadTags(fragmentID int) ([]string, error) {
	list, err := h.Client.QuickQuery(`tbl_memory_fragment_tag`, `tag_name`, map[string]any{
		`fragment_id`: fragmentID,
	}).Order("tag_name asc").All()
	if err != nil {
		return nil, err
	}
	tags := make([]string, 0, len(list))
	for _, item := range list {
		tag := strings.TrimSpace(cast.ToString(item[`tag_name`]))
		if tag == `` {
			continue
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// memoryFragmentSyncSearchIndex 同步知识片段搜索索引。
func (h *CSqlite) memoryFragmentSyncSearchIndex(fragmentID int, title, contentText string, tags []string, now int64) error {
	tagText := strings.Join(tags, ` `)
	searchText := strings.TrimSpace(strings.ToLower(strings.Join([]string{title, contentText, tagText}, ` `)))
	if _, err := h.Client.ExecBySql(`delete from tbl_memory_fragment_fts where fragment_id = ?`, fragmentID).Exec(); err != nil {
		return err
	}
	if _, err := h.Client.QuickCreate(`tbl_memory_fragment_fts`, map[string]any{
		`fragment_id`:  fragmentID,
		`title`:        title,
		`content_text`: contentText,
		`tag_text`:     tagText,
		`search_text`:  searchText,
		`create_time`:  now,
		`update_time`:  now,
	}).Exec(); err != nil {
		return err
	}
	return nil
}

// memoryFragmentBuildChangeDesc 生成历史摘要。
func (h *CSqlite) memoryFragmentBuildChangeDesc(oldTitle, newTitle, oldContent, newContent string, oldTags, newTags []string) string {
	changeParts := make([]string, 0)
	if oldTitle != newTitle {
		changeParts = append(changeParts, `标题`)
	}
	if oldContent != newContent {
		changeParts = append(changeParts, `内容`)
	}
	if !h.memoryFragmentStringSliceEqual(oldTags, newTags) {
		changeParts = append(changeParts, `标签`)
	}
	if len(changeParts) == 0 {
		return ``
	}
	return strings.Join(changeParts, `、`) + `已更新`
}

// memoryFragmentFillDisplayFields 填充列表展示字段。
func (h *CSqlite) memoryFragmentFillDisplayFields(list []map[string]any) {
	for i := range list {
		tags := h.memoryFragmentSplitTagText(cast.ToString(list[i][`tags_text`]))
		list[i][`tags`] = tags
		list[i][`create_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(list[i][`create_time`]))
		list[i][`update_time_desc`] = h.memoryFragmentFormatTime(cast.ToInt64(list[i][`update_time`]))
		delete(list[i], `index_status`)
		delete(list[i], `index_version`)
	}
}

// memoryFragmentFormatTime 格式化时间。
func (h *CSqlite) memoryFragmentFormatTime(unixTime int64) string {
	if unixTime <= 0 {
		return ``
	}
	return gstool.TimeUnixToString(time.Unix(unixTime, 0), `Y-m-d H:i:s`)
}

// memoryFragmentNormalizeTags 规范化标签列表。
func (h *CSqlite) memoryFragmentNormalizeTags(tags []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.Trim(tag, ",，")
		if tag == `` {
			continue
		}
		lowerTag := strings.ToLower(tag)
		if seen[lowerTag] {
			continue
		}
		seen[lowerTag] = true
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}

// memoryFragmentSplitTagText 拆分标签文本。
func (h *CSqlite) memoryFragmentSplitTagText(tagText string) []string {
	if strings.TrimSpace(tagText) == `` {
		return []string{}
	}
	return h.memoryFragmentNormalizeTags(strings.Split(tagText, `,`))
}

// memoryFragmentToPlainText 提取 Markdown 纯文本。
func (h *CSqlite) memoryFragmentToPlainText(content string) string {
	text := strings.ReplaceAll(content, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	replacer := strings.NewReplacer(
		"```", " ",
		"`", " ",
		"#", " ",
		">", " ",
		"*", " ",
		"_", " ",
		"-", " ",
		"|", " ",
		"[", " ",
		"]", " ",
		"(", " ",
		")", " ",
	)
	text = replacer.Replace(text)
	linkReg := regexp.MustCompile(`https?://[^\s]+`)
	text = linkReg.ReplaceAllString(text, ` `)
	spaceReg := regexp.MustCompile(`\s+`)
	text = spaceReg.ReplaceAllString(text, ` `)
	return strings.TrimSpace(text)
}

// memoryFragmentStringSliceEqual 比较字符串切片是否一致。
func (h *CSqlite) memoryFragmentStringSliceEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

// memoryFragmentSearchTokens 生成搜索词列表。
func (h *CSqlite) memoryFragmentSearchTokens(query string) []string {
	if query == `` {
		return []string{}
	}
	query = h.memoryFragmentNormalizeSearchQuery(query)
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, token := range strings.Fields(query) {
		if token == `` || seen[token] {
			continue
		}
		seen[token] = true
		result = append(result, token)
	}
	return result
}

// memoryFragmentMatchSelectedTags 判断标签筛选是否命中。
func (h *CSqlite) memoryFragmentMatchSelectedTags(selectedTags, rowTags []string) bool {
	if len(selectedTags) == 0 {
		return true
	}
	tagMap := make(map[string]bool)
	for _, tag := range rowTags {
		tagMap[strings.ToLower(tag)] = true
	}
	for _, tag := range selectedTags {
		if !tagMap[strings.ToLower(tag)] {
			return false
		}
	}
	return true
}

// PromptChangeLogSave 记录提示词变更日志。
func (h *CSqlite) PromptChangeLogSave(configKey, configName, oldValue, newValue string) error {
	now := time.Now().Unix()
	_, err := h.Client.QuickCreate(`tbl_prompt_change_log`, map[string]any{
		`config_key`:  configKey,
		`config_name`: configName,
		`old_value`:   oldValue,
		`new_value`:   newValue,
		`create_time`: now,
		`update_time`: now,
	}).Exec()
	return err
}

// PromptChangeLogList 查询提示词变更日志，返回最近 limit 条记录。
func (h *CSqlite) PromptChangeLogList(limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 20
	}
	sql := `select * from tbl_prompt_change_log order by id desc limit ?`
	return h.Client.QueryBySql(sql, limit).All()
}

// memoryFragmentSearchScore 计算搜索得分。
func (h *CSqlite) memoryFragmentSearchScore(mode, query string, tokens []string, searchText string, tags []string, title string) (bool, int) {
	if query == `` {
		return true, 1
	}
	searchText = strings.ToLower(searchText)
	title = strings.ToLower(title)
	phraseMatched := strings.Contains(searchText, query)
	tokenMatchCount := 0
	allMatched := len(tokens) > 0
	titleScore := 0
	tagScore := 0
	for _, token := range tokens {
		if strings.Contains(searchText, token) {
			tokenMatchCount += 1
			if strings.Contains(title, token) {
				titleScore += 6
			}
			for _, tag := range tags {
				if strings.Contains(strings.ToLower(tag), token) {
					tagScore += 3
				}
			}
			continue
		}
		allMatched = false
	}
	if phraseMatched {
		if strings.Contains(title, query) {
			titleScore += 6
		}
		for _, tag := range tags {
			if strings.Contains(strings.ToLower(tag), query) {
				tagScore += 3
			}
		}
	}
	if mode != `keyword` {
		mode = `keyword`
	}
	if !allMatched && !phraseMatched {
		return false, 0
	}
	score := tokenMatchCount*10 + tagScore + titleScore
	if phraseMatched {
		score += 8
	}
	return true, score
}
