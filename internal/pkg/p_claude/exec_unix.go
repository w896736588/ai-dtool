//go:build !windows

package p_claude

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"
	"syscall"
)

// startClaude Unix 实现。
// 使用 Setsid + Setpgid 创建独立进程组，关闭时通过信号杀死整个进程组，
// 确保 npx chrome-devtools-mcp 等子进程一并终止。
func startClaude(ctx context.Context, args []string, workDir string, env []string, stdin io.Reader) (ptyResult, error) {
	cmd := exec.Command(`claude`, args...)
	cmd.Dir = workDir
	cmd.Env = env
	cmd.Stdin = stdin
	// Setsid：脱离父进程的会话，防止 Ctrl+C 等信号传播到子进程
	// Setpgid：以子进程 PID 创建新进程组，后续可通过 -pgid 杀死整组进程
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setpgid: true,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ptyResult{}, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ptyResult{}, err
	}

	if err := cmd.Start(); err != nil {
		return ptyResult{}, err
	}

	pgid := -cmd.Process.Pid // 进程组 ID = 进程自己的 PID（Setpgid=true）

	lineCh := make(chan string, 256)
	stderrCh := make(chan string, 64)

	// 实时读取 stdout
	go func() {
		defer close(lineCh)
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, maxScanTokenSize), maxScanTokenSize)
		for scanner.Scan() {
			lineCh <- scanner.Text()
		}
	}()

	// 实时读取 stderr，收集内容用于错误定位
	go func() {
		defer close(stderrCh)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			text := scanner.Text()
			log.Printf("[claude-exec] stderr: %s", text)
			stderrCh <- text
		}
	}()

	// 后台等待进程退出
	waitDone := make(chan struct{})
	var exitCode int
	var waitErr error
	go func() {
		defer close(waitDone)
		err := cmd.Wait()
		if err == nil {
			exitCode = 0
			return
		}
		if ee, ok := err.(*exec.ExitError); ok {
			if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
				exitCode = ws.ExitStatus()
				return
			}
			exitCode = 1
			return
		}
		exitCode = 1
		waitErr = err
	}()

	return ptyResult{
		lineCh:   lineCh,
		stderrCh: stderrCh,
		pid:      cmd.Process.Pid,
		waitFn: func() (int, error) {
			<-waitDone
			return exitCode, waitErr
		},
		closeFn: func() {
			// 先向整个进程组发 SIGKILL，CLaude CLI 及其子进程（npx、MCP server）一并被内核终止
			_ = syscall.Kill(pgid, syscall.SIGKILL)
			// 再确保主进程被回收（正常情况下已被上面信号杀死，此调用不会产生额外副作用）
			_ = cmd.Process.Kill()
		},
	}, nil
}
