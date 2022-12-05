package base

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"redis_manager/define"
	"time"
)

// SshClient ssh链接
var SshClient map[string]*ssh.Session

// createSsh 创建ssh连接
// @author frog
// @date 2022-04-11 15:34:36
func createSsh(sshConfig define.SshConfig, cmd string) ([]byte, error) {
	sshHost := sshConfig.Host
	sshUser := sshConfig.Username
	sshPassword := sshConfig.Password
	sshType := sshConfig.SshType
	sshPort := sshConfig.Port
	//创建ssh登陆配置
	log.Debugf(`ssh配置 %#v`, sshConfig)
	config := &ssh.ClientConfig{
		Timeout:         10 * time.Second, //ssh 连接time out 时间一秒钟, 如果ssh验证错误 会在一秒内返回
		User:            sshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以, 但是不够安全
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}

	if sshType == "password" {
		config.Auth = []ssh.AuthMethod{ssh.Password(sshPassword)}
	}
	//dial 获取ssh client
	addr := fmt.Sprintf("%s:%s", sshHost, sshPort)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		panic(fmt.Sprintf(`创建ssh client 失败 %s %#v %s`, addr, config, err.Error()))
	}
	defer sshClient.Close()
	//创建ssh-session
	session, err := sshClient.NewSession()
	if err != nil {
		log.Errorf(`创建ssh session 失败 %s`, err.Error())
	}
	defer session.Close()
	//执行并返回结果
	return session.CombinedOutput(cmd)
}

// Exec 执行命令
// @author frog
// @date 2022-04-11 16:32:51
func Exec(reqBody *define.SshDo, cmd string) string {
	//拿到前缀 配置暂时都塞到redisList里面了
	prefixExec := ``
	sshName := ``
	for _, value := range *RedisList {
		if value.UniKey == reqBody.UniKey {
			prefixExec = value.SshPrefix
			sshName = value.SshName
			break
		}
	}
	cmd = `sudo` + ` ` + prefixExec + ` ` + cmd
	log.Debugf(`命令 %s`, cmd)
	//执行远程命令
	log.Debugf(`sshName %s %#v`, sshName, *SshConfig)
	if sshConfig, ok := (*SshConfig)[sshName]; ok {
		combo, err := createSsh(sshConfig, cmd)
		if err != nil {
			log.Errorf(`远程执行cmd 失败 %s %s`, err.Error(), combo)
			return err.Error()
		} else {
			fmt.Printf(`结果 %#v`, string(combo))
			return string(combo)
		}
	} else {
		return `不存在的ssh链接`
	}
}
