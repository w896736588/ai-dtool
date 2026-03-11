package crawl4ai

import (
	"bytes"
	"context"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

// CrawlResult 表示单个网址的抓取结果。
type CrawlResult struct {
	URL      string `json:"url"`
	Success  bool   `json:"success"`
	Markdown string `json:"markdown"`
	Title    string `json:"title"`
	Error    string `json:"error"`
}

// Service 管理 Crawl4AI 的安装、启动与调用。
type Service struct {
	Env         *define.Env
	Log         *gstool.GsSlog
	mu          sync.Mutex
	process     *exec.Cmd
	startedByMe bool
	status      string
	statusText  string
	lastError   string
	installing  bool
}

// NewService 创建 Crawl4AI 服务管理器。
func NewService(env *define.Env, log *gstool.GsSlog) *Service {
	return &Service{
		Env:        env,
		Log:        log,
		status:     define.Crawl4AIStatusIdle,
		statusText: `等待初始化`,
	}
}

// EnsureReady 检测并确保 Crawl4AI 服务可用。
func (h *Service) EnsureReady() error {
	h.mu.Lock()
	h.setStatusLocked(define.Crawl4AIStatusInstalling, `正在初始化 Crawl4AI`, ``)
	h.mu.Unlock()
	if err := h.ensurePython(); err != nil {
		h.mu.Lock()
		h.setStatusLocked(define.Crawl4AIStatusFailed, `Crawl4AI 初始化失败`, err.Error())
		h.installing = false
		h.mu.Unlock()
		return err
	}
	h.mu.Lock()
	h.setStatusLocked(define.Crawl4AIStatusInstalling, `正在检查 Crawl4AI 依赖`, ``)
	h.mu.Unlock()
	if err := h.ensurePackages(); err != nil {
		h.mu.Lock()
		h.setStatusLocked(define.Crawl4AIStatusFailed, `Crawl4AI 依赖安装失败`, err.Error())
		h.installing = false
		h.mu.Unlock()
		return err
	}
	if h.isServiceRunning() {
		h.mu.Lock()
		h.setStatusLocked(define.Crawl4AIStatusReady, `Crawl4AI 服务已就绪`, ``)
		h.installing = false
		h.mu.Unlock()
		return nil
	}
	h.mu.Lock()
	h.setStatusLocked(define.Crawl4AIStatusInstalling, `正在启动 Crawl4AI 服务`, ``)
	h.mu.Unlock()
	if err := h.startService(); err != nil {
		h.mu.Lock()
		h.setStatusLocked(define.Crawl4AIStatusFailed, `Crawl4AI 服务启动失败`, err.Error())
		h.installing = false
		h.mu.Unlock()
		return err
	}
	if err := h.waitServiceReady(90 * time.Second); err != nil {
		h.mu.Lock()
		h.setStatusLocked(define.Crawl4AIStatusFailed, `Crawl4AI 服务启动失败`, err.Error())
		h.installing = false
		h.mu.Unlock()
		return err
	}
	h.mu.Lock()
	h.setStatusLocked(define.Crawl4AIStatusReady, `Crawl4AI 服务已就绪`, ``)
	h.installing = false
	h.mu.Unlock()
	return nil
}

// Stop 停止由当前应用启动的 Crawl4AI 服务。
func (h *Service) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.process == nil || h.process.Process == nil || !h.startedByMe {
		return
	}
	_ = h.process.Process.Kill()
	_, _ = h.process.Process.Wait()
	h.process = nil
	h.startedByMe = false
}

// EnsureReadyAsync 异步确保 Crawl4AI 服务可用。
func (h *Service) EnsureReadyAsync() {
	h.mu.Lock()
	if h.status == define.Crawl4AIStatusReady || h.installing {
		h.mu.Unlock()
		return
	}
	h.installing = true
	h.setStatusLocked(define.Crawl4AIStatusInstalling, `正在初始化 Crawl4AI`, ``)
	h.mu.Unlock()
	go func() {
		if err := h.EnsureReady(); err != nil {
			gstool.FmtPrintlnLogTime(`Crawl4AI 初始化失败 %s`, err.Error())
			return
		}
		gstool.FmtPrintlnLogTime(`Crawl4AI 服务已就绪 %s`, h.Env.Crawl4AIBaseURL)
	}()
}

// Status 返回 Crawl4AI 当前状态快照。
func (h *Service) Status() map[string]any {
	h.mu.Lock()
	defer h.mu.Unlock()
	return map[string]any{
		`status`:        h.status,
		`status_text`:   h.statusText,
		`error_message`: h.lastError,
		`is_ready`:      h.status == define.Crawl4AIStatusReady,
		`is_installing`: h.installing || h.status == define.Crawl4AIStatusInstalling,
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

// ensurePython 检测并记录 python 命令。
func (h *Service) ensurePython() error {
	if h.Env.PythonCommand != `` {
		return nil
	}
	candidateList := [][]string{
		{`python`, `-c`, `import sys;print(sys.executable)`},
		{`py`, `-3`, `-c`, `import sys;print(sys.executable)`},
	}
	for _, args := range candidateList {
		command := exec.Command(args[0], args[1:]...)
		command.SysProcAttr = windowsHideAttr()
		output, err := command.Output()
		if err == nil && strings.TrimSpace(string(output)) != `` {
			h.Env.PythonCommand = args[0]
			return nil
		}
	}
	return errors.New(`未检测到 Python，请先安装 Python 并加入 PATH`)
}

// ensurePackages 检测 Crawl4AI 相关依赖是否已安装，未安装则自动 pip install。
func (h *Service) ensurePackages() error {
	checkCode := "import importlib.util,sys;mods=['crawl4ai','fastapi','uvicorn'];missing=[m for m in mods if importlib.util.find_spec(m) is None];print('|'.join(missing));sys.exit(0 if not missing else 1)"
	if err := h.runPythonCode(checkCode, 30*time.Second); err == nil {
		return nil
	}
	return h.installPackages()
}

// installPackages 自动安装 Crawl4AI 相关依赖。
func (h *Service) installPackages() error {
	packageList := []string{`crawl4ai`, `fastapi`, `uvicorn`}
	packageList = append(packageList, h.pythonCompatPackages()...)
	baseArgs := h.pythonBaseArgs()
	baseArgs = append(baseArgs, `-m`, `pip`, `install`)
	baseArgs = append(baseArgs,
		`--index-url`, `https://pypi.org/simple`,
		`--trusted-host`, `pypi.org`,
		`--trusted-host`, `files.pythonhosted.org`,
	)
	baseArgs = append(baseArgs, packageList...)
	output, err := h.runCommandWithEnv(baseArgs, h.pythonInstallEnv(false))
	if err == nil {
		return nil
	}
	cleanArgs := append([]string{}, h.pythonBaseArgs()...)
	cleanArgs = append(cleanArgs, `-m`, `pip`, `--isolated`, `install`)
	cleanArgs = append(cleanArgs,
		`--index-url`, `https://pypi.org/simple`,
		`--trusted-host`, `pypi.org`,
		`--trusted-host`, `files.pythonhosted.org`,
	)
	cleanArgs = append(cleanArgs, packageList...)
	cleanOutput, cleanErr := h.runCommandWithEnv(cleanArgs, h.pythonInstallEnv(true))
	if cleanErr == nil {
		return nil
	}
	return fmt.Errorf(
		"自动安装 Crawl4AI 失败。\n首次安装输出：%s\n\n清理代理后重试输出：%s",
		strings.TrimSpace(string(output)),
		strings.TrimSpace(string(cleanOutput)),
	)
}

// pythonCompatPackages 根据 Python 版本返回兼容依赖。
func (h *Service) pythonCompatPackages() []string {
	// Python 3.9 + Windows 安装最新 greenlet 时可能退回源码编译，要求本地安装 VC++ Build Tools。
	// 这里优先锁定已有 wheel 的版本，避免自动安装失败。
	if h.pythonVersionAtMost(3, 9) {
		return []string{`greenlet==3.1.1`}
	}
	return nil
}

// startService 启动本地 Crawl4AI API 服务。
func (h *Service) startService() error {
	if h.Env.Crawl4AIScriptPath == `` {
		return errors.New(`Crawl4AI 服务脚本路径为空`)
	}
	args := append(h.pythonBaseArgs(), h.Env.Crawl4AIScriptPath)
	command := exec.Command(h.Env.PythonCommand, args...)
	command.SysProcAttr = windowsHideAttr()
	command.Dir = filepath.Dir(h.Env.Crawl4AIScriptPath)
	command.Env = append(os.Environ(),
		`PYTHONIOENCODING=utf-8`,
		`CRAWL4AI_HOST=`+h.Env.Crawl4AIHost,
		`CRAWL4AI_PORT=`+h.Env.Crawl4AIPort,
		`CRAWL4AI_DATA_DIR=`+h.Env.Crawl4AIDataPath,
		`CRAWL4AI_HEADLESS=true`,
	)
	logFilePath := filepath.Join(h.Env.LogPath, `crawl4ai-service.log`)
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	command.Stdout = logFile
	command.Stderr = logFile
	if err = command.Start(); err != nil {
		_ = logFile.Close()
		return err
	}
	h.process = command
	h.startedByMe = true
	go func() {
		_ = command.Wait()
		_ = logFile.Close()
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.process == command {
			h.process = nil
			h.startedByMe = false
		}
	}()
	return nil
}

// isServiceRunning 检测服务是否已运行。
func (h *Service) isServiceRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(h.Env.Crawl4AIBaseURL + `/healthz`)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode < 300
}

// waitServiceReady 等待服务就绪。
func (h *Service) waitServiceReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if h.isServiceRunning() {
			return nil
		}
		time.Sleep(1500 * time.Millisecond)
	}
	return errors.New(`Crawl4AI 服务启动超时`)
}

// setStatusLocked 在已加锁场景下更新状态信息。
func (h *Service) setStatusLocked(status, statusText, lastError string) {
	h.status = status
	h.statusText = statusText
	h.lastError = lastError
}

// runPythonCode 执行一段 Python 代码。
func (h *Service) runPythonCode(code string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	args := append(h.pythonBaseArgs(), `-c`, code)
	command := exec.CommandContext(ctx, h.Env.PythonCommand, args...)
	command.SysProcAttr = windowsHideAttr()
	command.Env = append(os.Environ(), `PYTHONIOENCODING=utf-8`)
	output, err := command.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New(`执行 Python 检测超时`)
	}
	if err != nil {
		return fmt.Errorf(`执行 Python 检测失败: %s`, strings.TrimSpace(string(output)))
	}
	return nil
}

// pythonVersionAtMost 判断当前 Python 版本是否小于等于指定版本。
func (h *Service) pythonVersionAtMost(major, minor int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := append(h.pythonBaseArgs(), `-c`, `import sys;print(f"{sys.version_info[0]}.{sys.version_info[1]}")`)
	command := exec.CommandContext(ctx, h.Env.PythonCommand, args...)
	command.SysProcAttr = windowsHideAttr()
	command.Env = append(os.Environ(), `PYTHONIOENCODING=utf-8`)
	output, err := command.CombinedOutput()
	if err != nil {
		return false
	}
	versionText := strings.TrimSpace(string(output))
	parts := strings.Split(versionText, `.`)
	if len(parts) < 2 {
		return false
	}
	pyMajor, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	pyMinor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	if pyMajor != major {
		return pyMajor < major
	}
	return pyMinor <= minor
}

// runCommandWithEnv 使用指定环境执行 Python 命令。
func (h *Service) runCommandWithEnv(args []string, env []string) ([]byte, error) {
	command := exec.Command(h.Env.PythonCommand, args...)
	command.SysProcAttr = windowsHideAttr()
	command.Env = env
	return command.CombinedOutput()
}

// pythonInstallEnv 返回安装依赖时使用的环境变量。
func (h *Service) pythonInstallEnv(cleanProxy bool) []string {
	envList := os.Environ()
	filteredEnv := make([]string, 0, len(envList)+6)
	for _, item := range envList {
		upperItem := strings.ToUpper(item)
		if cleanProxy {
			if strings.HasPrefix(upperItem, `HTTP_PROXY=`) ||
				strings.HasPrefix(upperItem, `HTTPS_PROXY=`) ||
				strings.HasPrefix(upperItem, `ALL_PROXY=`) ||
				strings.HasPrefix(upperItem, `NO_PROXY=`) ||
				strings.HasPrefix(upperItem, `PIP_PROXY=`) {
				continue
			}
		}
		filteredEnv = append(filteredEnv, item)
	}
	filteredEnv = append(filteredEnv,
		`PYTHONIOENCODING=utf-8`,
		`PIP_DISABLE_PIP_VERSION_CHECK=1`,
		`PIP_DEFAULT_TIMEOUT=120`,
	)
	if cleanProxy {
		// Windows 下 urllib/request 可能从系统代理读取配置，这里显式直连，避免 pip 继续走代理。
		filteredEnv = append(filteredEnv,
			`NO_PROXY=*`,
			`no_proxy=*`,
		)
	}
	return filteredEnv
}

// pythonBaseArgs 返回 Python 启动基础参数。
func (h *Service) pythonBaseArgs() []string {
	if h.Env.PythonCommand == `py` {
		return []string{`-3`}
	}
	return []string{}
}

// windowsHideAttr 返回隐藏窗口配置。
func windowsHideAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}

// IsPortListening 检测端口是否已监听。
func IsPortListening(host, port string) bool {
	conn, err := net.DialTimeout(`tcp`, net.JoinHostPort(host, port), time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
