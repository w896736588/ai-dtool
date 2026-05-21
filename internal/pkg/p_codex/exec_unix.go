//go:build !windows

package p_codex

import (
	"bufio"
	"context"
	"log"
	"os/exec"
	"syscall"
)

// startCodex Unix 实现。
// 使用 Setsid + Setpgid 创建独立进程组，关闭时通过信号杀死整个进程组，
// 确保子进程一并终止。
func startCodex(ctx context.Context, args []string, workDir string, env []string) (ptyResult, error) {
	cmd := exec.Command(`codex`, args...)
	cmd.Dir = workDir
	cmd.Env = env
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

	pgid := -cmd.Process.Pid

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

	// 实时读取 stderr
	go func() {
		defer close(stderrCh)
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			text := scanner.Text()
			log.Printf("[codex-exec] stderr: %s", text)
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
			_ = syscall.Kill(pgid, syscall.SIGKILL)
			_ = cmd.Process.Kill()
		},
	}, nil
}
