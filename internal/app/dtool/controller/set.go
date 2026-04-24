package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"path/filepath"
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
		_, _ = common.DbMain.Client.QuickCreate(`tbl_mysql`, updateData).Exec()
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_mysql`,
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
	arrangePrompt, err := memoryConfigValue(define.GlobalMemoryArrangePrompt)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	arrangeModelID, err := memoryConfigValue(define.GlobalMemoryArrangeModelID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dailyReportPrompt, err := memoryConfigValue(define.GlobalHomeTaskDailyReportPrompt)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dailyReportModelID, err := memoryConfigValue(define.GlobalHomeTaskDailyReportModelID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	fragmentPrompt, err := memoryConfigValue(define.GlobalHomeTaskFragmentPrompt)
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
		`memory_db_name`:                    ``,
		`memory_db_configured`:              memoryConfig.Dir != ``,
		`memory_db_is_git_repo`:             memoryConfig.GitRepoEnabled,
		`memory_db_auto_push_delay_minutes`: memoryConfig.AutoPushDelayMinutes,
		`memory_config_file`:                memoryConfigFilePath(),
		`memory_arrange_prompt`:             arrangePrompt,
		`memory_arrange_model_id`:           cast.ToInt(arrangeModelID),
		`home_task_daily_report_prompt`:     dailyReportPrompt,
		`home_task_daily_report_model_id`:   cast.ToInt(dailyReportModelID),
		`home_task_fragment_prompt`:         fragmentPrompt,
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
	if err := common.DbMain.SetGlobalValue(`记忆整理提示词`, define.GlobalMemoryArrangePrompt, memoryArrangePrompt, `知识片段 AI 整理提示词`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if err := common.DbMain.SetGlobalValue(`记忆整理模型`, define.GlobalMemoryArrangeModelID, cast.ToString(memoryArrangeModelID), `知识片段 AI 整理所用模型 id`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
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
		// 工作日报仅允许使用 LLM 模型 / only LLM models are allowed for daily report.
		if strings.ToLower(cast.ToString(modelInfo[`model_type`])) != `llm` {
			gsgin.GinResponseError(c, `工作日报仅支持选择 LLM 模型`, nil)
			return
		}
	}
	if err := common.DbMain.SetGlobalValue(`工作日报提示词`, define.GlobalHomeTaskDailyReportPrompt, homeTaskDailyReportPrompt, `首页任务工作日报 AI 提示词`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if err := common.DbMain.SetGlobalValue(`工作日报模型`, define.GlobalHomeTaskDailyReportModelID, cast.ToString(homeTaskDailyReportModelID), `首页任务工作日报所用模型 id`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskFragmentPrompt := strings.TrimSpace(cast.ToString(dataMap[`home_task_fragment_prompt`]))
	if err := common.DbMain.SetGlobalValue(`任务知识片段提示词`, define.GlobalHomeTaskFragmentPrompt, homeTaskFragmentPrompt, `新建任务时自动创建知识片段的提示词模板`); err != nil {
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

func memoryConfigValue(key string) (string, error) {
	value, err := common.DbMain.GlobalValue(key)
	if err != nil {
		if memoryConfigValueMissing(err) {
			return ``, nil
		}
		return ``, err
	}
	return value, nil
}

func memoryConfigValueMissing(err error) bool {
	errText := strings.ToLower(err.Error())
	return strings.Contains(errText, `not found`) || strings.Contains(errText, `no rows`)
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
