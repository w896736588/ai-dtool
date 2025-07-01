package test

import (
	"fmt"
	"github.com/pion/stun"
	"net"
	"testing"
	"time"
)

// TestBailian 百炼 qwen2.5-coder-3b-instruct 模型
func TestBailian(t *testing.T) {
	ip, err := getPublicIPWithSTUN()
	if err != nil {
		fmt.Println("获取IP失败:", err)
		return
	}
	fmt.Println("STUN获取的公网IP:", ip)
}

func getPublicIPWithSTUN() (string, error) {
	// 1. 创建UDP连接
	conn, err := net.Dial("udp", "stun.l.google.com:19302") // Google公共STUN服务器
	if err != nil {
		return "", fmt.Errorf("创建UDP连接失败: %v", err)
	}
	defer conn.Close()

	// 2. 设置超时
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return "", fmt.Errorf("设置超时失败: %v", err)
	}

	// 3. 创建STUN客户端
	client, err := stun.NewClient(conn)
	if err != nil {
		return "", fmt.Errorf("创建STUN客户端失败: %v", err)
	}
	defer client.Close()

	// 4. 构建STUN请求
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// 5. 处理响应
	var publicIP string
	err = client.Do(message, func(res stun.Event) {
		if res.Error != nil {
			return
		}

		// 解析XOR-MAPPED-ADDRESS属性
		var xorAddr stun.XORMappedAddress
		if err := xorAddr.GetFrom(res.Message); err != nil {
			return
		}
		publicIP = xorAddr.IP.String()
	})

	if err != nil {
		return "", fmt.Errorf("STUN请求失败: %v", err)
	}

	if publicIP == "" {
		return "", fmt.Errorf("未从STUN响应中获取到IP地址")
	}

	return publicIP, nil
}
