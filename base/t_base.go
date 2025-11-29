package base

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/pion/stun"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TBase struct {
	StartMillUnix int64
}

// GetCombineKey 组装key
func (h *TBase) GetCombineKey(params ...any) string {
	strParamsList := gstool.Array2Str(&params)
	return strings.Join(strParamsList, `#`)
}

// ExplainCombineKey 分解key
func (h *TBase) ExplainCombineKey(uniqueKey string) []string {
	return strings.Split(uniqueKey, `#`)
}

func (h *TBase) GetUnique(prefix string) string {
	h.StartMillUnix += 1
	return fmt.Sprintf(`%s%d`, prefix, h.StartMillUnix)
}

func (h *TBase) GetPublicIPWithSTUN() (string, error) {
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

func (h *TBase) DiffText(text1, text2 string) string {
	// 创建临时文件
	tmpFileName1 := Component.TBase.GetUnique(`diff`) + `.md`
	tmpFile1 := filepath.Join(Component.Env.LogPath, tmpFileName1)

	defer func() {
		_ = gstool.FileDelete(tmpFile1)
	}()

	err := gstool.FileCreate(Component.Env.LogPath, tmpFileName1, text1)
	if err != nil {
		return ``
	}

	//defer gstool.FileDelete(tmpFile1)
	tmpFileName2 := Component.TBase.GetUnique(`diff`) + `.md`
	tmpFile2 := filepath.Join(Component.Env.LogPath, tmpFileName2)

	defer func() {
		_ = gstool.FileDelete(tmpFile2)
	}()

	err = gstool.FileCreate(Component.Env.LogPath, tmpFileName2, text2)
	if err != nil {
		return ``
	}

	// 执行 git diff --numstat file1 file2
	cmd := exec.Command("git", "diff", "--no-index", "--shortstat", tmpFile1, tmpFile2)
	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")
	stats := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			stats = lines[i]
			break
		}
	}
	return stats

}
