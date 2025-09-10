package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstask"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// SetSshList ssh列表
func SetSshList(c *gin.Context) {
	all, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	allSsh := map[int]map[string]any{}
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
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_ssh`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_ssh`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_ssh`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGitList(c *gin.Context) {
	allGit, allGitErr := base.Component.TSqlite.Client.QuickQuery(`tbl_git`, `*`, nil).All()
	if allGitErr != nil {
		gsgin.GinResponseError(c, allGitErr.Error(), nil)
		return
	}
	allGitGroup, allGitGroupErr := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeGit,
	}).All()
	if allGitGroupErr != nil {
		gsgin.GinResponseError(c, allGitGroupErr.Error(), nil)
		return
	}
	allSsh, allSshErr := base.Component.TSqlite.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
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
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_git`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_git`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_git`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGitGroupList(c *gin.Context) {
	all, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
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
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_group`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_group`, map[string]any{
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
	sshList, sshListErr := base.Component.TSqlite.GetAllSshConfig()
	if sshListErr != nil {
		gsgin.GinResponseError(c, sshListErr.Error(), nil)
		return
	}
	searchDir := cast.ToString(dataMap[`dir`])
	existMap := make(map[string]string)
	gitDirList := make([]map[string]any, 0)
	for _, sshConfig := range sshList {
		findDirList := base.Component.TCode.FindCode(sshConfig, searchDir)
		for _, findDir := range findDirList {
			if strings.Index(findDir, searchDir) != 0 {
				continue
			}
			if existMap[findDir] == `EXIST` {
				continue
			}
			existMap[findDir] = `EXIST`
			//查找group_id
			gitInfo, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_git`, `git_group_id`, map[string]any{
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
	allSupervisor, allSupervisorErr := base.Component.TSqlite.Client.QuickQuery(`tbl_supervisor`, `*`, nil).All()
	if allSupervisorErr != nil {
		gsgin.GinResponseError(c, allSupervisorErr.Error(), nil)
		return
	}
	allSsh, allSshErr := base.Component.TSqlite.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
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
		_, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_supervisor`, updateData).Exec()
		if createErr != nil {
			gstool.FmtPrintlnLogTime(`创建失败 %s`, createErr.Error())
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_supervisor`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_supervisor`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SetRedisList redis列表
func SetRedisList(c *gin.Context) {
	allRedis, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_redis`, `*`, nil).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	allSsh, allSshErr := base.Component.TSqlite.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
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
		sshConfig, sshConfigErr := base.Component.TSqlite.GetSshConfig(redisConfig[`ssh_id`])
		if sshConfigErr != nil {
			return &gstask.Result{
				Err:    gstool.Error(`获取ssh配置失败 %s`, sshConfigErr.Error()),
				Result: redisConfig[`id`],
			}
		}
		gsRedis.Ssh = &gsssh.SshConfig{
			Name:     cast.ToString(sshConfig[`name`]),
			Host:     cast.ToString(sshConfig[`host`]),
			Port:     cast.ToString(sshConfig[`port`]),
			UserName: cast.ToString(sshConfig[`username`]),
			Password: cast.ToString(sshConfig[`password`]),
			GsSlog:   base.Component.GsLog,
		}
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
	ssh := &gsssh.SshConfig{
		Name:     cast.ToString(sshConfig[`name`]),
		Host:     cast.ToString(sshConfig[`host`]),
		Port:     cast.ToString(sshConfig[`port`]),
		Password: cast.ToString(sshConfig[`password`]),
		UserName: cast.ToString(sshConfig[`username`]),
	}
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

func SetRedisAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `host`, `port`, `username`, `password`, `ssh_id`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_redis`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_redis`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_redis`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetMysqlList(c *gin.Context) {
	allMysql, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_mysql`, `*`, nil).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	allSsh, allSshErr := base.Component.TSqlite.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	if allSshErr != nil {
		gsgin.GinResponseError(c, allSshErr.Error(), nil)
		return
	}
	for gitKey, gitValue := range allMysql {
		allMysql[gitKey][`ssh_name`] = ``
		gitSshId := cast.ToInt(gitValue[`ssh_id`])
		if gitSshId != 0 {
			for _, sshValue := range allSsh {
				if cast.ToInt(sshValue[`id`]) == gitSshId {
					allMysql[gitKey][`ssh_name`] = sshValue[`name`]
				}
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, allMysql)
}

func SetMysqlAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `host`, `port`, `username`, `dbname`, `password`, `ssh_id`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_mysql`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_mysql`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_mysql`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetVariableGroupList(c *gin.Context) {
	all, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
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
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_group`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetCmdGroupList(c *gin.Context) {
	all, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
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
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_group`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetSmartLinkGroupList(c *gin.Context) {
	all, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
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
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_group`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetDockerComposeList(c *gin.Context) {
	all, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`status`: 1,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

func SetDockerComposeAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `compose_yml_path`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_docker_compose`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_docker_compose`,
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
		ret, err := base.Component.TSqlite.Client.QuickUpdate(`tbl_docker_compose`, map[string]any{
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
	allGit, allGitErr := base.Component.TSqlite.Client.QuickQuery(`tbl_gitlab_token`, `*`, nil).All()
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
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_gitlab_token`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_gitlab_token`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_gitlab_token`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetGlobalList(c *gin.Context) {
	allGit, allGitErr := base.Component.TSqlite.Client.QuickQuery(`tbl_global`, `*`, nil).All()
	if allGitErr != nil {
		gsgin.GinResponseError(c, allGitErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, allGit)
}

func SetGlobalAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`key`, `value`, `name`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_global`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_global`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_global`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetAccountList(c *gin.Context) {
	allAccount, allAccountErr := base.Component.TSqlite.Client.QuickQuery(`tbl_account`, `*`, nil).All()
	if allAccountErr != nil {
		gsgin.GinResponseError(c, allAccountErr.Error(), nil)
		return
	}
	allAccountGroup, allAccountGroupErr := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
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
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_account`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_account`,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_account`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SetAccountGroupList(c *gin.Context) {
	all, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
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
		_, _ = base.Component.TSqlite.Client.QuickCreate(`tbl_group`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_group`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}
