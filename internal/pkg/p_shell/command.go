package p_shell

import (
	"fmt"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

type Command struct {
	commandList []string
	sudo        string //sudo前缀
}

func NewCommand() *Command {
	command := Command{}
	command.Init()
	return &command
}

func (h *Command) Init() *Command {
	h.commandList = make([]string, 0)
	h.sudo = ``
	return h
}

func (h *Command) SetCommand(command string) *Command {
	h.commandList = append(h.commandList, command)
	return h
}

func (h *Command) Sudo() *Command {
	h.sudo = `sudo `
	return h
}

func (h *Command) GetCommand() *gstool.GsCons {
	gstool.ArrayFilterEmpty(&h.commandList)
	command := gstool.ConsNew(strings.Join(h.commandList, `;`))
	return command
}

func (h *Command) GitShowBranch() *Command {
	h.SetCommand(h.sudo + `git symbolic-ref --short -q HEAD`)
	return h
}

func (h *Command) Echo(msg string) *Command {
	h.SetCommand(`echo ` + msg)
	return h
}

func (h *Command) GitShowOriginBranch() *Command {
	h.SetCommand(h.sudo + `git ls-remote --heads origin "$(` + h.sudo + ` git symbolic-ref --short -q HEAD)"`)
	return h
}

// GitShowAllOriginBranches 查询远程所有分支
func (h *Command) GitShowAllOriginBranches() *Command {
	h.SetCommand(h.sudo + `git ls-remote --heads origin`)
	return h
}

func (h *Command) GitIgnoreAll() *Command {
	h.SetCommand(h.sudo + `git checkout .`)
	return h
}

// RemoteOriginBranch 注意这个命令会让get fetch失效，仅用于那些非常大的仓库
func (h *Command) RemoteOriginBranch(branch string) *Command {
	branch = gstool.SReplaces(branch, map[string]string{
		` `: ``,
	})
	if branch == `master` {
		return h
	}
	h.SetCommand(h.sudo + ` git remote set-branches origin '` + branch + `'`)
	return h
}

func (h *Command) GitCleanAll() *Command {
	h.SetCommand(h.sudo + `git clean -df`)
	return h
}

func (h *Command) GitFetch() *Command {
	h.SetCommand(h.sudo + `git fetch`)
	return h
}

func (h *Command) Cd(dir string) *Command {
	h.SetCommand(fmt.Sprintf(`cd %s`, dir))
	return h
}

func (h *Command) GitCheckout(branch string) *Command {
	h.SetCommand(fmt.Sprintf(`%sgit checkout %s`, h.sudo, branch))
	return h
}

// GitCheckoutNewBranch 基于当前分支创建并切换新分支
func (h *Command) GitCheckoutNewBranch(branch string) *Command {
	h.SetCommand(fmt.Sprintf(`%sgit checkout -b %s`, h.sudo, branch))
	return h
}

func (h *Command) GitPullOrigin(branch string) *Command {
	h.SetCommand(fmt.Sprintf(`%sgit pull --quiet origin %s`, h.sudo, branch))
	return h
}

// GitPullOriginCurrentBranch 拉取当前本地分支对应的远程分支（运行时动态解析）。
func (h *Command) GitPullOriginCurrentBranch() *Command {
	h.SetCommand(fmt.Sprintf(`%sgit pull --quiet origin "$(%sgit symbolic-ref --short -q HEAD)"`, h.sudo, h.sudo))
	return h
}

// GitPushOriginSetUpstream 推送并建立上游跟踪
func (h *Command) GitPushOriginSetUpstream(branch string) *Command {
	h.SetCommand(fmt.Sprintf(`%sgit push -u origin %s`, h.sudo, branch))
	return h
}

// GitCleanCredentialCache 清除认证信息 防止缓存上一次认证的错误
func (h *Command) GitCleanCredentialCache() *Command {
	h.SetCommand(fmt.Sprintf(`%sgit credential-cache exit; unset GIT_ASKPASS`, h.sudo))
	return h
}

func (h *Command) GitPull() *Command {
	h.SetCommand(fmt.Sprintf(`%sgit pull  --quiet`, h.sudo))
	return h
}

func (h *Command) GitStatus() *Command {
	h.SetCommand(fmt.Sprintf(`%sgit status `, h.sudo))
	return h
}

func (h *Command) GitCommitLog() *Command {
	h.SetCommand(fmt.Sprintf(`%sgit log --oneline -n 20 `, h.sudo))
	return h
}

func (h *Command) WechatKefuStatus(appid string) *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker ps |awk '{print $NF}'|grep -v NAMES|xargs -I {} bash -c " echo {} && sudo docker exec {} ps -ef | grep -i %s " `, h.sudo, appid))
	return h
}

func (h *Command) WechatKefuProcess(dockerName, appid string) *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker exec %s ps -ef | grep -i %s`, h.sudo, dockerName, appid))
	return h
}

func (h *Command) Kill9(processName string) *Command {
	h.SetCommand(fmt.Sprintf(`%s kill -9 $(ps aux | grep "%s" | grep -v grep | awk '{print $2}')`, h.Sudo(), processName))
	return h
}

func (h *Command) DockerSearchPidList(dockerName, processName string) *Command {
	h.SetCommand(fmt.Sprintf(`%s docker exec %s sh -c "ps aux | grep %s | grep -v grep  | awk '{print "\n" \$2}'"`, h.sudo, dockerName, processName))
	return h
}

func (h *Command) DockerKill9(dockerName, processName string) *Command {
	h.SetCommand(fmt.Sprintf(`%s docker exec %s sh -c "ps aux | grep %s | grep -v grep  | awk '{print \$2}' | xargs kill"`, h.sudo, dockerName, processName))
	return h
}

func (h *Command) DockerNameList() *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker ps |awk '{print $NF}'`, h.sudo))
	return h
}

func (h *Command) DockerExecKill(dockerName, pid string) *Command {
	h.SetCommand(fmt.Sprintf(`%ssudo docker exec %s kill -9 %s`, h.sudo, dockerName, pid))
	return h
}

func (h *Command) DockerExecPhpWechatKefu(dockerName, codePath, commandName string) *Command {
	h.SetCommand(fmt.Sprintf(`%s docker exec %s php /var/www/%s/scan/protected/yiic OpenPushWechatKefuOpen %s & `, h.sudo, dockerName, codePath, commandName))
	return h
}

func (h *Command) DockerExecConsumerRestartAll(dockerName string) *Command {
	h.SetCommand(fmt.Sprintf(`%s docker exec %s supervisorctl restart all`, h.sudo, dockerName))
	return h
}

func (h *Command) ConsumerRestartAll() *Command {
	h.SetCommand(fmt.Sprintf(`%ssupervisorctl restart all`, h.sudo))
	return h
}

func (h *Command) DockerExecConsumerStopAll(dockerName string) *Command {
	h.SetCommand(fmt.Sprintf(`%s docker exec %s supervisorctl stop all`, h.sudo, dockerName))
	return h
}

func (h *Command) ConsumerStopAll() *Command {
	h.SetCommand(fmt.Sprintf(`%ssupervisorctl stop all`, h.sudo))
	return h
}

func (h *Command) ConsumerRestart(dockerName, consumerName string) *Command {
	if dockerName == `` {
		h.SetCommand(fmt.Sprintf(`%ssupervisorctl restart %s:`, h.sudo, consumerName))
	} else {
		h.SetCommand(fmt.Sprintf(`%sdocker exec %s supervisorctl restart %s:`, h.sudo, dockerName, consumerName))
	}
	return h
}

func (h *Command) ConsumerStop(dockerName, consumerName string) *Command {
	if dockerName == `` {
		h.SetCommand(fmt.Sprintf(`%s supervisorctl stop %s:`, h.sudo, consumerName))
	} else {
		h.SetCommand(fmt.Sprintf(`%s docker exec %s supervisorctl stop %s:`, h.sudo, dockerName, consumerName))
	}

	return h
}

func (h *Command) ConsumerStatus() *Command {
	h.SetCommand(fmt.Sprintf(`%ssupervisorctl status`, h.sudo))
	return h
}

func (h *Command) ConsumerStatusGrep(dockerName, consumerName string) *Command {
	if dockerName == `` {
		h.SetCommand(fmt.Sprintf(`%ssupervisorctl status | grep -i -E '%s"|%s "'`, h.sudo, consumerName, consumerName))
	} else {
		h.SetCommand(fmt.Sprintf(`%sdocker exec %s supervisorctl status | grep -i -E '%s:|%s '`, h.sudo, dockerName, consumerName, consumerName))
	}
	return h
}

func (h *Command) DockerExecConsumerStatus(dockerName string) *Command {
	h.SetCommand(fmt.Sprintf(`%s docker exec %s supervisorctl status`, h.sudo, dockerName))
	return h
}

func (h *Command) Cat(fileName string) *Command {
	h.SetCommand(fmt.Sprintf(`%s cat %s`, h.sudo, fileName))
	return h
}

func (h *Command) Append(fileName, content string) *Command {
	h.SetCommand(fmt.Sprintf(`%s printf '%s' >> %s`, h.sudo, content, fileName))
	return h
}

func (h *Command) ConsumerConfigList(dockerName string) *Command {
	// 用 for 循环替代 awk system()：避免 PTY 模式下子进程输出时序不稳定导致结果丢失
	listCmd := `for f in *.conf; do [ -f "$f" ] && echo "$f---$(head -n 1 "$f")"; done`
	if dockerName == `` {
		h.SetCommand(fmt.Sprintf(`%scd /etc/supervisor/conf.d/; %s`, h.sudo, listCmd))
	} else {
		h.SetCommand(fmt.Sprintf(`%scd /var/www/dockerfiles/dev_test/docker_volumes/supervisor/etc/supervisor/conf.d; %s`, h.sudo, listCmd))
	}
	return h
}

func (h *Command) DockerRestart() *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker-compose restart`, h.sudo))
	return h
}

func (h *Command) DockerExec(dockerName, dockerCommand string) *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker exec %s %s`, h.sudo, dockerName, dockerCommand))
	return h
}

func (h *Command) FindGitDir(dirPath string, depth int) *Command {
	h.SetCommand(fmt.Sprintf(`%s find %s -maxdepth %d -type d -exec sh -c '  
    for dir; do  
        if [ -d "$dir/.git" ]; then  
            echo "$dir"  
        fi  
    done  
' sh {} +`, h.sudo, dirPath, depth))
	return h
}

func (h *Command) DockerComposeServices(dockerCmd, envFile string) *Command {
	h.SetCommand(fmt.Sprintf(`%s %s %s config --services`, h.sudo, dockerCmd, h.getEnvFileCommand(envFile)))
	return h
}

func (h *Command) getEnvFileCommand(envFile string) string {
	if envFile == `` {
		return ``
	}
	return ` --env-file ` + envFile
}

func (h *Command) DockerComposeStop(dockerCmd, envFile string) *Command {
	h.SetCommand(fmt.Sprintf(`%s %s %s down`, h.sudo, dockerCmd, h.getEnvFileCommand(envFile)))
	return h
}

func (h *Command) DockerComposeStatus(dockerCmd, envFile string) *Command {
	h.SetCommand(fmt.Sprintf(`%s docker stats $(sudo %s %s ps -q) --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}"`, h.sudo, dockerCmd, h.getEnvFileCommand(envFile)))
	return h
}

func (h *Command) DockerComposeRestart(dockerCmd, envFile string, services []string) *Command {
	h.SetCommand(fmt.Sprintf(`%s %s %s restart %s`, h.sudo, dockerCmd, h.getEnvFileCommand(envFile), strings.Join(services, ` `)))
	return h
}

func (h *Command) DockerComposeStopService(dockerCmd, envFile string, services []string) *Command {
	h.SetCommand(fmt.Sprintf(`%s %s %s stop %s`, h.sudo, dockerCmd, h.getEnvFileCommand(envFile), strings.Join(services, ` `)))
	return h
}

func (h *Command) DockerComposeStart(dockerCmd, envFile string) *Command {
	h.SetCommand(fmt.Sprintf(`%s %s %s up -d`, h.sudo, dockerCmd, h.getEnvFileCommand(envFile)))
	return h
}

func (h *Command) DockerComposeUpd(dockerCmd, envFile, service string) *Command {
	h.SetCommand(fmt.Sprintf(`%s %s %s up -d %s`, h.sudo, dockerCmd, h.getEnvFileCommand(envFile), service))
	return h
}

func (h *Command) DockerComposeConfig(dockerCmd, envFile string) *Command {
	h.SetCommand(fmt.Sprintf(`%s %s %s config`, h.sudo, dockerCmd, h.getEnvFileCommand(envFile)))
	return h
}

func (h *Command) DockerImageList() *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker image ls --format "{{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.CreatedSince}}\t{{.Size}}"`, h.sudo))
	return h
}

func (h *Command) DockerImageContainers(imageRef string) *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker ps -a --filter "ancestor=%s" --format "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}"`, h.sudo, imageRef))
	return h
}

func (h *Command) DockerImageRemove(imageRef string) *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker image rm %s`, h.sudo, imageRef))
	return h
}

func (h *Command) DockerContainerStop(containerID string) *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker stop %s`, h.sudo, containerID))
	return h
}

func (h *Command) DockerContainerRemove(containerID string) *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker rm %s`, h.sudo, containerID))
	return h
}

// DockerContainerLogTruncate 清空 Docker json 日志文件。
func (h *Command) DockerContainerLogTruncate() *Command {
	h.SetCommand(fmt.Sprintf(`%sbash -lc 'shopt -s nullglob; logs=(/var/lib/docker/containers/*/*-json.log); if [ ${#logs[@]} -eq 0 ]; then exit 0; fi; truncate -s 0 /var/lib/docker/containers/*/*-json.log'`, h.sudo))
	return h
}

func (h *Command) GitSetSafe(codeDir string) *Command {
	h.SetCommand(fmt.Sprintf(`%sgit config --global --add safe.directory %s `, h.sudo, codeDir))
	return h
}
