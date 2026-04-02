package business

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestReloadEditableRuntimeConfigFallsBackToHomeDToolPaths(t *testing.T) {
	oldViper := component.ConfigViper
	oldEnv := component.EnvClient
	t.Cleanup(func() {
		component.ConfigViper = oldViper
		component.EnvClient = oldEnv
	})

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	component.ConfigViper = viper.New()
	component.ConfigViper.Set(`base.dbFileName`, `frog.db`)
	component.ConfigViper.Set(`base.dbPath`, `C:\repo`)
	component.ConfigViper.Set(`path.webkit_driver_path`, ``)
	component.ConfigViper.Set(`path.webkit_data_path`, ``)
	component.ConfigViper.Set(`path.webkit_download_path`, ``)
	component.EnvClient = &define.Env{
		AppName:     `dtool`,
		RootPath:    t.TempDir(),
		ConfigBase:  &define.Base{},
		WebConfig:   &define.WebConfig{},
		DbConfig:    &define.DbConfig{},
		LogDbConfig: &define.DbConfig{},
	}

	ReloadEditableRuntimeConfig()

	if got := filepath.Clean(component.EnvClient.WebkitDriverPath); got != filepath.Join(homeDir, `.dtool`, `webkit_driver`) {
		t.Fatalf("WebkitDriverPath = %q, want %q", got, filepath.Join(homeDir, `.dtool`, `webkit_driver`))
	}
	if got := filepath.Clean(component.EnvClient.WebkitDataPath); got != filepath.Join(homeDir, `.dtool`, `webkit_data`) {
		t.Fatalf("WebkitDataPath = %q, want %q", got, filepath.Join(homeDir, `.dtool`, `webkit_data`))
	}
	if got := filepath.Clean(component.EnvClient.WebkitDownloadPath); got != filepath.Join(homeDir, `.dtool`, `webkit_download`) {
		t.Fatalf("WebkitDownloadPath = %q, want %q", got, filepath.Join(homeDir, `.dtool`, `webkit_download`))
	}
}

func TestReloadEditableRuntimeConfigFallsBackToHomeDToolDirWhenDbPathUnset(t *testing.T) {
	oldViper := component.ConfigViper
	oldEnv := component.EnvClient
	t.Cleanup(func() {
		component.ConfigViper = oldViper
		component.EnvClient = oldEnv
	})

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	component.ConfigViper = viper.New()
	component.ConfigViper.Set(`base.dbFileName`, ``)
	component.ConfigViper.Set(`base.dbPath`, ``)
	component.ConfigViper.Set(`path.webkit_driver_path`, ``)
	component.ConfigViper.Set(`path.webkit_data_path`, ``)
	component.ConfigViper.Set(`path.webkit_download_path`, ``)
	component.EnvClient = &define.Env{
		AppName:     `dtool`,
		RootPath:    t.TempDir(),
		ConfigBase:  &define.Base{},
		WebConfig:   &define.WebConfig{},
		DbConfig:    &define.DbConfig{},
		LogDbConfig: &define.DbConfig{},
	}

	ReloadEditableRuntimeConfig()

	if got := filepath.Clean(component.EnvClient.DbConfig.DbPath); got != filepath.Join(homeDir, `.dtool`) {
		t.Fatalf("DbPath = %q, want %q", got, filepath.Join(homeDir, `.dtool`))
	}
	if component.EnvClient.DbConfig.DbName != `dtool.db` {
		t.Fatalf("DbName = %q, want %q", component.EnvClient.DbConfig.DbName, `dtool.db`)
	}
	if got := filepath.Clean(component.EnvClient.LogDbConfig.DbPath); got != filepath.Join(homeDir, `.dtool`) {
		t.Fatalf("LogDbPath = %q, want %q", got, filepath.Join(homeDir, `.dtool`))
	}
}
