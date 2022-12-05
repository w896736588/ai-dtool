package base

import (
	"fmt"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"golang.org/x/crypto/ssh"
	"net"
	"redis_manager/define"
	"time"
)

var RedisRunList map[string]*redis.Client

func InitRedis() {
	go func() {
		RedisRunList = make(map[string]*redis.Client)
		for _, redisConfig := range *RedisList {
			redisRun, err := getRedisClient(&redisConfig)
			if err != nil {
				log.Error("普通 redis %#v 初始化 ping 异常", redisConfig)
				continue
			}
			RedisRunList[redisConfig.UniKey] = redisRun
		}
	}()

}

//获取普通链接的redis
func getRedisClient(redisConfig *define.RedisConfig) (*redis.Client, error) {
	redisCli := redis.NewClient(&redis.Options{
		Network:     "tcp", // 连接方式，默认使用tcp，可省略
		DialTimeout: time.Second * 30,
		Addr:        redisConfig.Host,
		Password:    redisConfig.Password,
		DB:          0,
		PoolSize:    cast.ToInt(redisConfig.PoolSize),
	})

	if err := redisCli.Ping().Err(); nil != err {
		log.Printf("connect to redis err: %v\n", err)
		return nil, err
	}

	return redisCli, nil
}

func getSshRedisClient(redisConfig *define.RedisConfig) (*redis.Client, error) {
	log.Printf(`redisConfig:%#v`, redisConfig)
	cli, err := getSSHClient(redisConfig.SshUser, redisConfig.SshPassword, fmt.Sprintf(`%s:%s`, redisConfig.SshHost, redisConfig.SshPort))
	if nil != err {
		log.Printf("get ssh client err: %v\n", err)
		return nil, err
	}

	redisCli := redis.NewClient(&redis.Options{
		Network:     "tcp", // 连接方式，默认使用tcp，可省略
		DialTimeout: time.Second * 30,
		Addr:        redisConfig.Host,
		Password:    redisConfig.Password,
		DB:          0,
		PoolSize:    cast.ToInt(redisConfig.PoolSize),
		Dialer: func() (conn net.Conn, e error) {
			log.Printf(`Dialer %s:%s`, redisConfig.SshHost, redisConfig.SshPort)
			return cli.Dial(`tcp`, fmt.Sprintf(`%s:%s`, redisConfig.SshHost, redisConfig.SshPort))
		},
	})

	if err = redisCli.Ping().Err(); nil != err {
		log.Printf("connect to redis err: %v\n", err)
		return nil, err
	}

	return redisCli, nil
}

func getSSHClient(user, pass, addr string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		//需要验证服务端，不做验证返回nil，没有该参数会报错
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	sshConn, err := net.Dial("tcp", addr)
	if nil != err {
		fmt.Println("net dial err: ", err)
		return nil, err
	}

	clientConn, chans, reqs, err := ssh.NewClientConn(sshConn, addr, config)
	if nil != err {
		sshConn.Close()
		fmt.Println("ssh client conn err: ", err)
		return nil, err
	}

	client := ssh.NewClient(clientConn, chans, reqs)

	return client, nil
}
