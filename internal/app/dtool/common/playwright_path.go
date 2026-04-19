package common

import (
	"os"
	"path/filepath"
	"strings"
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

// ResolvePlaywrightPaths 返回 Playwright 所需的三个默认目录，统一存放于 ~/.dtool/{subDir} 下。 // Returns default Playwright directories under ~/.dtool/{subDir}.
func ResolvePlaywrightPaths(subDir string) (driverPath, dataPath, downloadPath string) {
	base := resolveDToolSubDir(subDir)
	return filepath.Join(base, `webkit_driver`),
		filepath.Join(base, `webkit_data`),
		filepath.Join(base, `webkit_download`)
}

// resolveDToolSubDir 解析 ~/.dtool/{subDir} 路径。 // Resolves ~/.dtool/{subDir}.
func resolveDToolSubDir(subDir string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(homeDir) == `` {
		return filepath.Join(defaultDToolDirName, subDir)
	}
	return filepath.Join(homeDir, defaultDToolDirName, subDir)
}
