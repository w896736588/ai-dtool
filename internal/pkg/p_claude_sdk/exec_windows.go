//go:build windows

package p_claude_sdk

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

// =============================================================================
// Windows 平台：复用 p_claude 的 Job Object 机制管理子进程
// =============================================================================
// 将 claude CLI 子进程加入 kill-on-close Job Object，确保 Go 进程崩溃时
// 所有子进程（包括 MCP server、npx 等）被内核自动终止。
//
// 注意：此处直接复制 p_claude 的 Windows 进程管理逻辑，因为：
//   1. p_claude 的 startClaude 函数是未导出的
//   2. SDK 模式后续可能需额外的进程管理配置
// =============================================================================

const (
	obBasicLimitInformation     = 2
	obLimitKillOnJobClose       = 0x00002000
	createFlagsBreakawayFromJob = 0x01000000
	processAccessSetQuota       = 0x0100
)

// obBasicLimitInfo 对应 Windows JOBOBJECT_BASIC_LIMIT_INFORMATION。
type obBasicLimitInfo struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	_                       uint32 // padding
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	_                       uint32 // padding
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

var (
	modKernel32Sdk                  = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObjectWSdk         = modKernel32Sdk.NewProc("CreateJobObjectW")
	procSetInformationJobObjectSdk  = modKernel32Sdk.NewProc("SetInformationJobObject")
	procAssignProcessToJobObjectSdk = modKernel32Sdk.NewProc("AssignProcessToJobObject")
)

// createSdkKillOnCloseJob 创建 kill-on-close Job Object。
func createSdkKillOnCloseJob() (syscall.Handle, error) {
	jobName, _ := syscall.UTF16PtrFromString(
		fmt.Sprintf(`Local\dtool-sdk-claude-%d`, os.Getpid()),
	)
	h, _, err := procCreateJobObjectWSdk.Call(0, uintptr(unsafe.Pointer(jobName)))
	if h == 0 {
		return 0, fmt.Errorf("CreateJobObject 失败: %w", err)
	}
	job := syscall.Handle(h)

	info := obBasicLimitInfo{}
	info.LimitFlags = obLimitKillOnJobClose

	ret, _, err := procSetInformationJobObjectSdk.Call(
		uintptr(job),
		obBasicLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		unsafe.Sizeof(info),
	)
	if ret == 0 {
		syscall.CloseHandle(job)
		return 0, fmt.Errorf("SetInformationJobObject 失败: %w", err)
	}

	return job, nil
}

// assignSdkProcessToJob 将进程加入 Job Object。
func assignSdkProcessToJob(job syscall.Handle, pid int) error {
	hProc, err := syscall.OpenProcess(
		processAccessSetQuota|syscall.PROCESS_TERMINATE,
		false,
		uint32(pid),
	)
	if err != nil {
		return fmt.Errorf("OpenProcess 失败: %w", err)
	}
	defer syscall.CloseHandle(hProc)

	ret, _, err := procAssignProcessToJobObjectSdk.Call(uintptr(job), uintptr(hProc))
	if ret == 0 {
		return fmt.Errorf("AssignProcessToJobObject 失败: %w", err)
	}
	return nil
}

// startSdkClaude Windows 实现：启动 claude CLI 子进程并纳入 Job Object 管理。
func startSdkClaude(ctx context.Context, args []string, workDir string, env []string, stdin io.Reader) (sdkProcessResult, error) {
	job, jobErr := createSdkKillOnCloseJob()
	if jobErr != nil {
		log.Printf("[sdk-exec] 创建 Job Object 失败（降级运行，无孤儿进程保护）: %v", jobErr)
	}

	cmd := exec.Command(ClaudeCLIExecutable, args...)
	cmd.Dir = workDir
	cmd.Env = env
	cmd.Stdin = stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | createFlagsBreakawayFromJob,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		if job != 0 {
			syscall.CloseHandle(job)
		}
		return sdkProcessResult{}, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		if job != 0 {
			syscall.CloseHandle(job)
		}
		return sdkProcessResult{}, err
	}

	if err := cmd.Start(); err != nil {
		if job != 0 {
			syscall.CloseHandle(job)
		}
		return sdkProcessResult{}, err
	}

	// 将主进程加入 Job Object
	if job != 0 {
		if err := assignSdkProcessToJob(job, cmd.Process.Pid); err != nil {
			log.Printf("[sdk-exec] 分配进程到 Job Object 失败: %v", err)
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

	return sdkProcessResult{
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
