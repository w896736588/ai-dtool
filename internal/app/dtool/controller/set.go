package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstask"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	ini "gopkg.in/ini.v1"
)

// SetSshList ssh列表
func SetSshList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	// 是否检查连接状态，1检查，0不检查，默认0
	isCheckConnection := cast.ToInt(dataMap[`is_check_connection`])

	all, allErr := common.DbMain.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	allSsh := map[int]map[string]any{}

	// 只有在需要检查连接状态时才执行连接测试
	if isCheckConnection == 1 {
		//返回连接状态
		task := gstask.NewTask()
		for _, sshValue := range all {
			allSsh[cast.ToInt(sshValue[`id`])] = sshValue
			callBack := gstask.CallbackFunc{
				Func: func() *gstask.Result {
					return testSshConn(sshValue)
				},
				Timeout: 3 * time.Second,
				Id:      cast.ToString(sshValue[`id`]),
			}
			task.Add(callBack)
		}
		resultList := task.RunAll()
		//填充链接状态
		for _, result := range resultList {
			for sshId, _ := range allSsh {
				if sshId == cast.ToInt(result.Id) {
					if result.Err != nil {
						allSsh[sshId][`status`] = result.Err.Error()
					} else {
						allSsh[sshId][`status`] = `success`
					}
				}
			}
		}
	} else {
		// 不检查连接状态，直接填充数据
		for _, sshValue := range all {
			allSsh[cast.ToInt(sshValue[`id`])] = sshValue
		}
	}

	returnList := make([]map[string]any, 0)
	for _, sshValue := range allSsh {
		returnList = append(returnList, sshValue)
	}
	gsgin.GinResponseSuccess(c, ``, returnList)
}

func SetSshAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `host`, `port`, `username`, `password`, `home`})
	if cast.ToString(updateData[`db_type`]) == `` {
		updateData[`db_type`] = DbTypeMysql
	}
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickCreate(`tbl_ssh`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_ssh`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetSshDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_ssh`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGitList(c *gin.Context) {
	allGit, allGitErr := common.DbMain.Client.QuickQuery(`tbl_git`, `*`, nil).All()
	if allGitErr != nil {
		gsgin.GinResponseError(c, allGitErr.Error(), nil)
		return
	}
	allGitGroup, allGitGroupErr := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeGit,
	}).All()
	if allGitGroupErr != nil {
		gsgin.GinResponseError(c, allGitGroupErr.Error(), nil)
		return
	}
	allSsh, allSshErr := common.DbMain.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	if allSshErr != nil {
		gsgin.GinResponseError(c, allSshErr.Error(), nil)
		return
	}
	for gitKey, gitValue := range allGit {
		allGit[gitKey][`ssh_name`] = ``
		allGit[gitKey][`git_group_name`] = ``
		gitGroupId := cast.ToInt(gitValue[`git_group_id`])
		if gitGroupId != 0 {
			for _, gitGroupValue := range allGitGroup {
				if cast.ToInt(gitGroupValue[`id`]) == gitGroupId {
					allGit[gitKey][`git_group_name`] = gitGroupValue[`name`]
				}
			}
		}
		gitSshId := cast.ToInt(gitValue[`ssh_id`])
		if gitSshId != 0 {
			for _, sshValue := range allSsh {
				if cast.ToInt(sshValue[`id`]) == gitSshId {
					allGit[gitKey][`ssh_name`] = sshValue[`name`]
				}
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, allGit)
}

func SetGitAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `ssh_id`, `code_path`, `git_group_id`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickCreate(`tbl_git`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_git`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGitDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_git`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGitGroupList(c *gin.Context) {
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeGit,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

func SetGitGroupAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		updateData[`type`] = define.GroupTypeGit
		_, _ = common.DbMain.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_group`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGitGroupDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGitQuickList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToString(dataMap[`dir`]) == `` {
		gsgin.GinResponseError(c, `dir不能为空`, nil)
		return
	}
	sshList, sshListErr := common.DbMain.GetAllSshConfig()
	if sshListErr != nil {
		gsgin.GinResponseError(c, sshListErr.Error(), nil)
		return
	}
	searchDir := cast.ToString(dataMap[`dir`])
	existMap := make(map[string]string)
	gitDirList := make([]map[string]any, 0)
	for _, sshConfig := range sshList {
		findDirList := business.FindCode(sshConfig, searchDir)
		for _, findDir := range findDirList {
			if strings.Index(findDir, searchDir) != 0 {
				continue
			}
			if existMap[findDir] == `EXIST` {
				continue
			}
			existMap[findDir] = `EXIST`
			//查找group_id
			gitInfo, _ := common.DbMain.Client.QuickQuery(`tbl_git`, `git_group_id`, map[string]any{
				`code_path`: findDir,
			}).One()
			gitDirList = append(gitDirList, map[string]any{
				`code_path`: findDir,
				`name`: gstool.SReplaces(findDir, map[string]string{
					searchDir: ``,
				}),
				`ssh_id`:       cast.ToString(sshConfig[`id`]),
				`ssh_name`:     cast.ToString(sshConfig[`name`]),
				`git_group_id`: cast.ToString(gitInfo[`git_group_id`]),
			})
		}
	}
	gsgin.GinResponseSuccess(c, ``, gitDirList)
}

func SetSupervisorctlList(c *gin.Context) {
	allSupervisor, allSupervisorErr := common.DbMain.Client.QuickQuery(`tbl_supervisor`, `*`, nil).All()
	if allSupervisorErr != nil {
		gsgin.GinResponseError(c, allSupervisorErr.Error(), nil)
		return
	}
	allSsh, allSshErr := common.DbMain.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	if allSshErr != nil {
		gsgin.GinResponseError(c, allSshErr.Error(), nil)
		return
	}
	for gitKey, gitValue := range allSupervisor {
		allSupervisor[gitKey][`ssh_name`] = ``
		gitSshId := cast.ToInt(gitValue[`ssh_id`])
		if gitSshId != 0 {
			for _, sshValue := range allSsh {
				if cast.ToInt(sshValue[`id`]) == gitSshId {
					allSupervisor[gitKey][`ssh_name`] = sshValue[`name`]
				}
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, allSupervisor)
}

func SetSupervisorAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `ssh_id`, `docker_name`, `config_dir`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, createErr := common.DbMain.Client.QuickCreate(`tbl_supervisor`, updateData).Exec()
		if createErr != nil {
			gstool.FmtPrintlnLogTime(`创建失败 %s`, createErr.Error())
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_supervisor`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetSupervisorDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_supervisor`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetRedisList redis列表
func SetRedisList(c *gin.Context) {
	allRedis, allErr := common.DbMain.Client.QuickQuery(`tbl_redis`, `*`, nil).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	allSsh, allSshErr := common.DbMain.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	if allSshErr != nil {
		gsgin.GinResponseError(c, allSshErr.Error(), nil)
		return
	}
	//返回连接状态
	task := gstask.NewTask()
	for gitKey, gitValue := range allRedis {
		allRedis[gitKey][`ssh_name`] = ``
		gitSshId := cast.ToInt(gitValue[`ssh_id`])
		if gitSshId != 0 {
			for _, sshValue := range allSsh {
				if cast.ToInt(sshValue[`id`]) == gitSshId {
					allRedis[gitKey][`ssh_name`] = sshValue[`name`]
				}
			}
		}
		callBack := gstask.CallbackFunc{
			Func: func() *gstask.Result {
				return testRedisConn(gitValue)
			},
			Timeout: 3 * time.Second,
			Id:      cast.ToString(gitValue[`id`]),
		}
		task.Add(callBack)
	}
	resultList := task.RunAll()
	//填充链接状态
	for _, result := range resultList {
		for redisKey, redisValue := range allRedis {
			if cast.ToInt(redisValue[`id`]) == cast.ToInt(result.Id) {
				if result.Err != nil {
					allRedis[redisKey][`status`] = result.Err.Error()
				} else {
					allRedis[redisKey][`status`] = `success`
				}
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, allRedis)
}

func testRedisConn(redisConfig map[string]any) *gstask.Result {
	gsRedis := &gsdb.GsRedis{
		RedisConfig: &gsdb.RedisConfig{
			Name:              cast.ToString(redisConfig[`name`]),
			Host:              cast.ToString(redisConfig[`host`]),
			Port:              cast.ToInt64(redisConfig[`port`]),
			Password:          cast.ToString(redisConfig[`password`]),
			MaxOpenConns:      1,
			MaxIdleConns:      1,
			Default:           0,
			Username:          cast.ToString(redisConfig[`username`]),
			MaxLifetimeSecond: 3600,
		},
	}
	if cast.ToInt(redisConfig[`ssh_id`]) != 0 {
		sshConfig, sshConfigErr := common.DbMain.GetSshConfig(redisConfig[`ssh_id`])
		if sshConfigErr != nil {
			return &gstask.Result{
				Err:    gstool.Error(`获取ssh配置失败 %s`, sshConfigErr.Error()),
				Result: redisConfig[`id`],
			}
		}
		gsRedis.SshBridge = gsssh.NewSshBridge(gsssh.NewSsh(&gsssh.SshConfig{
			Name:     cast.ToString(sshConfig[`name`]),
			Host:     cast.ToString(sshConfig[`host`]),
			Port:     cast.ToString(sshConfig[`port`]),
			UserName: cast.ToString(sshConfig[`username`]),
			Password: cast.ToString(sshConfig[`password`]),
		}))
	}
	connErr := gsRedis.CreateConn()
	if connErr != nil {
		return &gstask.Result{
			Err:    connErr,
			Result: redisConfig[`id`],
		}
	}
	_ = gsRedis.Client.Close()
	gsRedis = nil
	return &gstask.Result{
		Err:    nil,
		Result: redisConfig[`id`],
	}
}

func testSshConn(sshConfig map[string]any) *gstask.Result {
	ssh := gsssh.NewSsh(&gsssh.SshConfig{
		Name:     cast.ToString(sshConfig[`name`]),
		Host:     cast.ToString(sshConfig[`host`]),
		Port:     cast.ToString(sshConfig[`port`]),
		Password: cast.ToString(sshConfig[`password`]),
		UserName: cast.ToString(sshConfig[`username`]),
	})
	connErr := ssh.ConnectAuthPassword()
	if connErr != nil {
		return &gstask.Result{
			Err:    connErr,
			Result: sshConfig[`id`],
		}
	}
	ssh.Close()
	return &gstask.Result{
		Err:    nil,
		Result: sshConfig[`id`],
	}
}

func testDbConn(dbConfig map[string]any) *gstask.Result {
	dbType := cast.ToString(dbConfig[`db_type`])
	if dbType == `` {
		dbType = DbTypeMysql
	}
	sshBridge := func() *gsssh.SshBridge {
		if cast.ToInt(dbConfig[`ssh_id`]) == 0 {
			return nil
		}
		sshConfig, sshConfigErr := common.DbMain.GetSshConfig(dbConfig[`ssh_id`])
		if sshConfigErr != nil {
			return nil
		}
		return gsssh.NewSshBridge(gsssh.NewSsh(&gsssh.SshConfig{
			Name:     cast.ToString(sshConfig[`name`]),
			Host:     cast.ToString(sshConfig[`host`]),
			Port:     cast.ToString(sshConfig[`port`]),
			UserName: cast.ToString(sshConfig[`username`]),
			Password: cast.ToString(sshConfig[`password`]),
		}))
	}()
	var connErr error
	if dbType == DbTypePgsql {
		gsPgsql := &gsdb.GsPgsql{
			PgsqlConfig: &gsdb.PgsqlConfig{
				Name:     cast.ToString(dbConfig[`name`]),
				Host:     cast.ToString(dbConfig[`host`]),
				Port:     cast.ToInt64(dbConfig[`port`]),
				Password: cast.ToString(dbConfig[`password`]),
				Username: cast.ToString(dbConfig[`username`]),
				Dbname:   cast.ToString(dbConfig[`dbname`]),
			},
		}
		gsPgsql.SshBridge = sshBridge
		connErr = gsPgsql.CreateConn()
	} else {
		gsMysql := &gsdb.GsMysql{
			MysqlConfig: &gsdb.MysqlConfig{
				Name:     cast.ToString(dbConfig[`name`]),
				Host:     cast.ToString(dbConfig[`host`]),
				Port:     cast.ToInt64(dbConfig[`port`]),
				Password: cast.ToString(dbConfig[`password`]),
				Username: cast.ToString(dbConfig[`username`]),
				Dbname:   cast.ToString(dbConfig[`dbname`]),
			},
		}
		gsMysql.SshBridge = sshBridge
		connErr = gsMysql.CreateConn()
	}
	if connErr != nil {
		return &gstask.Result{
			Err:    connErr,
			Result: dbConfig[`id`],
		}
	}
	return &gstask.Result{
		Err:    nil,
		Result: dbConfig[`id`],
	}
}

func SetRedisAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `host`, `port`, `username`, `password`, `ssh_id`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickCreate(`tbl_redis`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_redis`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetRedisDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_redis`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetMysqlList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	// 是否检查连接状态，1检查，0或未传不检查
	isCheckConnection := cast.ToInt(dataMap[`check_status`])

	allMysql, allErr := common.DbMain.Client.QuickQuery(`tbl_mysql`, `*`, nil).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	allSsh, allSshErr := common.DbMain.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	if allSshErr != nil {
		gsgin.GinResponseError(c, allSshErr.Error(), nil)
		return
	}
	for mysqlKey, mysqlValue := range allMysql {
		allMysql[mysqlKey][`ssh_name`] = ``
		gitSshId := cast.ToInt(mysqlValue[`ssh_id`])
		if gitSshId != 0 {
			for _, sshValue := range allSsh {
				if cast.ToInt(sshValue[`id`]) == gitSshId {
					allMysql[mysqlKey][`ssh_name`] = sshValue[`name`]
				}
			}
		}
	}

	if isCheckConnection == 1 {
		task := gstask.NewTask()
		for _, mysqlValue := range allMysql {
			callBack := gstask.CallbackFunc{
				Func: func() *gstask.Result {
					return testDbConn(mysqlValue)
				},
				Timeout: 2 * time.Second,
				Id:      cast.ToString(mysqlValue[`id`]),
			}
			task.Add(callBack)
		}
		resultList := task.RunAll()
		for _, result := range resultList {
			for mysqlKey, mysqlValue := range allMysql {
				if cast.ToInt(mysqlValue[`id`]) == cast.ToInt(result.Id) {
					if result.Err != nil {
						allMysql[mysqlKey][`status`] = result.Err.Error()
					} else {
						allMysql[mysqlKey][`status`] = `success`
					}
				}
			}
		}
	}

	gsgin.GinResponseSuccess(c, ``, allMysql)
}

func SetMysqlAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `host`, `port`, `username`, `dbname`, `password`, `ssh_id`, `db_type`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, createErr := common.DbMain.Client.QuickCreate(`tbl_mysql`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, createErr.Error(), nil)
			return
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, updateErr := common.DbMain.Client.QuickUpdate(`tbl_mysql`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		if updateErr != nil {
			gsgin.GinResponseError(c, updateErr.Error(), nil)
			return
		}
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetMysqlDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_mysql`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetVariableGroupList(c *gin.Context) {
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeVariable,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

func SetVariableGroupAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		updateData[`type`] = define.GroupTypeVariable
		_, _ = common.DbMain.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_group`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetVariableGroupDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetCmdGroupList(c *gin.Context) {
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeCmd,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

func SetCmdGroupAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		updateData[`type`] = define.GroupTypeCmd
		_, _ = common.DbMain.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_group`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetCmdGroupDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetSmartLinkGroupList(c *gin.Context) {
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeSmartLink,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

func SetSmartLinkGroupAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		updateData[`type`] = define.GroupTypeSmartLink
		_, _ = common.DbMain.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_group`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetSmartLinkGroupDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetDockerComposeList(c *gin.Context) {
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`status`: 1,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	allSsh, allSshErr := common.DbMain.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	if allSshErr != nil {
		gsgin.GinResponseError(c, allSshErr.Error(), nil)
		return
	}
	for key, value := range all {
		all[key][`ssh_name`] = ``
		gitSshId := cast.ToInt(value[`ssh_id`])
		if gitSshId != 0 {
			for _, sshValue := range allSsh {
				if cast.ToInt(sshValue[`id`]) == gitSshId {
					all[key][`ssh_name`] = sshValue[`name`]
				}
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

func SetDockerComposeAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `compose_yml_path`, `env_file`, `ssh_id`, `docker_cmd`, `default_service`, `upload_exes`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickCreate(`tbl_docker_compose`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_docker_compose`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetDockerComposeDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		ret, err := common.DbMain.Client.QuickUpdate(`tbl_docker_compose`, map[string]any{
			`id`: dataMap[`id`],
		}, map[string]any{
			`status`: 0,
		}).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		} else {
			if ret == 0 {
				gsgin.GinResponseError(c, `删除失败`, nil)
				return
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGitlabTokenList(c *gin.Context) {
	allGit, allGitErr := common.DbMain.Client.QuickQuery(`tbl_gitlab_token`, `*`, nil).All()
	if allGitErr != nil {
		gsgin.GinResponseError(c, allGitErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, allGit)
}

func SetGitlabTokenAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `url`, `access_token`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickCreate(`tbl_gitlab_token`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_gitlab_token`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGitlabTokenDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_gitlab_token`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGlobalList(c *gin.Context) {
	allGit, allGitErr := common.DbMain.Client.QuickQuery(`tbl_global`, `*`, nil).All()
	if allGitErr != nil {
		gsgin.GinResponseError(c, allGitErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, allGit)
}

func SetGlobalAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`key`, `value`, `name`, `desc`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickCreate(`tbl_global`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_global`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGlobalDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_global`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetMemoryConfigGet 返回记忆配置页面数据 / return memory settings page data.
func SetMemoryConfigGet(c *gin.Context) {
	mainDBConfig := business.ReadMainDBConfig()
	memoryConfig := business.ReadMemoryConfigFromINI()
	arrangePrompt, err := memoryConfigValue(define.MemoryConfigArrangePrompt)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	arrangeModelID, err := memoryConfigValue(define.MemoryConfigArrangeModelID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	aiSearchModelID, err := memoryConfigValue(define.MemoryConfigAiSearchModelID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`db_dir`:                            mainDBConfig.Dir,
		`db_name`:                           mainDBConfig.DBName,
		`db_configured`:                     mainDBConfig.Dir != `` && mainDBConfig.DBName != ``,
		`db_is_git_repo`:                    mainDBConfig.GitRepoEnabled,
		`db_auto_push_delay_minutes`:        business.ReadMainDBAutoSyncConfig().AutoSyncMinutes,
		`log_db_path`:                       component.EnvClient.LogDbConfig.DbPath,
		`memory_dir`:                        memoryConfig.Dir,
		`memory_db_configured`:              memoryConfig.Dir != ``,
		`memory_db_is_git_repo`:             memoryConfig.GitRepoEnabled,
		`memory_db_auto_push_delay_minutes`: memoryConfig.AutoPushDelayMinutes,
		`memory_config_file`:                memoryConfigFilePath(),
		`memory_arrange_prompt`:             arrangePrompt,
		`memory_arrange_model_id`:           cast.ToInt(arrangeModelID),
		`memory_ai_search_model_id`:         cast.ToInt(aiSearchModelID),
		`safe_password`:                     component.ConfigViper.GetString(`safe.password`),
		`run_mode`:                          component.EnvClient.SmartLinkConfig.RunMode,
		`client_version`:                    component.EnvClient.SmartLinkConfig.ClientVersion,
	})
}

// SetMemoryConfigSave 仅保存 AI 相关配置 / save AI-related memory settings only.
func SetMemoryConfigSave(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	memoryArrangePrompt := strings.TrimSpace(cast.ToString(dataMap[`memory_arrange_prompt`]))
	if memoryArrangePrompt == `` {
		memoryArrangePrompt = defaultMemoryArrangePrompt()
	}
	memoryArrangeModelID := cast.ToInt(dataMap[`memory_arrange_model_id`])
	if memoryArrangeModelID > 0 {
		modelInfo, err := common.DbMain.AiModelInfo(memoryArrangeModelID)
		if err != nil {
			gsgin.GinResponseError(c, `AI 模型不存在`, nil)
			return
		}
		// 记忆整理仅允许使用 LLM 模型 / only LLM models are allowed for memory arrangement.
		if strings.ToLower(cast.ToString(modelInfo[`model_type`])) != `llm` {
			gsgin.GinResponseError(c, `记忆整理仅支持选择 LLM 模型`, nil)
			return
		}
	}
	if err := common.DbMain.MemoryConfigSave(`记忆整理提示词`, define.MemoryConfigArrangePrompt, memoryArrangePrompt, `知识片段 AI 整理提示词`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if err := common.DbMain.MemoryConfigSave(`记忆整理模型`, define.MemoryConfigArrangeModelID, cast.ToString(memoryArrangeModelID), `知识片段 AI 整理所用模型 id`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	memoryAiSearchModelID := cast.ToInt(dataMap[`memory_ai_search_model_id`])
	if memoryAiSearchModelID > 0 {
		modelInfo, err := common.DbMain.AiModelInfo(memoryAiSearchModelID)
		if err != nil {
			gsgin.GinResponseError(c, `AI 搜索模型不存在`, nil)
			return
		}
		if strings.ToLower(cast.ToString(modelInfo[`model_type`])) != `llm` {
			gsgin.GinResponseError(c, `AI 搜索仅支持选择 LLM 模型`, nil)
			return
		}
	}
	if err := common.DbMain.MemoryConfigSave(`AI搜索模型`, define.MemoryConfigAiSearchModelID, cast.ToString(memoryAiSearchModelID), `知识片段 AI 智能搜索所用模型 id`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetRuntimeConfigSave 保存可编辑的 ini 配置并重新加载运行时配置。 // Save editable ini config values and reload runtime config.
func SetRuntimeConfigSave(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	configFile := memoryConfigFilePath()
	if strings.TrimSpace(configFile) == `` {
		gsgin.GinResponseError(c, `未找到配置文件路径`, nil)
		return
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{
		Loose: true,
	}, configFile)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 保存前读取当前密码，用于判断密码是否修改
	oldSafePassword := component.ConfigViper.GetString(`safe.password`)

	baseSection := cfg.Section(`base`)
	safeSection := cfg.Section(`safe`)

	setIniKey(baseSection, `dbPath`, strings.TrimSpace(cast.ToString(dataMap[`db_path`])))
	setIniKey(baseSection, `dbFileName`, strings.TrimSpace(cast.ToString(dataMap[`db_file_name`])))
	setIniKey(baseSection, `dbIsGitRepo`, cast.ToString(cast.ToBool(dataMap[`db_is_git_repo`])))
	setIniKey(baseSection, `logDbPath`, strings.TrimSpace(cast.ToString(dataMap[`log_db_path`])))
	setIniKey(baseSection, `memoryDbPath`, strings.TrimSpace(cast.ToString(dataMap[`memory_db_path`])))
	setIniKey(baseSection, `memoryDbIsGitRepo`, cast.ToString(cast.ToBool(dataMap[`memory_db_is_git_repo`])))
	setIniKey(baseSection, `memoryDbAutoPushDelayMinutes`, cast.ToString(cast.ToInt(dataMap[`memory_db_auto_push_delay_minutes`])))

	// 保存 safe 配置
	newSafePassword := strings.TrimSpace(cast.ToString(dataMap[`safe_password`]))
	setIniKey(safeSection, `password`, newSafePassword)

	// 判断密码是否修改
	safeChanged := oldSafePassword != newSafePassword

	if err = cfg.SaveTo(configFile); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	if component.ConfigViper != nil {
		// 保存后重新读取整个 ini，确保其他未编辑配置也保持最新。 // Re-read the whole ini after save so all config values stay in sync.
		if readErr := component.ConfigViper.ReadInConfig(); readErr != nil {
			gsgin.GinResponseError(c, readErr.Error(), nil)
			return
		}
	}
	business.ReloadEditableRuntimeConfig()

	// 如果密码修改了，需要重新登录
	needRelogin := safeChanged && newSafePassword != ``

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`config_file`:  configFile,
		`reloaded`:     true,
		`need_restart`: true,
		`safe_changed`: safeChanged,
		`need_relogin`: needRelogin,
	})
}

// SetRuntimeConfigItemSave 保存单个运行时配置项（用于独立编辑保存）。 // Save a single runtime config item for independent editing.
func SetRuntimeConfigItemSave(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	configKey := strings.TrimSpace(cast.ToString(dataMap[`key`]))
	configValue := dataMap[`value`]
	sectionName := strings.TrimSpace(cast.ToString(dataMap[`section`]))

	if configKey == `` || sectionName == `` {
		gsgin.GinResponseError(c, `配置项 key 和 section 不能为空`, nil)
		return
	}

	configFile := memoryConfigFilePath()
	if strings.TrimSpace(configFile) == `` {
		gsgin.GinResponseError(c, `未找到配置文件路径`, nil)
		return
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{
		Loose: true,
	}, configFile)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	section := cfg.Section(sectionName)

	// 根据 key 处理不同类型的配置项
	needRestart := false
	switch configKey {
	case `run_mode`:
		value := strings.TrimSpace(cast.ToString(configValue))
		if value != string(define.SmartLinkRunModeServer) && value != string(define.SmartLinkRunModeLocalClient) {
			gsgin.GinResponseError(c, `run_mode 必须是 server 或 local_client`, nil)
			return
		}
		setIniKey(section, configKey, value)
		// 更新内存中的配置
		component.EnvClient.SmartLinkConfig.RunMode = define.SmartLinkRunMode(value)
		needRestart = false
	case `client_version`:
		value := strings.TrimSpace(cast.ToString(configValue))
		setIniKey(section, configKey, value)
		component.EnvClient.SmartLinkConfig.ClientVersion = value
		needRestart = false
	case `db_path`:
		setIniKey(section, configKey, strings.TrimSpace(cast.ToString(configValue)))
		needRestart = false
	case `dbFileName`:
		setIniKey(section, configKey, strings.TrimSpace(cast.ToString(configValue)))
		needRestart = false
	case `logDbPath`:
		setIniKey(section, configKey, strings.TrimSpace(cast.ToString(configValue)))
		needRestart = false
	case `memoryDbPath`:
		setIniKey(section, configKey, strings.TrimSpace(cast.ToString(configValue)))
		needRestart = false
	case `db_is_git_repo`:
		setIniKey(section, configKey, cast.ToString(cast.ToBool(configValue)))
		needRestart = false
	case `memoryDbIsGitRepo`:
		setIniKey(section, configKey, cast.ToString(cast.ToBool(configValue)))
		needRestart = false
	case `dbAutoPushDelayMinutes`:
		setIniKey(section, configKey, cast.ToString(cast.ToInt(configValue)))
		needRestart = false
	case `memoryDbAutoPushDelayMinutes`:
		setIniKey(section, configKey, cast.ToString(cast.ToInt(configValue)))
		needRestart = false
	case `password`:
		oldSafePassword := component.ConfigViper.GetString(`safe.password`)
		newSafePassword := strings.TrimSpace(cast.ToString(configValue))
		setIniKey(section, configKey, newSafePassword)
		needRestart = false
		// 如果密码修改了，需要重新登录
		if oldSafePassword != newSafePassword && newSafePassword != `` {
			if err = cfg.SaveTo(configFile); err != nil {
				gsgin.GinResponseError(c, err.Error(), nil)
				return
			}
			if component.ConfigViper != nil {
				_ = component.ConfigViper.ReadInConfig()
			}
			business.ReloadEditableRuntimeConfig()
			gsgin.GinResponseSuccess(c, ``, map[string]any{
				`config_file`:  configFile,
				`reloaded`:     true,
				`need_restart`: false,
				`need_relogin`: true,
			})
			return
		}
	default:
		// 通用字符串配置
		setIniKey(section, configKey, strings.TrimSpace(cast.ToString(configValue)))
	}

	if err = cfg.SaveTo(configFile); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	if component.ConfigViper != nil {
		_ = component.ConfigViper.ReadInConfig()
	}
	business.ReloadEditableRuntimeConfig()

	// 热重载分发：根据配置项 key 调用对应热重载函数
	var hotReloadErr error
	switch configKey {
	case `db_path`, `dbFileName`:
		hotReloadErr = business.HotReloadMainDB(configKey)
	case `logDbPath`:
		hotReloadErr = business.HotReloadLogDB()
	case `memoryDbPath`, `memoryDbIsGitRepo`:
		hotReloadErr = business.HotReloadMemoryDB()
	case `db_is_git_repo`:
		hotReloadErr = business.HotReloadDBGitFlag()
	case `dbAutoPushDelayMinutes`:
		hotReloadErr = business.HotReloadAutoSyncDelay()
	case `memoryDbAutoPushDelayMinutes`:
		hotReloadErr = business.HotReloadMemoryAutoSyncDelay()
	}

	if hotReloadErr != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`配置已保存但热重载失败: %s`, hotReloadErr.Error()), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`config_file`:  configFile,
		`reloaded`:     true,
		`need_restart`: needRestart,
	})
}

const runtimeDatabaseSyncTargetMain = `main`
const runtimeDatabaseSyncTargetMemory = `memory`

// SetRuntimeDatabaseGitSync 手动触发主库或记忆库的 git commit push。 // Manually trigger git commit and push for the main or memory database.
func SetRuntimeDatabaseGitSync(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	target := strings.TrimSpace(cast.ToString(dataMap[`target`]))
	// target 只允许主库或记忆库两种同步入口，避免误触发其他路径。 // Only allow main or memory targets so the manual sync route stays explicit.
	switch target {
	case runtimeDatabaseSyncTargetMain:
		config := business.ReadMainDBConfig()
		if !config.GitRepoEnabled {
			gsgin.GinResponseError(c, `主库未开启 Git 同步`, nil)
			return
		}
		config.IsGitRepo = true
		changed, err := business.SyncMainDBFile(config, business.NewMemoryGit())
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, ``, map[string]any{
			`target`:  target,
			`changed`: changed,
		})
		return
	case runtimeDatabaseSyncTargetMemory:
		config := business.ReadMemoryConfigFromINI()
		if !config.GitRepoEnabled {
			gsgin.GinResponseError(c, `记忆库未开启 Git 同步`, nil)
			return
		}
		config.IsGitRepo = true
		changed, err := business.SyncMemoryDBFile(config, business.NewMemoryGit())
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		gsgin.GinResponseSuccess(c, ``, map[string]any{
			`target`:  target,
			`changed`: changed,
		})
		return
	default:
		gsgin.GinResponseError(c, `target 参数无效`, nil)
		return
	}
}

func setIniKey(section *ini.Section, key, value string) {
	if section == nil {
		return
	}
	section.Key(key).SetValue(value)
}

// memoryConfigFilePath 返回当前运行中的 ini 配置文件路径 / return active ini config file path.
func memoryConfigFilePath() string {
	if component.EnvClient == nil {
		return ``
	}
	configFileName := component.EnvClient.ConfigFile
	// 仅在未携带扩展名时补 `.ini` / append `.ini` only when extension is missing.
	if filepath.Ext(configFileName) == `` {
		configFileName += `.ini`
	}
	return filepath.Join(component.EnvClient.ConfigPath, configFileName)
}

func homeTaskConfigValue(key string) (string, error) {
	value, err := common.DbMain.HomeTaskConfigValue(key)
	if err != nil {
		if common.DbRowMissing(err) {
			return ``, nil
		}
		return ``, err
	}
	return value, nil
}

func memoryConfigValue(key string) (string, error) {
	value, err := common.DbMain.MemoryConfigValue(key)
	if err != nil {
		if common.DbRowMissing(err) {
			return ``, nil
		}
		return ``, err
	}
	return value, nil
}

func SetAccountList(c *gin.Context) {
	allAccount, allAccountErr := common.DbMain.Client.QuickQuery(`tbl_account`, `*`, nil).All()
	if allAccountErr != nil {
		gsgin.GinResponseError(c, allAccountErr.Error(), nil)
		return
	}
	allAccountGroup, allAccountGroupErr := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeAccount,
	}).All()
	if allAccountGroupErr != nil {
		gsgin.GinResponseError(c, allAccountGroupErr.Error(), nil)
		return
	}
	for AccountKey, AccountValue := range allAccount {
		allAccount[AccountKey][`account_group_name`] = ``
		AccountGroupId := cast.ToInt(AccountValue[`account_group_id`])
		if AccountGroupId != 0 {
			for _, AccountGroupValue := range allAccountGroup {
				if cast.ToInt(AccountGroupValue[`id`]) == AccountGroupId {
					allAccount[AccountKey][`account_group_name`] = AccountGroupValue[`name`]
				}
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, allAccount)
}

func SetAccountAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`username`, `password`, `account_group_id`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickCreate(`tbl_account`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_account`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetAccountDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_account`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetAccountGroupList(c *gin.Context) {
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeAccount,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

func SetAccountGroupAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		updateData[`type`] = define.GroupTypeAccount
		_, _ = common.DbMain.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_group`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetCronConfigGet 返回定时任务配置。 // Return scheduled task settings.
func SetCronConfigGet(c *gin.Context) {
	one, err := common.DbMain.CronTaskByType(define.CronTaskTypeDailyReport)
	if err != nil && !common.DbRowMissing(err) {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`cron_daily_report_enabled`: cast.ToInt(one[`enabled`]),
		`cron_daily_report_time`:    strings.TrimSpace(cast.ToString(one[`trigger_time`])),
	})
}

// SetCronConfigSave 保存定时任务配置并热重载调度器。 // Save scheduled task settings and hot-reload the scheduler.
func SetCronConfigSave(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	enabled := cast.ToInt(dataMap[`cron_daily_report_enabled`])
	triggerTime := strings.TrimSpace(cast.ToString(dataMap[`cron_daily_report_time`]))
	if enabled == 1 {
		if triggerTime == `` {
			gsgin.GinResponseError(c, `启用定时任务时触发时间不能为空`, nil)
			return
		}
		if _, err := time.Parse(`15:04`, triggerTime); err != nil {
			gsgin.GinResponseError(c, `时间格式无效，请使用 HH:MM 格式`, nil)
			return
		}
	}
	if err := common.DbMain.CronTaskSave(define.CronTaskTypeDailyReport, `AI 生成工作日报`, enabled, triggerTime); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if err := business.HotReloadCronScheduler(); err != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`配置已保存但热重载失败: %s`, err.Error()), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetAccountGroupDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetHomeTaskConfigGet 返回任务清单配置页面数据。
func SetHomeTaskConfigGet(c *gin.Context) {
	dailyReportPrompt, err := homeTaskConfigValue(define.HomeTaskConfigDailyReportPrompt)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dailyReportModelID, err := homeTaskConfigValue(define.HomeTaskConfigDailyReportModelID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	fragmentPrompt, err := homeTaskConfigValue(define.HomeTaskConfigFragmentPrompt)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	tapdSmartLinkID, err := homeTaskConfigValue(define.HomeTaskConfigTapdSmartLinkID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	tapdLinkLabel, err := homeTaskConfigValue(define.HomeTaskConfigTapdLinkLabel)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	tapdCssSelector, err := homeTaskConfigValue(define.HomeTaskConfigTapdCssSelector)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	tapdWaitSeconds, err := homeTaskConfigValue(define.HomeTaskConfigTapdWaitSeconds)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	promptDev, err := homeTaskConfigValue(define.HomeTaskConfigPromptDev)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	promptApiGen, err := homeTaskConfigValue(define.HomeTaskConfigPromptApiGen)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	promptApiTest, err := homeTaskConfigValue(define.HomeTaskConfigPromptApiTest)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	promptDesign, err := homeTaskConfigValue(define.HomeTaskConfigPromptDesign)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	devEnvironment, err := homeTaskConfigValue(define.HomeTaskConfigDevEnvironment)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`home_task_daily_report_prompt`:   dailyReportPrompt,
		`home_task_daily_report_model_id`: cast.ToInt(dailyReportModelID),
		`home_task_fragment_prompt`:       fragmentPrompt,
		`home_task_tapd_smart_link_id`:    cast.ToInt(tapdSmartLinkID),
		`home_task_tapd_link_label`:       tapdLinkLabel,
		`home_task_tapd_css_selector`:     tapdCssSelector,
		`home_task_tapd_wait_seconds`:     cast.ToInt(tapdWaitSeconds),
		`home_task_prompt_dev`:            promptDev,
		`home_task_prompt_api_gen`:        promptApiGen,
		`home_task_prompt_api_test`:       promptApiTest,
		`home_task_prompt_design`:         promptDesign,
		`home_task_dev_environment`:       devEnvironment,
	})
}

// promptConfigKeys 需要记录变更日志的提示词配置 key 及其中文名称。
var promptConfigKeys = map[string]string{
	define.HomeTaskConfigDailyReportPrompt: `工作日报提示词`,
	define.HomeTaskConfigFragmentPrompt:    `任务知识片段提示词`,
	define.HomeTaskConfigPromptDev:         `需求分析设计提示词`,
	define.HomeTaskConfigPromptApiGen:      `接口生成提示词`,
	define.HomeTaskConfigPromptApiTest:     `接口自动化测试提示词`,
	define.HomeTaskConfigPromptDesign:      `开发设计提示词`,
	define.HomeTaskConfigDevEnvironment:    `开发环境`,
}

// saveHomeTaskPromptWithLog 保存提示词配置并记录变更日志（仅当值真正变化时才写日志）。
func saveHomeTaskPromptWithLog(key, name, newValue, desc string) {
	oldValue, _ := homeTaskConfigValue(key)
	if oldValue == newValue {
		return
	}
	_ = common.DbMain.PromptChangeLogSave(key, name, oldValue, newValue)
}

// SetHomeTaskConfigSave 保存任务清单配置。
func SetHomeTaskConfigSave(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	homeTaskDailyReportPrompt := strings.TrimSpace(cast.ToString(dataMap[`home_task_daily_report_prompt`]))
	if homeTaskDailyReportPrompt == `` {
		homeTaskDailyReportPrompt = defaultHomeTaskDailyReportPrompt()
	}
	homeTaskDailyReportModelID := cast.ToInt(dataMap[`home_task_daily_report_model_id`])
	if homeTaskDailyReportModelID > 0 {
		modelInfo, err := common.DbMain.AiModelInfo(homeTaskDailyReportModelID)
		if err != nil {
			gsgin.GinResponseError(c, `AI 模型不存在`, nil)
			return
		}
		if strings.ToLower(cast.ToString(modelInfo[`model_type`])) != `llm` {
			gsgin.GinResponseError(c, `工作日报仅支持选择 LLM 模型`, nil)
			return
		}
	}
	saveHomeTaskPromptWithLog(define.HomeTaskConfigDailyReportPrompt, `工作日报提示词`, homeTaskDailyReportPrompt, `首页任务工作日报 AI 提示词`)
	if err := common.DbMain.HomeTaskConfigSave(`工作日报提示词`, define.HomeTaskConfigDailyReportPrompt, homeTaskDailyReportPrompt, `首页任务工作日报 AI 提示词`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if err := common.DbMain.HomeTaskConfigSave(`工作日报模型`, define.HomeTaskConfigDailyReportModelID, cast.ToString(homeTaskDailyReportModelID), `首页任务工作日报所用模型 id`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskFragmentPrompt := strings.TrimSpace(cast.ToString(dataMap[`home_task_fragment_prompt`]))
	saveHomeTaskPromptWithLog(define.HomeTaskConfigFragmentPrompt, `任务知识片段提示词`, homeTaskFragmentPrompt, `新建任务时自动创建知识片段的提示词模板`)
	if err := common.DbMain.HomeTaskConfigSave(`任务知识片段提示词`, define.HomeTaskConfigFragmentPrompt, homeTaskFragmentPrompt, `新建任务时自动创建知识片段的提示词模板`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskTapdSmartLinkID := cast.ToString(cast.ToInt(dataMap[`home_task_tapd_smart_link_id`]))
	if err := common.DbMain.HomeTaskConfigSave(`TAPD自定义网页ID`, define.HomeTaskConfigTapdSmartLinkID, homeTaskTapdSmartLinkID, `TAPD登录页所选自定义网页ID`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskTapdLinkLabel := strings.TrimSpace(cast.ToString(dataMap[`home_task_tapd_link_label`]))
	if err := common.DbMain.HomeTaskConfigSave(`TAPD链接标签`, define.HomeTaskConfigTapdLinkLabel, homeTaskTapdLinkLabel, `TAPD登录页所选链接的label`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskTapdCssSelector := strings.TrimSpace(cast.ToString(dataMap[`home_task_tapd_css_selector`]))
	if err := common.DbMain.HomeTaskConfigSave(`TAPD抓取CSS选择器`, define.HomeTaskConfigTapdCssSelector, homeTaskTapdCssSelector, `TAPD网页抓取区域CSS选择器`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskTapdWaitSeconds := cast.ToString(cast.ToInt(dataMap[`home_task_tapd_wait_seconds`]))
	if err := common.DbMain.HomeTaskConfigSave(`TAPD抓取等待秒数`, define.HomeTaskConfigTapdWaitSeconds, homeTaskTapdWaitSeconds, `TAPD网页抓取前等待秒数`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskPromptDev := strings.TrimSpace(cast.ToString(dataMap[`home_task_prompt_dev`]))
	saveHomeTaskPromptWithLog(define.HomeTaskConfigPromptDev, `需求分析设计提示词`, homeTaskPromptDev, `工作流-需求开发提示词模板`)
	if err := common.DbMain.HomeTaskConfigSave(`需求分析设计提示词`, define.HomeTaskConfigPromptDev, homeTaskPromptDev, `工作流-需求开发提示词模板`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskPromptApiGen := strings.TrimSpace(cast.ToString(dataMap[`home_task_prompt_api_gen`]))
	saveHomeTaskPromptWithLog(define.HomeTaskConfigPromptApiGen, `接口生成提示词`, homeTaskPromptApiGen, `工作流-接口生成提示词模板`)
	if err := common.DbMain.HomeTaskConfigSave(`接口生成提示词`, define.HomeTaskConfigPromptApiGen, homeTaskPromptApiGen, `工作流-接口生成提示词模板`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskPromptApiTest := strings.TrimSpace(cast.ToString(dataMap[`home_task_prompt_api_test`]))
	saveHomeTaskPromptWithLog(define.HomeTaskConfigPromptApiTest, `接口自动化测试提示词`, homeTaskPromptApiTest, `工作流-接口自动化测试提示词模板`)
	if err := common.DbMain.HomeTaskConfigSave(`接口自动化测试提示词`, define.HomeTaskConfigPromptApiTest, homeTaskPromptApiTest, `工作流-接口自动化测试提示词模板`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskPromptDesign := strings.TrimSpace(cast.ToString(dataMap[`home_task_prompt_design`]))
	saveHomeTaskPromptWithLog(define.HomeTaskConfigPromptDesign, `开发设计提示词`, homeTaskPromptDesign, `工作流-开发设计提示词模板`)
	if err := common.DbMain.HomeTaskConfigSave(`开发设计提示词`, define.HomeTaskConfigPromptDesign, homeTaskPromptDesign, `工作流-开发设计提示词模板`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskDevEnvironment := strings.TrimSpace(cast.ToString(dataMap[`home_task_dev_environment`]))
	saveHomeTaskPromptWithLog(define.HomeTaskConfigDevEnvironment, `开发环境`, homeTaskDevEnvironment, `工作流-开发环境描述`)
	if err := common.DbMain.HomeTaskConfigSave(`开发环境`, define.HomeTaskConfigDevEnvironment, homeTaskDevEnvironment, `工作流-开发环境描述`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetLocalDirList 浏览本地目录，返回指定路径下的子目录列表。
func SetLocalDirList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	dirPath := strings.TrimSpace(cast.ToString(dataMap[`path`]))
	if dirPath == `` {
		// 未传路径时返回根目录（Windows 返回驱动器列表，其他返回 /）
		if drives, err := listWindowsDrives(); err == nil && len(drives) > 0 {
			gsgin.GinResponseSuccess(c, ``, drives)
			return
		}
		dirPath = `/`
	}

	info, statErr := os.Stat(dirPath)
	if statErr != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`路径不可访问: %s`, statErr.Error()), nil)
		return
	}
	if !info.IsDir() {
		gsgin.GinResponseError(c, `指定路径不是目录`, nil)
		return
	}

	entries, readErr := os.ReadDir(dirPath)
	if readErr != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`读取目录失败: %s`, readErr.Error()), nil)
		return
	}

	result := make([]map[string]any, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, `.`) || name == `$RECYCLE.BIN` || name == `System Volume Information` {
			continue
		}
		fullPath := filepath.Join(dirPath, name)
		hasChildren := false
		if subEntries, subErr := os.ReadDir(fullPath); subErr == nil {
			for _, sub := range subEntries {
				if sub.IsDir() && !strings.HasPrefix(sub.Name(), `.`) {
					hasChildren = true
					break
				}
			}
		}
		result = append(result, map[string]any{
			`label`:        name,
			`value`:        fullPath,
			`has_children`: hasChildren,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return cast.ToString(result[i][`label`]) < cast.ToString(result[j][`label`])
	})

	gsgin.GinResponseSuccess(c, ``, result)
}

// listWindowsDrives 返回 Windows 系统上可用的驱动器盘符列表。
func listWindowsDrives() ([]map[string]any, error) {
	drives := make([]map[string]any, 0)
	for _, letter := range `ABCDEFGHIJKLMNOPQRSTUVWXYZ` {
		drive := string(letter) + `:/`
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, map[string]any{
				`label`:        drive,
				`value`:        drive,
				`has_children`: true,
			})
		}
	}
	return drives, nil
}

// SetPromptChangeLogList 返回提示词变更日志（最近 20 条）。
func SetPromptChangeLogList(c *gin.Context) {
	list, err := common.DbMain.PromptChangeLogList(20)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	for i := range list {
		list[i][`create_time_desc`] = gstool.TimeUnixToString(time.Unix(cast.ToInt64(list[i][`create_time`]), 0), `Y-m-d H:i:s`)
	}
	gsgin.GinResponseSuccess(c, ``, list)
}
