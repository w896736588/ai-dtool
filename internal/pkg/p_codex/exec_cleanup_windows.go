//go:build windows

package p_codex

import (
	"log"
	"os/exec"
)

// cleanupOrphanedCodexProcesses Windows 实现。
// 通过 PowerShell 查找命令行中包含 codex 的残留进程并终止。
// 用于 Go 进程崩溃重启后清理残留孤儿进程。
func cleanupOrphanedCodexProcesses() {
	//nolint:gosec // 清理孤儿进程的命令，参数由程序内部构造
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-CimInstance Win32_Process -Filter "name='codex.exe'" | Where-Object { $_.CommandLine -like '*codex exec*' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue }`,
	)
	if err := cmd.Run(); err != nil {
		log.Printf("[codex-exec] 清理残留 Codex 进程失败（忽略继续）: %v", err)
	}
}
