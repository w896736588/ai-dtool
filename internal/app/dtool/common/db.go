package common

import (
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"errors"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

var DbMain = &CSqlite{}

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
