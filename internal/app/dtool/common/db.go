package common

import (
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"errors"
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

// ZcodeConfigGet 获取 zcode 配置（全局只有一条）。
func (h *CSqlite) ZcodeConfigGet() (map[string]any, error) {
	one, err := h.Client.QuickQuery(`tbl_zcode_config`, `*`, nil).Order(`id desc`).One()
	if err != nil {
		if DbRowMissing(err) {
			return nil, nil
		}
		return nil, err
	}
	return one, nil
}

// ZcodeConfigSave 保存 zcode 配置（upsert，始终只有一条记录）。
func (h *CSqlite) ZcodeConfigSave(zcodeDir string) (int64, error) {
	now := time.Now().Unix()
	one, _ := h.Client.QuickQuery(`tbl_zcode_config`, `*`, nil).Order(`id desc`).One()
	updateData := map[string]any{
		`zcode_dir`:  zcodeDir,
		`updated_at`: cast.ToString(now),
	}
	if one != nil && cast.ToInt(one[`id`]) > 0 {
		id := cast.ToInt64(one[`id`])
		_, err := h.Client.QuickUpdate(`tbl_zcode_config`, map[string]any{`id`: id}, updateData).Exec()
		return id, err
	}
	updateData[`created_at`] = cast.ToString(now)
	newID, err := h.Client.QuickCreate(`tbl_zcode_config`, updateData).Exec()
	return newID, err
}

// ZcodeConfigDelete 删除 zcode 配置。
func (h *CSqlite) ZcodeConfigDelete() error {
	_, err := h.Client.QuickDelete(`tbl_zcode_config`, nil).Exec()
	return err
}

// ZcodeProjectMappingReplace 替换 zcode 项目映射：先删除旧数据再批量插入。
func (h *CSqlite) ZcodeProjectMappingReplace(configID int64, items []ZcodeProjectMappingItem) error {
	if configID <= 0 {
		return nil
	}
	// 删除该 config 下的所有旧映射
	_, _ = h.Client.QuickDelete(`tbl_zcode_project_mapping`, map[string]any{
		`zcode_config_id`: configID,
	}).Exec()
	for _, item := range items {
		_, err := h.Client.QuickCreate(`tbl_zcode_project_mapping`, map[string]any{
			`zcode_config_id`: configID,
			`project_key`:     item.ProjectKey,
			`workspace_path`:  item.WorkspacePath,
			`settings_path`:   item.SettingsPath,
		}).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// ZcodeProjectMappingItem 项目映射条目。
type ZcodeProjectMappingItem struct {
	ProjectKey    string
	WorkspacePath string
	SettingsPath  string
}

// ZcodeProjectMappingList 获取所有项目映射。
func (h *CSqlite) ZcodeProjectMappingList() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_zcode_project_mapping`, `*`, nil).All()
}

// ZcodeProjectMappingGetByWorkspacePath 按工作目录精确匹配 settings 路径。
func (h *CSqlite) ZcodeProjectMappingGetByWorkspacePath(workspacePath string) (map[string]any, error) {
	one, err := h.Client.QuickQuery(`tbl_zcode_project_mapping`, `*`, map[string]any{
		`workspace_path`: workspacePath,
	}).One()
	if err != nil {
		if DbRowMissing(err) {
			return nil, nil
		}
		return nil, err
	}
	return one, nil
}

// PromptChangeLogList 查询提示词变更日志，返回最近 limit 条记录。
func (h *CSqlite) PromptChangeLogList(limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 20
	}
	sql := `select * from tbl_prompt_change_log order by id desc limit ?`
	return h.Client.QueryBySql(sql, limit).All()
}
