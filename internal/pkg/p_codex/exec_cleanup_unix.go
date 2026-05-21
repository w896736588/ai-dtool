//go:build !windows

package p_codex

import (
	"log"
	"os/exec"
)

// cleanupOrphanedCodexProcesses Unix（macOS/Linux）实现。
// 通过 pkill 杀死所有命令行中包含 "codex exec" 的残留进程。
// 用于 Go 进程崩溃重启后清理残留孤儿进程。
func cleanupOrphanedCodexProcesses() {
	//nolint:gosec // 清理孤儿进程的命令，参数由程序内部构造
	cmd := exec.Command("pkill", "-f", "codex exec")
	// pkill 在找不到匹配进程时返回 1，视为正常
	if err := cmd.Run(); err != nil {
		log.Printf("[codex-exec] 清理残留 Codex 进程失败（忽略继续）: %v", err)
	}
}
