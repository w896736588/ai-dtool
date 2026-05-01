package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
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

const (
	// dockerImageFieldCount 镜像列表输出字段数量。
	dockerImageFieldCount = 5
	// dockerContainerFieldCount 容器列表输出字段数量。
	dockerContainerFieldCount = 5
	// dockerActionTimeoutSeconds Docker 相关操作统一超时时间，避免前端长时间无响应。
	dockerActionTimeoutSeconds = 40
	// byteSizeUnit 1024 进制字节单位换算基数。
	byteSizeUnit = 1024
)

var (
	// byteSizeUnits 字节格式化展示单位。
	byteSizeUnits = []string{"B", "KB", "MB", "GB", "TB"}
)

func DockerComposeList(c *gin.Context) {
	dataMap := make(map[string]any)
	if err := gsgin.GinPostBody(c, &dataMap); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	queryWhere := map[string]any{
		`status`: 1,
	}
	sshId := cast.ToInt(dataMap[`ssh_id`])
	if sshId > 0 {
		queryWhere[`ssh_id`] = sshId
	}
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, queryWhere).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	allSsh, allSshErr := common.DbMain.Client.QuickQuery(`tbl_ssh`, `*`, nil).All()
	if allSshErr != nil {
		gsgin.GinResponseError(c, allSshErr.Error(), nil)
		return
	}
	applyDockerComposeSshNames(all, allSsh)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: all,
	})
}

func applyDockerComposeSshNames(composeList []map[string]any, sshList []map[string]any) {
	sshNameMap := make(map[int]string, len(sshList))
	for _, sshValue := range sshList {
		sshId := cast.ToInt(sshValue[`id`])
		if sshId == 0 {
			continue
		}
		sshNameMap[sshId] = cast.ToString(sshValue[`name`])
	}
	for _, composeValue := range composeList {
		composeValue[`ssh_name`] = sshNameMap[cast.ToInt(composeValue[`ssh_id`])]
	}
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
	one, oneErr := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	envFile := cast.ToString(one[`env_file`])
	composeYmlPath := one[`compose_yml_path`].(string)
	command1 := p_shell.NewCommand()
	command1.Sudo()
	command1.Cd(path.Dir(composeYmlPath))
	command1.DockerComposeServices(cast.ToString(one[`docker_cmd`]), envFile)
	result1, _ := sshClient.RunCommandWait(command1.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	services := parseDockerComposeServiceNames(result1)
	list := make([]map[string]any, 0)
	for _, v := range services {
		list = append(list, map[string]any{
			`name`: v,
		})
	}
	gstool.ArrayMapSort(&list, `name`, `asc`)
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
	catCommand := p_shell.NewCommand().Sudo().Cat(cast.ToString(data[`config_path`]))
	ret, _ := sshClient.RunCommandWait(catCommand.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
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
	one, oneErr := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	service := cast.ToString(data[`service`])
	envFile := cast.ToString(one[`env_file`])
	composeYmlPath := one[`compose_yml_path`].(string)
	command := p_shell.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeRestart(cast.ToString(one[`docker_cmd`]), envFile, []string{service})
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
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
	one, oneErr := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	composeYmlPath := one[`compose_yml_path`].(string)
	envFile := cast.ToString(one[`env_file`])
	command := p_shell.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeStatus(cast.ToString(one[`docker_cmd`]), envFile)
	status, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	headers := []string{`服务名`, `CPU 使用率`, `内存用量 / 内存上限`, `内存使用率`, `网络收发流量`, `磁盘块设备读写量`}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`status`:  ParseStats(status),
		`headers`: headers,
	})
}

var (
	ansi  = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	space = regexp.MustCompile(`\s{2,}`) // 2+ 空格 → \t
	// docker compose service 名过滤规则：首字符字母/数字，后续允许字母数字._-
	dockerComposeServiceNameReg = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)
)

// parseDockerComposeServiceNames 清洗 ssh 命令回显，仅保留合法的 compose service 名。
func parseDockerComposeServiceNames(raw string) []string {
	lines := strings.Split(raw, "\n")
	services := make([]string, 0, len(lines))
	seen := make(map[string]struct{})
	for _, line := range lines {
		// 兼容优化后 ssh 输出中的提示符、回显和控制字符
		clean := ansi.ReplaceAllString(line, "")
		clean = strings.TrimSpace(strings.ReplaceAll(clean, "\r", ""))
		if clean == "" {
			continue
		}
		if !dockerComposeServiceNameReg.MatchString(clean) {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		services = append(services, clean)
	}
	return services
}

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
	one, oneErr := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	service := cast.ToString(data[`service`])
	composeYmlPath := one[`compose_yml_path`].(string)
	envFile := cast.ToString(one[`env_file`])
	command := p_shell.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	if service != `` {
		command.DockerComposeStopService(cast.ToString(one[`docker_cmd`]), envFile, []string{service})
	} else {
		command.DockerComposeStop(cast.ToString(one[`docker_cmd`]), envFile)
	}
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
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
	one, oneErr := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: id,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	composeYmlPath := one[`compose_yml_path`].(string)
	envFile := cast.ToString(one[`env_file`])
	service := cast.ToString(data[`service`])
	command := p_shell.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeUpd(cast.ToString(one[`docker_cmd`]), envFile, service)
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

// DockerContainerLogTruncate 清理当前环境全部 Docker 容器日志。
func DockerContainerLogTruncate(c *gin.Context) {
	_, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	command := p_shell.NewCommand()
	command.Sudo()
	command.DockerContainerLogTruncate()
	result, runErr := sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	if runErr != nil {
		gsgin.GinResponseError(c, runErr.Error(), map[string]any{
			`raw`: result,
		})
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func DockerImageList(c *gin.Context) {
	_, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	command := p_shell.NewCommand()
	command.Sudo()
	command.DockerImageList()
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: parseDockerImageRows(result),
	})
}

func DockerImageContainers(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	imageRef := cast.ToString(data[`image_ref`])
	if imageRef == `` {
		gsgin.GinResponseError(c, `image_ref is empty`, nil)
		return
	}
	command := p_shell.NewCommand()
	command.Sudo()
	command.DockerImageContainers(imageRef)
	result, _ := sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: parseDockerContainerRows(result),
	})
}

func DockerImageRemove(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	imageRef := cast.ToString(data[`image_ref`])
	if imageRef == `` {
		gsgin.GinResponseError(c, `image_ref is empty`, nil)
		return
	}
	command := p_shell.NewCommand()
	command.Sudo()
	command.DockerImageRemove(imageRef)
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func DockerContainerStop(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	containerId := cast.ToString(data[`container_id`])
	if containerId == `` {
		gsgin.GinResponseError(c, `container_id is empty`, nil)
		return
	}
	command := p_shell.NewCommand()
	command.Sudo()
	command.DockerContainerStop(containerId)
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func DockerContainerRemove(c *gin.Context) {
	data, sshClient, err := getDockerComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	containerId := cast.ToString(data[`container_id`])
	if containerId == `` {
		gsgin.GinResponseError(c, `container_id is empty`, nil)
		return
	}
	command := p_shell.NewCommand()
	command.Sudo()
	command.DockerContainerRemove(containerId)
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

func parseDockerImageRows(raw string) []map[string]string {
	rows := parseDockerRows(raw, dockerImageFieldCount)
	list := make([]map[string]string, 0, len(rows))
	for _, fields := range rows {
		imageRef := fields[2]
		if fields[0] != `<none>` && fields[1] != `<none>` {
			imageRef = fields[0] + `:` + fields[1]
		}
		list = append(list, map[string]string{
			`repository`: fields[0],
			`tag`:        fields[1],
			`image_id`:   fields[2],
			`created`:    fields[3],
			`size`:       fields[4],
			`image_ref`:  imageRef,
		})
	}
	return list
}

func parseDockerContainerRows(raw string) []map[string]string {
	rows := parseDockerRows(raw, dockerContainerFieldCount)
	list := make([]map[string]string, 0, len(rows))
	for _, fields := range rows {
		list = append(list, map[string]string{
			`container_id`:   fields[0],
			`container_name`: fields[1],
			`image`:          fields[2],
			`state`:          fields[3],
			`status`:         fields[4],
		})
	}
	return list
}

// formatByteSize 将字节数转换为便于页面展示的大小字符串。
func formatByteSize(size int64) string {
	if size < byteSizeUnit {
		return cast.ToString(size) + `B`
	}
	value := float64(size)
	unitIndex := 0
	for value >= byteSizeUnit && unitIndex < len(byteSizeUnits)-1 {
		value = value / byteSizeUnit
		unitIndex++
	}
	return fmt.Sprintf(`%.2f%s`, value, byteSizeUnits[unitIndex])
}

func parseDockerRows(raw string, fieldCount int) [][]string {
	lines := strings.Split(raw, "\n")
	list := make([][]string, 0, len(lines))
	for _, line := range lines {
		clean := ansi.ReplaceAllString(line, "")
		clean = strings.TrimSpace(strings.ReplaceAll(clean, "\r", ""))
		if clean == "" {
			continue
		}
		fields := strings.Split(clean, "\t")
		if len(fields) != fieldCount {
			continue
		}
		list = append(list, fields)
	}
	return list
}

func getDockerComponent(c *gin.Context) (map[string]interface{}, *gsssh.SshTerminal, error) {
	dataMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &dataMap)
	if err != nil {
		return nil, nil, err
	}
	sshId := dataMap[`ssh_id`]
	if cast.ToInt(sshId) == 0 {
		return nil, nil, errors.New(`缺少ssh_id参数`)
	}
	sseDistributeId := cast.ToString(dataMap[`sse_distribute_id`])
	sshConfig, _ := common.DbMain.GetSshConfig(sshId)
	uniqueKey := p_common.TBaseClient.GetCombineKey(sshId, sseDistributeId)
	//sshClient, sshClientErr := base.Component.TShell.GetClient(sshConfig, uniqueKey, sseId, func(s string) []string {
	//	return stripANSI(s)
	//})
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: sseDistributeId,
	}
	sshClient, sshClientErr := component.ShellClient.GetClient(sshConfig, uniqueKey, sse, func(s string) []string {
		return []string{p_common.TBaseClient.FilterTerminalChars(s)}
	}, nil, nil)
	if sshClientErr != nil {
		return nil, nil, sshClientErr
	}
	return dataMap, sshClient, nil
}

// DockerServiceRestart 通过 docker_id 重启指定的 Docker Compose 服务。
// 与 DockerComposeRestart 不同，本接口只需传 docker_id 和 service，
// ssh_id 从 tbl_docker_compose 配置中自动解析，方便外部脚本和 Agent 调用。
func DockerServiceRestart(c *gin.Context) {
	dataMap := make(map[string]any)
	if err := gsgin.GinPostBody(c, &dataMap); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dockerId := cast.ToInt(dataMap[`docker_id`])
	if dockerId == 0 {
		gsgin.GinResponseError(c, `docker_id is empty`, nil)
		return
	}
	service := cast.ToString(dataMap[`service`])
	if service == `` {
		gsgin.GinResponseError(c, `service is empty`, nil)
		return
	}
	// 查询 compose 配置
	one, oneErr := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: dockerId,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	// 从配置中自动获取 ssh_id 并建立 SSH 连接
	sshId := cast.ToString(one[`ssh_id`])
	sshConfig, _ := common.DbMain.GetSshConfig(sshId)
	sse := &p_sse.SseShell{
		Sse: gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
	}
	uniqueKey := p_common.TBaseClient.GetCombineKey(sshId, ``)
	sshClient, sshClientErr := component.ShellClient.GetClient(sshConfig, uniqueKey, sse, func(s string) []string {
		return []string{p_common.TBaseClient.FilterTerminalChars(s)}
	}, nil, nil)
	if sshClientErr != nil {
		gsgin.GinResponseError(c, sshClientErr.Error(), nil)
		return
	}
	// 执行 docker compose restart 指定服务
	envFile := cast.ToString(one[`env_file`])
	composeYmlPath := one[`compose_yml_path`].(string)
	command := p_shell.NewCommand()
	command.Sudo()
	command.Cd(path.Dir(composeYmlPath))
	command.DockerComposeRestart(cast.ToString(one[`docker_cmd`]), envFile, []string{service})
	_, _ = sshClient.RunCommandWait(command.GetCommand().ToStr(), dockerActionTimeoutSeconds*time.Second)
	gsgin.GinResponseSuccess(c, ``, map[string]any{})
}

// dockerComposeLogsPrefix docker compose logs 命令允许的前缀
const dockerComposeLogsPrefix = `docker compose logs`

// DockerServiceLogs 通过 docker_id 查询 Docker Compose 服务日志。
// 只需传入 docker_id 和 command（必须以 "docker compose logs" 开头），
// ssh_id 从 tbl_docker_compose 配置中自动解析，方便外部脚本和 Agent 调用。
func DockerServiceLogs(c *gin.Context) {
	dataMap := make(map[string]any)
	if err := gsgin.GinPostBody(c, &dataMap); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	dockerId := cast.ToInt(dataMap[`docker_id`])
	if dockerId == 0 {
		gsgin.GinResponseError(c, `docker_id is empty`, nil)
		return
	}
	commandStr := cast.ToString(dataMap[`command`])
	if commandStr == `` {
		gsgin.GinResponseError(c, `command is empty`, nil)
		return
	}
	// 校验 command 必须以 "docker compose logs" 开头
	if !strings.HasPrefix(commandStr, dockerComposeLogsPrefix) {
		gsgin.GinResponseError(c, `command 必须以 "docker compose logs" 开头`, nil)
		return
	}
	// 禁止持续输出参数，避免命令挂住
	if strings.Contains(commandStr, ` -f`) || strings.Contains(commandStr, ` --follow`) {
		gsgin.GinResponseError(c, `禁止使用 -f / --follow 参数，会导致持续输出`, nil)
		return
	}
	// 查询 compose 配置
	one, oneErr := common.DbMain.Client.QuickQuery(`tbl_docker_compose`, `*`, map[string]any{
		`id`: dockerId,
	}).One()
	if oneErr != nil {
		gsgin.GinResponseError(c, oneErr.Error(), nil)
		return
	}
	// 从配置中自动获取 ssh_id 并建立 SSH 连接
	sshId := cast.ToString(one[`ssh_id`])
	sshConfig, _ := common.DbMain.GetSshConfig(sshId)
	sse := &p_sse.SseShell{
		Sse: gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
	}
	uniqueKey := p_common.TBaseClient.GetCombineKey(sshId, ``)
	sshClient, sshClientErr := component.ShellClient.GetClient(sshConfig, uniqueKey, sse, func(s string) []string {
		return []string{p_common.TBaseClient.FilterTerminalChars(s)}
	}, nil, nil)
	if sshClientErr != nil {
		gsgin.GinResponseError(c, sshClientErr.Error(), nil)
		return
	}
	// 在 compose yml 目录下执行用户提供的 logs 命令
	composeYmlPath := one[`compose_yml_path`].(string)
	command := p_shell.NewCommand()
	command.Cd(path.Dir(composeYmlPath))
	// SetCommand 不会自动追加 sudo 前缀，需手动拼接
	command.SetCommand("sudo " + commandStr)
	fullCommand := command.GetCommand().ToStr()
	result, _ := sshClient.RunCommandWait(fullCommand, dockerActionTimeoutSeconds*time.Second)
	// 清洗 ANSI 控制字符
	cleanResult := ansi.ReplaceAllString(result, "")
	cleanResult = strings.TrimSpace(cleanResult)
	// 移除末尾可能混入的 SSH 提示符（如 user@host:path$）
	lines := strings.Split(cleanResult, "\n")
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		// SSH 提示符通常以 $ 或 # 结尾，且包含 @ 和 :
		if strings.Contains(lastLine, "@") && strings.Contains(lastLine, ":") &&
			(strings.HasSuffix(lastLine, "$") || strings.HasSuffix(lastLine, "#")) {
			lines = lines[:len(lines)-1]
			cleanResult = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`logs`: cleanResult,
	})
}
