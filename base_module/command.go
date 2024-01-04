package base_module

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
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
	gstool.ArrayFilterEmptyString(&h.commandList)
	command := gstool.ConsNew(strings.Join(h.commandList, `;`))
	gstool.FmtPrintlnLog(`执行命令 %s`, command.ToStr())
	return command
}

func (h *Command) GitShowBranch() *Command {
	h.SetCommand(h.sudo + `git symbolic-ref --short -q HEAD`)
	return h
}

func (h *Command) GitIgnoreAll() *Command {
	h.SetCommand(h.sudo + `git checkout .`)
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

func (h *Command) GitPullOrigin(branch string) *Command {
	h.SetCommand(fmt.Sprintf(`%sgit pull origin %s`, h.sudo, branch))
	return h
}

func (h *Command) GitPull() *Command {
	h.SetCommand(fmt.Sprintf(`%sgit pull `, h.sudo))
	return h
}

func (h *Command) GitStatus() *Command {
	h.SetCommand(fmt.Sprintf(`%sgit status `, h.sudo))
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
	h.SetCommand(fmt.Sprintf(`%s docker exec %s php /var/www/%s/scan/protected/yiic OpenPushWechatKefuOpen %s &`, h.sudo, dockerName, codePath, commandName))
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
		h.SetCommand(fmt.Sprintf(`%ssupervisorctl status | grep %s`, h.sudo, consumerName))
	} else {
		h.SetCommand(fmt.Sprintf(`%sdocker exec %s supervisorctl status | grep %s`, h.sudo, dockerName, consumerName))
	}
	return h
}

func (h *Command) DockerExecConsumerStatus(dockerName string) *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker exec %s supervisorctl status`, h.sudo, dockerName))
	return h
}

func (h *Command) ConsumerConfigCat(fileName, dockerName string) *Command {
	if dockerName == `` {
		h.SetCommand(fmt.Sprintf(`%scat %s`, h.sudo, fileName))
	} else {
		h.SetCommand(fmt.Sprintf(`%scat /var/www/dockerfiles/dev_test/docker_volumes/supervisor%s`, h.sudo, fileName))
	}

	return h
}

func (h *Command) ConsumerConfigList(dockerName string) *Command {
	if dockerName == `` {
		h.SetCommand(fmt.Sprintf(`%scd /etc/supervisor/conf.d/; ls | grep '\.conf$' | awk '{printf ""$1"---"; system("head -n 1 "$1)}'`, h.sudo))
	} else {
		h.SetCommand(fmt.Sprintf(`%scd /var/www/dockerfiles/dev_test/docker_volumes/supervisor/etc/supervisor/conf.d; ls | grep '\.conf$' | awk '{printf ""$1"---"; system("head -n 1 "$1)}'`, h.sudo))
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

func (h *Command) DockerPs() *Command {
	h.SetCommand(fmt.Sprintf(`%sdocker stats --no-stream`, h.sudo))
	return h
}
