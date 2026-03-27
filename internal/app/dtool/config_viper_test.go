package dtool

import (
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
	cfgBody := []byte("[base]\ndbFileName=frog.db\n")
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
