package common

import (
	"os"
	"path/filepath"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

const (
	// defaultDToolDirName 表示用户家目录下的 dtool 默认目录。 // Represents the default dtool directory under the user's home directory.
	defaultDToolDirName = `.dtool`
)

// ResolveDefaultDToolDir 解析 dtool 默认目录，优先使用显式配置，否则回落到 ~/.dtool。 // Resolves the dtool directory, preferring explicit config and falling back to ~/.dtool.
func ResolveDefaultDToolDir(configValue string) string {
	trimmedValue := strings.TrimSpace(configValue)
	if trimmedValue != `` {
		return trimmedValue
	}

	homeDir, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(homeDir) == `` {
		// 无法获取家目录时退回当前工作目录下的 .dtool，避免返回空路径。 // Fall back to a workspace-local .dtool path when the home directory is unavailable.
		return defaultDToolDirName
	}
	return filepath.Join(homeDir, defaultDToolDirName)
}

// ResolvePlaywrightPath 解析 Playwright 目录配置，空值时回落到 ~/.dtool。 // Resolves Playwright path config and falls back to ~/.dtool when unset.
func ResolvePlaywrightPath(configValue, defaultLeafDir, drive string) string {
	trimmedValue := strings.TrimSpace(configValue)
	if trimmedValue != `` {
		// 仅在显式配置时替换盘符占位符，保持现有 {DRIVE} 语义。 // Replace the drive placeholder only for explicit config values to preserve existing semantics.
		return gstool.SReplaces(trimmedValue, map[string]string{
			`{DRIVE}`: drive,
		})
	}

	homeDir, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(homeDir) == `` {
		// 无法获取家目录时退回当前工作目录下的 .dtool，避免返回空路径。 // Fall back to the current workspace-local .dtool path when the home directory is unavailable.
		return filepath.Join(defaultDToolDirName, defaultLeafDir)
	}
	return filepath.Join(homeDir, defaultDToolDirName, defaultLeafDir)
}
