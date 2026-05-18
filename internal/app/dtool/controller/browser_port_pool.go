package controller

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// browserPortItem 端口池中的单个端口项
type browserPortItem struct {
	Config     define.McpChromeDevtoolsConfigItem
	InUse      bool
	preWarmCmd *exec.Cmd
}

// browserPortPool 浏览器调试端口池，管理所有端口的预热浏览器进程
type browserPortPool struct {
	mu         sync.Mutex
	items      []*browserPortItem
	executable string
	preWarmDir string
}

var globalBrowserPortPool *browserPortPool

const (
	browserPortPoolPlaywrightInitTimeout = 60  // Playwright 初始化最大等待秒数
	browserPortPoolPortReleaseWaitMs     = 300 // 端口释放后等待 OS 回收的毫秒数
)

// InitBrowserPortPool 初始化端口池，为 tbl_chrome_devtools_config 中所有端口预启动 headless 浏览器。
// 初始化在 goroutine 中执行，轮询等待 Playwright 就绪后读取 DB 并逐端口启动。
func InitBrowserPortPool() {
	for i := 0; i < browserPortPoolPlaywrightInitTimeout; i++ {
		if component.PlaywrightClient != nil && component.PlaywrightClient.Pw != nil {
			break
		}
		time.Sleep(time.Second)
	}
	if component.PlaywrightClient == nil || component.PlaywrightClient.Pw == nil {
		gstool.FmtPrintlnLogTime("[端口池] Playwright 初始化超时，跳过端口预热")
		return
	}

	executablePath := component.PlaywrightClient.Pw.Chromium.ExecutablePath()
	preWarmDir := filepath.Join(component.EnvClient.WebkitDataPath, "prewarm")

	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_chrome_devtools_config ORDER BY id`,
	).All()
	if err != nil {
		gstool.FmtPrintlnLogTime("[端口池] 读取端口配置失败: %v", err)
		return
	}

	if len(rows) == 0 {
		gstool.FmtPrintlnLogTime("[端口池] 无端口配置，跳过预热")
		return
	}

	pool := &browserPortPool{
		items:      make([]*browserPortItem, 0, len(rows)),
		executable: executablePath,
		preWarmDir: preWarmDir,
	}

	for _, row := range rows {
		item := &browserPortItem{
			Config: define.McpChromeDevtoolsConfigItem{
				Id:     cast.ToInt(row["id"]),
				Name:   cast.ToString(row["name"]),
				Port:   cast.ToInt(row["port"]),
				Remark: cast.ToString(row["remark"]),
			},
		}
		if err := pool.startPreWarm(item); err != nil {
			gstool.FmtPrintlnLogTime("[端口池] 预热端口 %d 失败: %v", item.Config.Port, err)
		} else {
			gstool.FmtPrintlnLogTime("[端口池] 端口 %d (%s) 预热成功", item.Config.Port, item.Config.Name)
		}
		pool.items = append(pool.items, item)
	}

	globalBrowserPortPool = pool
	gstool.FmtPrintlnLogTime("[端口池] 初始化完成，共管理 %d 个端口", len(rows))
}

// Acquire 获取一个空闲端口的配置。返回前会杀掉该端口的预热进程以释放端口给 Playwright 使用。
func (p *browserPortPool) Acquire() (*define.McpChromeDevtoolsConfigItem, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, item := range p.items {
		if !item.InUse {
			if item.preWarmCmd != nil && item.preWarmCmd.Process != nil {
				_ = item.preWarmCmd.Process.Kill()
				_ = item.preWarmCmd.Wait()
				item.preWarmCmd = nil
				// 等待 OS 释放端口，避免 Playwright 启动时端口冲突
				time.Sleep(browserPortPoolPortReleaseWaitMs * time.Millisecond)
			}
			item.InUse = true
			gstool.FmtPrintlnLogTime("[端口池] 分配端口 %d (%s)", item.Config.Port, item.Config.Name)
			return &item.Config, nil
		}
	}
	return nil, fmt.Errorf("没有可用的调试端口")
}

// Release 释放端口，重新启动预热浏览器使端口恢复 idle 状态。
func (p *browserPortPool) Release(port int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, item := range p.items {
		if item.Config.Port == port {
			item.InUse = false
			if err := p.startPreWarm(item); err != nil {
				gstool.FmtPrintlnLogTime("[端口池] 重新预热端口 %d 失败: %v", port, err)
			} else {
				gstool.FmtPrintlnLogTime("[端口池] 端口 %d 已释放并重新预热", port)
			}
			return
		}
	}
}

// Shutdown 停止所有预热浏览器进程。
func (p *browserPortPool) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, item := range p.items {
		if item.preWarmCmd != nil && item.preWarmCmd.Process != nil {
			_ = item.preWarmCmd.Process.Kill()
			item.preWarmCmd = nil
		}
	}
	gstool.FmtPrintlnLogTime("[端口池] 已停止所有预热进程")
}

// ShutdownBrowserPortPool 停止全局端口池的所有预热进程（供 config.Stop 调用）。
func ShutdownBrowserPortPool() {
	if globalBrowserPortPool != nil {
		globalBrowserPortPool.Shutdown()
	}
}

// acquireBrowserPort 从端口池获取一个空闲端口。
func acquireBrowserPort() (*define.McpChromeDevtoolsConfigItem, error) {
	if globalBrowserPortPool != nil {
		return globalBrowserPortPool.Acquire()
	}
	return nil, fmt.Errorf("端口池未初始化，无法分配端口")
}

// releaseBrowserPort 释放端口回池。
func releaseBrowserPort(port int) {
	if globalBrowserPortPool != nil {
		globalBrowserPortPool.Release(port)
	}
}

// startPreWarm 为指定端口启动一个 headless Chromium 预热进程。
func (p *browserPortPool) startPreWarm(item *browserPortItem) error {
	if item.preWarmCmd != nil && item.preWarmCmd.Process != nil {
		_ = item.preWarmCmd.Process.Kill()
		item.preWarmCmd = nil
	}

	userDataDir := filepath.Join(p.preWarmDir, fmt.Sprintf("port_%d", item.Config.Port))
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return fmt.Errorf("创建预热目录失败: %w", err)
	}

	item.preWarmCmd = exec.Command(p.executable,
		fmt.Sprintf("--remote-debugging-port=%d", item.Config.Port),
		fmt.Sprintf("--user-data-dir=%s", userDataDir),
		"--no-first-run",
		"--no-default-browser-check",
		"--headless=new",
		"about:blank",
	)

	if err := item.preWarmCmd.Start(); err != nil {
		return fmt.Errorf("启动浏览器失败: %w", err)
	}
	return nil
}
