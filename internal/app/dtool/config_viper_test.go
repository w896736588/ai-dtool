package dtool

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfigViperReadsINI(t *testing.T) {
	t.Parallel()

	cfgDir := filepath.Join(t.TempDir(), "config", AppName)
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	cfgPath := filepath.Join(cfgDir, "company.ini")
	cfgBody := []byte("[base]\ndbFileName=frog.db\ndbIsGitRepo=true\nmemoryDbPath=D:/repo/memory\nmemoryDbFileName=memory.db\nmemoryDbIsGitRepo=true\nmemoryDbAutoPushDelayMinutes=7\n")
	if err := os.WriteFile(cfgPath, cfgBody, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	v := newConfigViper()
	v.AddConfigPath(cfgDir)
	v.SetConfigName("company")
	v.SetConfigType("ini")

	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("read ini config: %v", err)
	}

	if got := v.GetString("base.dbFileName"); got != "frog.db" {
		t.Fatalf("dbFileName = %q, want %q", got, "frog.db")
	}
	if got := v.GetBool("base.dbIsGitRepo"); !got {
		t.Fatalf("dbIsGitRepo = %v, want true", got)
	}
	if got := v.GetString("base.memoryDbPath"); got != "D:/repo/memory" {
		t.Fatalf("memoryDbPath = %q, want %q", got, "D:/repo/memory")
	}
	if got := v.GetString("base.memoryDbFileName"); got != "memory.db" {
		t.Fatalf("memoryDbFileName = %q, want %q", got, "memory.db")
	}
	if got := v.GetBool("base.memoryDbIsGitRepo"); !got {
		t.Fatalf("memoryDbIsGitRepo = %v, want true", got)
	}
	if got := v.GetInt("base.memoryDbAutoPushDelayMinutes"); got != 7 {
		t.Fatalf("memoryDbAutoPushDelayMinutes = %d, want %d", got, 7)
	}
}

func TestFormatEnvSummary(t *testing.T) {
	t.Parallel()

	env := &define.Env{
		AppName:            "dtool",
		RootPath:           `C:\work\frog\dev_tool_master`,
		ConfigFile:         "company",
		ConfigPath:         `C:\work\frog\dev_tool_master\config\dtool`,
		LogPath:            `C:\work\frog\dev_tool_master\logs`,
		NodePath:           "node",
		WebkitDriverPath:   `C:/devtool/webkit_driver`,
		WebkitDownloadPath: `C:/devtool/webkit_download`,
		WebkitDataPath:     `C:/devtool/webkit_data`,
		Crawl4AIBaseURL:    "http://127.0.0.1:11235",
		Crawl4AIDataPath:   `C:\work\frog\dev_tool_master\upload\crawl4ai`,
		Crawl4AIScriptPath: `C:\work\frog\dev_tool_master\script\crawl4ai_service.py`,
		DbConfig: &define.DbConfig{
			DbName: "frog.db",
			DbPath: `C:\work\frog\dev_tool_db\zhima`,
		},
		LogDbConfig: &define.DbConfig{
			DbName: "frog.log.db",
			DbPath: `C:\work\frog\dev_tool_db\zhima`,
		},
		WebConfig: &define.WebConfig{
			WebPath: `C:\work\frog\dev_tool_master\web\dist`,
		},
	}

	got := formatEnvSummary(env)

	wantContains := []string{
		"配置摘要",
		"[基础]",
		"应用: dtool",
		"根目录: C:\\work\\frog\\dev_tool_master",
		"[数据库]",
		"完整路径: C:\\work\\frog\\dev_tool_db\\zhima\\frog.db",
		"log库完整路径: C:\\work\\frog\\dev_tool_db\\zhima\\frog.log.db",
		"[Web]",
		"目录: C:\\work\\frog\\dev_tool_master\\web\\dist",
		"[Playwright]",
		"Node: node",
		"[Crawl4AI]",
		"地址: http://127.0.0.1:11235",
		"[日志]",
		"目录: C:\\work\\frog\\dev_tool_master\\logs",
	}
	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Fatalf("summary missing %q\nfull summary:\n%s", want, got)
		}
	}

	unwanted := []string{
		"PkgPath",
		"DatabaseUpPath",
		"MemoryDatabaseUpPath",
		"Ports:",
		"PythonCommand:",
	}
	for _, s := range unwanted {
		if strings.Contains(got, s) {
			t.Fatalf("summary should not contain %q\nfull summary:\n%s", s, got)
		}
	}
}

func TestBuildLogDBName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		mainDBName string
		want       string
	}{
		{
			name:       "带后缀主库名",
			mainDBName: "dtool.db",
			want:       "dtool.log.db",
		},
		{
			name:       "不带后缀主库名",
			mainDBName: "dtool",
			want:       "dtool.log.db",
		},
		{
			name:       "多段后缀主库名",
			mainDBName: "dtool.test.db",
			want:       "dtool.test.log.db",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := buildLogDBName(testCase.mainDBName)
			if got != testCase.want {
				t.Fatalf("buildLogDBName() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestInitEnvFallsBackToHomeDToolPathsWhenPlaywrightPathsUnset(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	oldEnv := component.EnvClient
	t.Cleanup(func() {
		component.EnvClient = oldEnv
	})

	rootDir := t.TempDir()
	cfgDir := filepath.Join(rootDir, "config", AppName)
	if err = os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	cfgPath := filepath.Join(cfgDir, "config.ini")
	cfgBody := []byte("[run]\nports=17170\n[path]\nwebkit_driver_path=\nwebkit_data_path=\nwebkit_download_path=\n[base]\ndbFileName=frog.db\ndbPath=\n")
	if err = os.WriteFile(cfgPath, cfgBody, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	component.EnvClient = &define.Env{
		RootPath: rootDir,
	}

	v := newConfigViper()
	InitEnv(AppName, "config", v)

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

func TestInitEnvFallsBackToHomeDToolDirWhenDbPathUnset(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}

	oldEnv := component.EnvClient
	t.Cleanup(func() {
		component.EnvClient = oldEnv
	})

	rootDir := t.TempDir()
	cfgDir := filepath.Join(rootDir, "config", AppName)
	if err = os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	cfgPath := filepath.Join(cfgDir, "config.ini")
	cfgBody := []byte("[run]\nports=17170\n[base]\ndbPath=\ndbFileName=\n")
	if err = os.WriteFile(cfgPath, cfgBody, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	component.EnvClient = &define.Env{
		RootPath: rootDir,
	}

	v := newConfigViper()
	InitEnv(AppName, "config", v)

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
