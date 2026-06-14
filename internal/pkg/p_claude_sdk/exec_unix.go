//go:build !windows

package p_claude_sdk

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"
	"syscall"
)

// startSdkClaude Unix 实现：启动 claude CLI 子进程。
// 利用 Unix 进程组管理，确保进程退出时自动终止所有子进程。
func startSdkClaude(ctx context.Context, args []string, workDir string, env []string, stdin io.Reader) (sdkProcessResult, error) {
	cmd := exec.Command(ClaudeCLIExecutable, args...)
	cmd.Dir = workDir
	cmd.Env = env
	cmd.Stdin = stdin
	// 设置进程组，便于后续统一 kill
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return sdkProcessResult{}, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return sdkProcessResult{}, err
	}

	if err := cmd.Start(); err != nil {
		return sdkProcessResult{}, err
	}

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
			log.Printf("[sdk-exec] stderr: %s", text)
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
			exitCode = ee.ExitCode()
			return
		}
		exitCode = 1
		waitErr = err
	}()

	// 进程组 ID：Setpgid=true 时 pgid==pid，因此直接使用 pid 作为进程组 ID
	// （若 Getpgid 失败 pgid 仍为 pid，对于新进程组语义正确）
	pgid := cmd.Process.Pid
	if cmd.SysProcAttr != nil && cmd.SysProcAttr.Setpgid {
		if gid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
			pgid = gid
		}
	}

	return sdkProcessResult{
		lineCh:   lineCh,
		stderrCh: stderrCh,
		pid:      cmd.Process.Pid,
		waitFn: func() (int, error) {
			<-waitDone
			return exitCode, waitErr
		},
		closeFn: func() {
			// kill 整个进程组（只有当 Setpgid 为 true 时使用 -pgid，否则仅 kill 主进程）
			if cmd.Process != nil {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
		},
	}, nil
}
