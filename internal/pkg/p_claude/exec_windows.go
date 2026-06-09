//go:build windows

package p_claude

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	jobObjectBasicLimitInformation = 2
	jobObjectLimitKillOnJobClose   = 0x00002000
	createBreakawayFromJob         = 0x01000000
	processSetQuota                = 0x0100 // syscall 中未定义该常量
)

// joBasicLimitInfo 对应 Windows JOBOBJECT_BASIC_LIMIT_INFORMATION（64 位，64 字节）。
type joBasicLimitInfo struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	_                       uint32 // 对齐 padding
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	_                       uint32 // 对齐 padding
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

var (
	modKernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObjectW         = modKernel32.NewProc("CreateJobObjectW")
	procSetInformationJobObject  = modKernel32.NewProc("SetInformationJobObject")
	procAssignProcessToJobObject = modKernel32.NewProc("AssignProcessToJobObject")
)

// createKillOnCloseJob 创建 kill-on-close Job Object。
// 当 Job 的最后一个句柄被关闭时（Go 进程退出 / closeFn 主动关闭），
// Windows 内核自动终止 Job 内所有进程及其子进程。
func createKillOnCloseJob() (syscall.Handle, error) {
	jobName, _ := syscall.UTF16PtrFromString(
		fmt.Sprintf(`Local\dtool-claude-%d`, os.Getpid()),
	)
	h, _, err := procCreateJobObjectW.Call(0, uintptr(unsafe.Pointer(jobName)))
	if h == 0 {
		return 0, fmt.Errorf("CreateJobObject 失败: %w", err)
	}
	job := syscall.Handle(h)

	info := joBasicLimitInfo{}
	info.LimitFlags = jobObjectLimitKillOnJobClose

	ret, _, err := procSetInformationJobObject.Call(
		uintptr(job),
		jobObjectBasicLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		unsafe.Sizeof(info),
	)
	if ret == 0 {
		syscall.CloseHandle(job)
		return 0, fmt.Errorf("SetInformationJobObject 失败: %w", err)
	}

	return job, nil
}

// assignProcessToJob 将进程加入 Job Object（其后续子进程自动继承 Job）。
func assignProcessToJob(job syscall.Handle, pid int) error {
	hProc, err := syscall.OpenProcess(
		processSetQuota|syscall.PROCESS_TERMINATE,
		false,
		uint32(pid),
	)
	if err != nil {
		return fmt.Errorf("OpenProcess 失败: %w", err)
	}
	defer syscall.CloseHandle(hProc)

	ret, _, err := procAssignProcessToJobObject.Call(uintptr(job), uintptr(hProc))
	if ret == 0 {
		return fmt.Errorf("AssignProcessToJobObject 失败: %w", err)
	}
	return nil
}

// startClaude Windows 实现。
// 启动时将 Claude CLI 进程加入 kill-on-close Job Object，
// 确保 Go 进程崩溃或退出时 npx chrome-devtools-mcp 等子进程不会残留为孤儿。
func startClaude(ctx context.Context, args []string, workDir string, env []string, stdin io.Reader) (ptyResult, error) {
	job, jobErr := createKillOnCloseJob()
	if jobErr != nil {
		log.Printf("[claude-exec] 创建 Job Object 失败（降级运行，无孤儿进程保护）: %v", jobErr)
	}

	cmd := exec.Command(`claude`, args...)
	cmd.Dir = workDir
	cmd.Env = env
	cmd.Stdin = stdin
	// CREATE_BREAKAWAY_FROM_JOB 允许进程脱离可能存在的父 Job（如终端 Job），
	// 以便将其加入我们自己的 Job Object。
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | createBreakawayFromJob,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		if job != 0 {
			syscall.CloseHandle(job)
		}
		return ptyResult{}, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		if job != 0 {
			syscall.CloseHandle(job)
		}
		return ptyResult{}, err
	}

	if err := cmd.Start(); err != nil {
		if job != 0 {
			syscall.CloseHandle(job)
		}
		return ptyResult{}, err
	}

	// 将 Claude CLI 进程加入 Job，其子进程（npx、MCP server、watchdog）自动继承
	if job != 0 {
		if err := assignProcessToJob(job, cmd.Process.Pid); err != nil {
			log.Printf("[claude-exec] 分配进程到 Job Object 失败: %v", err)
			syscall.CloseHandle(job)
			job = 0
		}
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
			exitCode = ee.ExitCode()
			return
		}
		exitCode = 1
		waitErr = err
	}()

	// 进程退出或 Go 进程崩溃，Windows 内核自动触发 Job kill-on-close
	return ptyResult{
		lineCh:   lineCh,
		stderrCh: stderrCh,
		pid:      cmd.Process.Pid,
		waitFn: func() (int, error) {
			<-waitDone
			return exitCode, waitErr
		},
		closeFn: func() {
			cmd.Process.Kill()
			if job != 0 {
				syscall.CloseHandle(job)
			}
		},
	}, nil
}
