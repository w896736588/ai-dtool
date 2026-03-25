package crawl4ai

import (
	"dev_tool/internal/app/dtool/define"
	"testing"
)

// TestBuildInstallGuideLocked 验证 Docker 安装指引内容。
func TestBuildInstallGuideLocked(t *testing.T) {
	service := &Service{
		Env:    &define.Env{},
		status: define.Crawl4AIStatusFailed,
	}

	guide := service.buildInstallGuideLocked()
	if !guide.NeedInstall {
		t.Fatalf("expected NeedInstall to be true when service is not ready")
	}
	if guide.InstallTarget != "docker" {
		t.Fatalf("expected InstallTarget to be docker, got %s", guide.InstallTarget)
	}
	if guide.PullCommand != "docker pull unclecode/crawl4ai:latest" {
		t.Fatalf("unexpected pull command: %s", guide.PullCommand)
	}
	if guide.RunCommand != "docker run -d --name crawl4ai -p 11235:11235 --shm-size=2g --restart always unclecode/crawl4ai:latest" {
		t.Fatalf("unexpected run command: %s", guide.RunCommand)
	}
	if guide.DocsURL != "http://localhost:11235/playground/" {
		t.Fatalf("unexpected docs url: %s", guide.DocsURL)
	}
}

// TestExtractURLs 验证网址提取逻辑。
func TestExtractURLs(t *testing.T) {
	service := &Service{}
	result := service.ExtractURLs("请抓取 https://example.com 和 https://openai.com。")
	if len(result) != 2 {
		t.Fatalf("expected 2 urls, got %d", len(result))
	}
}
