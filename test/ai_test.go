package test

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

// TestBailian 百炼 qwen2.5-coder-3b-instruct 模型
func TestBailian(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	bat := ` mingw32-make -v && cd C:\work\zkzf\zk_message_bus_service\app && mingw32-make linux_all`
	cmd := exec.Command(`cmd.exe`, `/C`, bat)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	if err != nil {
		fmt.Println("make 执行失败:", err)
	}
	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()
	fmt.Println("make 执行结果:" + stdoutStr + `   ` + stderrStr)
}
