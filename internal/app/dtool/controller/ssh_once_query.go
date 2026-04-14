package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/pkg/p_shell"
	"errors"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gsssh"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// oneShotCommandRunner 抽象一次性 SSH 执行器，供查询型接口使用。
type oneShotCommandRunner interface {
	RunCommandOnce(command string) (string, error)
	Close()
}

type sshOnceCommandRunner struct {
	client *gsssh.SshOnce
}

func (h *sshOnceCommandRunner) RunCommandOnce(command string) (string, error) {
	return h.client.RunCommandOnce(command)
}

func (h *sshOnceCommandRunner) Close() {}

type oneShotCommandRunnerFactory func(sshConfig map[string]any) (oneShotCommandRunner, error)
type oneShotCommandRunnerRelease func()

var defaultOneShotCommandRunnerFactory oneShotCommandRunnerFactory = func(sshConfig map[string]any) (oneShotCommandRunner, error) {
	client, err := component.ShellClient.GetSshOnce(sshConfig)
	if err != nil {
		return nil, err
	}
	return &sshOnceCommandRunner{client: client}, nil
}

var defaultOneShotCommandRunnerRelease oneShotCommandRunnerRelease = func() {}

func runSSHCommandOnceWithFactory(
	sshConfig map[string]any,
	commandText string,
	factory oneShotCommandRunnerFactory,
	release oneShotCommandRunnerRelease,
) (string, error) {
	if factory == nil {
		factory = defaultOneShotCommandRunnerFactory
	}
	if release == nil {
		release = defaultOneShotCommandRunnerRelease
	}

	runner, err := factory(sshConfig)
	if err != nil {
		return ``, err
	}
	defer func() {
		runner.Close()
		release()
	}()

	return runner.RunCommandOnce(commandText)
}

func getRequestDataAndSSHConfig(c *gin.Context) (map[string]interface{}, map[string]any, error) {
	dataMap := make(map[string]interface{})
	if err := gsgin.GinPostBody(c, &dataMap); err != nil {
		return nil, nil, err
	}
	sshId := dataMap[`ssh_id`]
	if cast.ToInt(sshId) == 0 {
		return nil, nil, errors.New(`缺少ssh_id参数`)
	}
	sshConfig, err := common.DbMain.GetSshConfig(sshId)
	if err != nil {
		return nil, nil, err
	}
	return dataMap, sshConfig, nil
}

func runSSHCommandOnce(sshConfig map[string]any, commandText string) (string, error) {
	return runSSHCommandOnceWithFactory(sshConfig, commandText, nil, nil)
}

func getGitCurrentBranchOnce(sshConfig map[string]any, codePath string) (string, error) {
	return getGitCurrentBranchOnceWithFactory(sshConfig, codePath, nil, nil)
}

func getGitCurrentBranchOnceWithFactory(
	sshConfig map[string]any,
	codePath string,
	factory oneShotCommandRunnerFactory,
	release oneShotCommandRunnerRelease,
) (string, error) {
	command := p_shell.NewCommand()
	command.Init()
	command.Cd(codePath)
	command.GitShowBranch()

	currentBranch, err := runSSHCommandOnceWithFactory(sshConfig, command.GetCommand().ToStr(), factory, release)
	if err != nil {
		return ``, err
	}
	return strings.TrimSpace(CleanBranchName(currentBranch)), nil
}
