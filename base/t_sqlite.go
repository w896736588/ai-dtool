package base

import (
	"errors"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"time"
)

type TSqlite struct {
	Client *gsdb.GsSqlite
	Env    *Env
}

func (h *TSqlite) InitTable() {
	//TODO 初始化表机构和变更
}

func (h *TSqlite) Login(username, password string) (int, error) {
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

func (h *TSqlite) CreateUser(username, password string) (int, error) {
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

func (h *TSqlite) GetSaltPassword(password string) string {
	return gstool.Md5(password + h.Env.AppName)
}

func (h *TSqlite) CreateSsh(name, host, port string, userid, isPublic int) (int, error) {
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

func (h *TSqlite) CreateSshUser(sshid, userid int, username, password string) (int, error) {
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

func (h *TSqlite) CreateCode(name, path string, sshid, codeGroupId int) (int, error) {
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

func (h *TSqlite) GetSshConfigByUserSshId(userSshId int) (map[string]any, error) {
	sql := `select us.username,us.password,s.host,s.port from tbl_user_ssh us 
			left join tbl_ssh s on s.id = us.ssh_id 
			where us.id = ?`
	return h.Client.QueryBySql(sql, userSshId).One()
}

func (h *TSqlite) GetSshConfigByUserId(userId int) ([]map[string]any, error) {
	sql := `select us.username,us.password,s.host,s.port,us.id from tbl_user_ssh us 
			left join tbl_ssh s on s.id = us.ssh_id 
			where us.user_id = ?`
	return h.Client.QueryBySql(sql, userId).All()
}

func (h *TSqlite) GetAllSshConfig() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
}

func (h *TSqlite) GetSshConfig(sshId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_ssh`, `*`, map[string]any{
		`id`: sshId,
	}).One()
}

func (h *TSqlite) GetRedisConfig(redisId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_redis`, `*`, map[string]any{
		`id`: redisId,
	}).One()
}

func (h *TSqlite) GetAllRedisConfig() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_redis`, `*`, nil).All()
}

func (h *TSqlite) GetMysqlConfig(mysqlId any) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_mysql`, `*`, map[string]any{
		`id`: mysqlId,
	}).One()
}

func (h *TSqlite) StarAdd(id, name, key, value, _type any) (int64, error) {
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

func (h *TSqlite) StarDel(id any) (int64, error) {
	return h.Client.QuickDelete(`tbl_star`, map[string]any{
		`id`: id,
	}).Exec()
}

func (h *TSqlite) StarList(_type any) ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_star`, `*`, map[string]any{
		`type`: _type,
	}).All()
}
