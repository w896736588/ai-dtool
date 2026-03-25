package crawl4ai

import (
	"bytes"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

const (
	crawl4AIDockerPullCommand = `docker pull unclecode/crawl4ai:latest`
	crawl4AIDockerRunCommand  = `docker run -d --name crawl4ai -p 11235:11235 --shm-size=2g --restart always unclecode/crawl4ai:latest`
	crawl4AIPlaygroundURL     = `http://localhost:11235/playground/`
)

// CrawlResult 表示单个网址的抓取结果。
type CrawlResult struct {
	URL      string `json:"url"`
	Success  bool   `json:"success"`
	Markdown string `json:"markdown"`
	Title    string `json:"title"`
	Error    string `json:"error"`
}

// InstallGuide 表示页面展示用的 Docker 安装与启动指引。
type InstallGuide struct {
	NeedInstall   bool   `json:"need_install"`
	InstallTarget string `json:"install_target"`
	Title         string `json:"title"`
	Tip           string `json:"tip"`
	PullCommand   string `json:"pull_command"`
	RunCommand    string `json:"run_command"`
	DocsURL       string `json:"docs_url"`
	UseWSLTip     string `json:"use_wsl_tip"`
}

// Service 管理 Crawl4AI 远端服务状态与调用。
type Service struct {
	Env        *define.Env
	Log        *gstool.GsSlog
	mu         sync.Mutex
	status     string
	statusText string
	lastError  string
}

// NewService 创建 Crawl4AI 服务管理器。
func NewService(env *define.Env, log *gstool.GsSlog) *Service {
	return &Service{
		Env:        env,
		Log:        log,
		status:     define.Crawl4AIStatusIdle,
		statusText: `等待连接 Crawl4AI 服务`,
	}
}

// EnsureReady 检测 Docker 启动的 Crawl4AI 服务是否可用。
func (h *Service) EnsureReady() error {
	h.mu.Lock()
	h.setStatusLocked(define.Crawl4AIStatusInstalling, `正在检查 Crawl4AI Docker 服务`, ``)
	h.mu.Unlock()

	if h.isServiceRunning() {
		h.mu.Lock()
		h.setStatusLocked(define.Crawl4AIStatusReady, `Crawl4AI 服务已就绪`, ``)
		h.mu.Unlock()
		return nil
	}

	err := errors.New(`未检测到 Crawl4AI 服务，请先按页面中的 Docker 指引完成安装和启动后重试`)
	h.mu.Lock()
	h.setStatusLocked(define.Crawl4AIStatusFailed, `Crawl4AI 服务未启动`, err.Error())
	h.mu.Unlock()
	return err
}

// Stop 保留空实现，Docker 服务由用户自行管理。
func (h *Service) Stop() {
}

// Status 返回 Crawl4AI 当前状态快照。
func (h *Service) Status() map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()
	installGuide := h.buildInstallGuideLocked()
	return map[string]any{
		`status`:        h.status,
		`status_text`:   h.statusText,
		`error_message`: h.lastError,
		`is_ready`:      h.status == define.Crawl4AIStatusReady,
		`is_installing`: h.status == define.Crawl4AIStatusInstalling,
		`need_install`:  installGuide.NeedInstall,
		`install_guide`: installGuide,
	}
}

// CrawlURLs 抓取多个网址并返回 markdown 内容。
func (h *Service) CrawlURLs(urlList []string, timeout time.Duration) ([]CrawlResult, error) {
	if len(urlList) == 0 {
		return nil, errors.New(`未提供可抓取的网址`)
	}
	if err := h.EnsureReady(); err != nil {
		return nil, err
	}

	bodyMap := map[string]any{
		`urls`:                 urlList,
		`cache_mode`:           `bypass`,
		`word_count_threshold`: 1,
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	request, err := http.NewRequest(http.MethodPost, h.Env.Crawl4AIBaseURL+`/crawl`, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	request.Header.Set(`Content-Type`, `application/json`)

	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 300 {
		return nil, errors.New(`Crawl4AI 请求失败: ` + string(responseBody))
	}

	result := struct {
		Data []CrawlResult `json:"data"`
	}{}
	if err = json.Unmarshal(responseBody, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

// ExtractURLs 从文本中提取网址列表。
func (h *Service) ExtractURLs(text string) []string {
	reg := regexp.MustCompile(`https?://[^\s<>"'，。；、]+`)
	matchList := reg.FindAllString(text, -1)
	result := make([]string, 0, len(matchList))
	existMap := make(map[string]bool)
	for _, item := range matchList {
		item = strings.TrimSpace(strings.TrimRight(item, `),.;]}`))
		if item == `` || existMap[item] {
			continue
		}
		existMap[item] = true
		result = append(result, item)
	}
	return result
}

// buildInstallGuideLocked 构建 Docker 安装指引。
func (h *Service) buildInstallGuideLocked() InstallGuide {
	needInstall := h.status != define.Crawl4AIStatusReady
	return InstallGuide{
		NeedInstall:   needInstall,
		InstallTarget: `docker`,
		Title:         `Crawl4AI Docker 安装指引`,
		Tip:           `建议通过 Docker 启动 Crawl4AI 服务，Windows 建议在 WSL 中执行下方命令。服务启动后刷新当前页面再执行信息抓取。`,
		PullCommand:   crawl4AIDockerPullCommand,
		RunCommand:    crawl4AIDockerRunCommand,
		DocsURL:       crawl4AIPlaygroundURL,
		UseWSLTip:     `Windows 环境建议使用 WSL 运行 Docker 命令。`,
	}
}

// isServiceRunning 检测远端 Crawl4AI 服务是否已运行。
func (h *Service) isServiceRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(h.Env.Crawl4AIBaseURL + `/healthz`)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode < 300
}

// setStatusLocked 在已加锁场景下更新状态信息。
func (h *Service) setStatusLocked(status, statusText, lastError string) {
	h.status = status
	h.statusText = statusText
	h.lastError = lastError
}
