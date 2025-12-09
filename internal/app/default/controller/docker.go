package controller

import (
	"dev_tool/base"
	"errors"
	"path"
	"regexp"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdefine"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func DockerComposeList(c *gin.Context) {
	dataMap, _, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	sshId := cast.ToInt(dataMap[`ssh_id`])
	all, allErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`status`: 1,
		`ssh_id`: sshId,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: all,
	})
}

func DockerComposeServices(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	id := cast.ToInt(data[`id`])
	if id == 0 {
		gsgin.GinResponseError(c, `id is empty`, nil)
		return
	}
	one, oneErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	envFile := cast.ToString(one[`env_file`])
	composeYmlPath := one[`compose_yml_path`].(string)
	command1 := base.NewCommand()
	command1.Sudo()
	command1.Cd(path.Dir(composeYmlPath))
	command1.DockerComposeServices(cast.ToString(one[`docker_cmd`]), envFile)
	result1, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr(), 40*time.Second)
	services := strings.Split(result1, "\n")
	services = gstool.ArrayFilterEmpty(&services)
	services = services[1:]
	list := make([]map[string]any, 0)
	for _, v := range services {
		list = append(list, map[string]any{
			`name`: v,
		})
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`services`: list,
	})
}

func DockerComposeConfigShow(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	catCommand := base.NewCommand().Sudo().ConsumerConfigCat(cast.ToString(data[`config_path`]), ``)
	ret, _ := sshClient.RunCommandWait(catCommand.GetCommand().ToStr(), 40*time.Second)
	retMsgList := make([]string, 0)
	retMsgList = append(retMsgList, ret)
	gsgin.GinResponseSuccess(c, ``, strings.Join(retMsgList, gsdefine.Enter))
}

func DockerComposeRestart(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	id := cast.ToInt(data[`id`])
	if id == 0 {
		gsgin.GinResponseError(c, `id is empty`, nil)
		return
	}
	one, oneErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	service := cast.ToString(data[`service`])
	envFile := cast.ToString(one[`env_file`])
	composeYmlPath := one[`compose_yml_path`].(string)
	command := base.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeRestart(cast.ToString(one[`docker_cmd`]), envFile, []string{service})
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func DockerComposeStatus(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	id := cast.ToInt(data[`id`])
	if id == 0 {
		gsgin.GinResponseError(c, `id is empty`, nil)
		return
	}
	one, oneErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	composeYmlPath := one[`compose_yml_path`].(string)
	envFile := cast.ToString(one[`env_file`])
	command := base.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeStatus(cast.ToString(one[`docker_cmd`]), envFile)
	status, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	headers := []string{`服务名`, `CPU 使用率`, `内存用量 / 内存上限`, `内存使用率`, `网络收发流量`, `磁盘块设备读写量`}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`status`:  ParseStats(status),
		`headers`: headers,
	})
}

var (
	ansi  = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	space = regexp.MustCompile(`\s{2,}`) // 2+ 空格 → \t
)

func ParseStats(text string) []map[string]string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var head []string
	var list []map[string]string

	for _, raw := range lines {
		line := ansi.ReplaceAllString(raw, "")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = space.ReplaceAllString(line, "\t")
		fields := strings.Split(line, "\t")
		if len(fields) < 6 {
			continue
		}
		// 第一行当表头
		if head == nil {
			head = fields
			continue
		}
		// 数据行 → map
		row := make(map[string]string, len(head))
		for i, v := range fields {
			row[head[i]] = v
		}
		list = append(list, row)
	}
	return list
}

func DockerComposeStop(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	id := cast.ToInt(data[`id`])
	if id == 0 {
		gsgin.GinResponseError(c, `id is empty`, nil)
		return
	}
	one, oneErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	service := cast.ToString(data[`service`])
	composeYmlPath := one[`compose_yml_path`].(string)
	envFile := cast.ToString(one[`env_file`])
	command := base.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	if service != `` {
		command.DockerComposeStopService(cast.ToString(one[`docker_cmd`]), envFile, []string{service})
	} else {
		command.DockerComposeStop(cast.ToString(one[`docker_cmd`]), envFile)
	}
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func DockerComposeStart(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	id := cast.ToInt(data[`id`])
	if id == 0 {
		gsgin.GinResponseError(c, `id is empty`, nil)
		return
	}
	one, oneErr := base.Component.TSqlite.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	composeYmlPath := one[`compose_yml_path`].(string)
	envFile := cast.ToString(one[`env_file`])
	command := base.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeStart(cast.ToString(one[`docker_cmd`]), envFile)
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), 40*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func getDockerComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshTerminal, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, nil, err
	}
	sshId := reqMap[`ssh_id`]
	if cast.ToInt(sshId) == 0 {
		return nil, nil, errors.New(`缺少ssh_id参数`)
	}
	sseId := cast.ToString(reqMap[`sse_id`])
	sshConfig, _ := base.Component.TSqlite.GetSshConfig(sshId)
	uniqueKey := base.Component.TBase.GetCombineKey(sshId, sseId)
	//sshClient, sshClientErr := base.Component.TShell.GetClient(sshConfig, uniqueKey, sseId, func(s string) []string {
	//	return stripANSI(s)
	//})
	sshClient, sshClientErr := base.Component.TShell.GetClient(sshConfig, uniqueKey, sseId, nil)
	if sshClientErr != nil {
		return nil, nil, sshClientErr
	}
	return reqMap, sshClient, nil
}

var garbage = regexp.MustCompile(`(?m)\x1b\[[0-9;?]*[a-zA-Z]|[\r\x00]`)

// 可选：把旋转进度条那几行也整行干掉
var spinner = regexp.MustCompile(`(?m)^.*Container.*Restarting.*\n`)

// 加载进度的点
var braille = regexp.MustCompile(`[\x28\x00-\x28\xFF]`)
