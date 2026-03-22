//go:build windows

package controller

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func resolveManagedProcessStartConfig(config managedProcessConfig) (managedProcessConfig, error) {
	resolved := config

	lookPath, err := exec.LookPath(config.Executable)
	if err != nil {
		return resolved, nil
	}
	resolved.Executable = lookPath

	// 优先直达真实 exe，避免 npm 生成的 .cmd shim 拉起 cmd.exe / Prefer the real exe to avoid npm .cmd shims opening cmd.exe.
	candidateList := []string{
		lookPath + `.exe`,
		filepath.Join(filepath.Dir(lookPath), `node_modules`, normalizeManagedExecutable(config.Executable), `bin`, normalizeManagedExecutable(config.Executable)+`.exe`),
	}
	for _, candidate := range candidateList {
		if candidate == `` {
			continue
		}
		if info, statErr := os.Stat(candidate); statErr == nil && !info.IsDir() {
			resolved.Executable = candidate
			return resolved, nil
		}
	}

	ext := strings.ToLower(filepath.Ext(lookPath))
	if ext == `.cmd` || ext == `.bat` {
		return resolved, nil
	}
	return resolved, nil
}
