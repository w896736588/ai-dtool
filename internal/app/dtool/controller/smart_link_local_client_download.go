package controller

import (
	"dev_tool/internal/app/dtool/component"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

const (
	// agentBuildStatusPending 表示任务已创建但尚未进入真正编译。
	// agentBuildStatusPending means the job exists but the actual build has not started yet.
	agentBuildStatusPending = "pending"
	// agentBuildStatusBuilding 表示服务端正在执行 go build 或产物校验。
	// agentBuildStatusBuilding means the server is running go build or validating the produced artifact.
	agentBuildStatusBuilding = "building"
	// agentBuildStatusReady 表示编译产物已经准备好，前端可以发起下载。
	// agentBuildStatusReady means the artifact is ready and the frontend can start downloading it.
	agentBuildStatusReady = "ready"
	// agentBuildStatusDownloading 表示后端正在把产物流式返回给浏览器。
	// agentBuildStatusDownloading means the backend is currently streaming the artifact to the browser.
	agentBuildStatusDownloading = "downloading"
	// agentBuildStatusCompleted 表示下载已经完成，按钮可以回到默认状态。
	// agentBuildStatusCompleted means the download finished and the button can eventually reset.
	agentBuildStatusCompleted = "completed"
	// agentBuildStatusFailed 表示任务在参数校验、编译或下载阶段失败。
	// agentBuildStatusFailed means the job failed during validation, build, or download.
	agentBuildStatusFailed = "failed"
)

var (
	// agentBuildConfigNameCleaner 统一配置名后缀，只保留文件名友好的字符。
	// agentBuildConfigNameCleaner normalizes config suffixes so generated artifact names stay filesystem-safe.
	agentBuildConfigNameCleaner = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	// GlobalAgentBuildJobStore 保存下载编译任务，供前端轮询进度。
	// GlobalAgentBuildJobStore keeps agent build jobs so the frontend can poll live progress.
	GlobalAgentBuildJobStore = newAgentBuildJobStore()
	agentBuildJobCounter     uint64
)

type agentDownloadSpec struct {
	Platform string
	Goos     string
	FileName string
}

type agentBuildJob struct {
	ID           string `json:"job_id"`
	Platform     string `json:"platform"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	FileName     string `json:"file_name"`
	FilePath     string `json:"-"`
	Progress     int    `json:"progress"`
	Error        string `json:"error"`
	Host         string `json:"host"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
	DownloadedAt int64  `json:"downloaded_at"`
}

type agentBuildJobStore struct {
	mu   sync.RWMutex
	jobs map[string]*agentBuildJob
}

// newAgentBuildJobStore 创建内存态任务仓库，避免为短生命周期下载任务落库。
// newAgentBuildJobStore creates an in-memory job store so short-lived download jobs do not need persistence.
func newAgentBuildJobStore() *agentBuildJobStore {
	return &agentBuildJobStore{
		jobs: make(map[string]*agentBuildJob),
	}
}

// save 保存或覆盖下载任务快照，供轮询接口实时读取。
// save stores or replaces the latest job snapshot for the polling status API.
func (h *agentBuildJobStore) save(job *agentBuildJob) {
	if job == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.jobs[job.ID] = job
}

// get 按任务 id 读取当前下载编译状态。
// get returns the current build job state by job id.
func (h *agentBuildJobStore) get(jobID string) *agentBuildJob {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.jobs[jobID]
}

// buildAgentConfigNameSuffix 提取当前配置文件名，作为下载产物后缀。
// buildAgentConfigNameSuffix extracts the active config file name for artifact suffixes.
func buildAgentConfigNameSuffix(configFile string) string {
	normalized := strings.TrimSpace(configFile)
	if normalized == "" {
		normalized = "config"
	}
	baseName := filepath.Base(normalized)
	ext := filepath.Ext(baseName)
	if ext != "" {
		baseName = strings.TrimSuffix(baseName, ext)
	}
	baseName = strings.TrimSpace(baseName)
	baseName = agentBuildConfigNameCleaner.ReplaceAllString(baseName, "-")
	baseName = strings.Trim(baseName, "-_")
	if baseName == "" {
		return "config"
	}
	return baseName
}

// resolveAgentDownloadSpec 根据平台和配置名生成目标二进制描述。
// resolveAgentDownloadSpec resolves platform-specific target metadata for the generated agent binary.
func resolveAgentDownloadSpec(platform, configSuffix string) (agentDownloadSpec, bool) {
	suffix := buildAgentConfigNameSuffix(configSuffix)
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "windows":
		return agentDownloadSpec{
			Platform: "windows",
			Goos:     "windows",
			FileName: fmt.Sprintf("dtool-agent-%s.exe", suffix),
		}, true
	case "macos":
		return agentDownloadSpec{
			Platform: "macos",
			Goos:     "darwin",
			FileName: fmt.Sprintf("dtool-agent-%s", suffix),
		}, true
	default:
		return agentDownloadSpec{}, false
	}
}

// buildAgentBuildLdflags 生成构建时注入的默认连接参数。
// buildAgentBuildLdflags builds ldflags that inject default host and client version into dtool-agent.
func buildAgentBuildLdflags(host, clientVersion string) string {
	return fmt.Sprintf("-X main.defaultServerURL=%s -X main.defaultClientVersion=%s", strings.TrimSpace(host), strings.TrimSpace(clientVersion))
}

// nextAgentBuildJobID 生成单进程内唯一任务 id，方便前端轮询和下载。
// nextAgentBuildJobID generates a process-unique job id for polling and download requests.
func nextAgentBuildJobID() string {
	counter := atomic.AddUint64(&agentBuildJobCounter, 1)
	return fmt.Sprintf("agent_build_%d_%d", time.Now().UnixNano(), counter)
}

// setProgress 更新任务阶段、进度和提示文案。
// setProgress updates the job phase, progress percentage, and user-facing message.
func (h *agentBuildJob) setProgress(status string, progress int, message string) {
	h.Status = status
	h.Progress = progress
	h.Message = message
	h.UpdatedAt = time.Now().Unix()
}

// setFailure 标记任务失败，并记录最终错误信息。
// setFailure marks the job as failed and stores the final error message.
func (h *agentBuildJob) setFailure(message string) {
	h.Status = agentBuildStatusFailed
	h.Error = message
	h.Message = message
	h.Progress = 100
	h.UpdatedAt = time.Now().Unix()
}

// setReady 标记任务编译完成，并保存实际产物路径。
// setReady marks the job as ready and stores the generated artifact path.
func (h *agentBuildJob) setReady(fileName, filePath string) {
	h.Status = agentBuildStatusReady
	h.FileName = fileName
	h.FilePath = filePath
	h.Progress = 100
	h.Message = "编译完成，等待下载"
	h.UpdatedAt = time.Now().Unix()
}

// SmartLinkClientBuildStart 创建客户端编译任务。
// SmartLinkClientBuildStart starts a background build job for the requested smart-link local client.
func SmartLinkClientBuildStart(c *gin.Context) {
	req := map[string]any{}
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}

	platform := strings.ToLower(strings.TrimSpace(cast.ToString(req["platform"])))
	host := strings.TrimSpace(cast.ToString(req["host"]))
	// host 是注入到 dtool-agent 的默认服务端地址，不能为空。
	// host becomes the default server URL injected into dtool-agent, so it must not be empty.
	if host == "" {
		gsgin.GinResponseError(c, "host 不能为空", nil)
		return
	}

	spec, ok := resolveAgentDownloadSpec(platform, component.EnvClient.ConfigFile)
	// 当前只开放 windows 和 macos，避免前端传入未约定平台造成构建语义漂移。
	// Only windows and macos are accepted so unsupported platform inputs cannot drift build semantics.
	if !ok {
		gsgin.GinResponseError(c, "platform 仅支持 windows 或 macos", nil)
		return
	}

	job := &agentBuildJob{
		ID:        nextAgentBuildJobID(),
		Platform:  spec.Platform,
		Status:    agentBuildStatusPending,
		Message:   "准备编译参数",
		Progress:  5,
		Host:      host,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		FileName:  spec.FileName,
	}
	GlobalAgentBuildJobStore.save(job)
	go runAgentBuildJob(job, spec)

	gsgin.GinResponseSuccess(c, "", map[string]any{
		"job_id":    job.ID,
		"status":    job.Status,
		"progress":  job.Progress,
		"message":   job.Message,
		"file_name": job.FileName,
	})
}

// SmartLinkClientBuildStatus 查询客户端编译任务状态。
// SmartLinkClientBuildStatus returns the current smart-link client build job progress.
func SmartLinkClientBuildStatus(c *gin.Context) {
	jobID := strings.TrimSpace(c.Query("job_id"))
	if jobID == "" {
		gsgin.GinResponseError(c, "job_id 不能为空", nil)
		return
	}
	job := GlobalAgentBuildJobStore.get(jobID)
	if job == nil {
		gsgin.GinResponseError(c, "编译任务不存在", nil)
		return
	}

	gsgin.GinResponseSuccess(c, "", job)
}

// SmartLinkClientBuildDownload 下载已编译好的客户端产物。
// SmartLinkClientBuildDownload serves the generated client binary once the build job is ready.
func SmartLinkClientBuildDownload(c *gin.Context) {
	jobID := strings.TrimSpace(c.Param("job_id"))
	if jobID == "" {
		gsgin.GinResponseError(c, "job_id 不能为空", nil)
		return
	}
	job := GlobalAgentBuildJobStore.get(jobID)
	if job == nil {
		gsgin.GinResponseError(c, "编译任务不存在", nil)
		return
	}
	// 只有 ready/downloading/completed 才允许取文件，避免下载到半成品或失败状态。
	// Only ready/downloading/completed jobs may serve files so partial or failed artifacts are never downloaded.
	if job.Status != agentBuildStatusReady && job.Status != agentBuildStatusDownloading && job.Status != agentBuildStatusCompleted {
		gsgin.GinResponseError(c, "编译任务尚未完成", nil)
		return
	}
	if strings.TrimSpace(job.FilePath) == "" {
		gsgin.GinResponseError(c, "编译产物不存在", nil)
		return
	}
	if _, err := os.Stat(job.FilePath); err != nil {
		gsgin.GinResponseError(c, "编译产物不存在", nil)
		return
	}

	job.setProgress(agentBuildStatusDownloading, 100, "正在下载客户端")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, job.FileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("X-Download-Filename", job.FileName)
	c.File(job.FilePath)
	job.Status = agentBuildStatusCompleted
	job.Message = "下载完成"
	job.DownloadedAt = time.Now().Unix()
	job.UpdatedAt = time.Now().Unix()
}

// runAgentBuildJob 在后台执行交叉编译，并把阶段进度回写到任务仓库。
// runAgentBuildJob performs the cross-compilation in background and writes progress back to the job store.
func runAgentBuildJob(job *agentBuildJob, spec agentDownloadSpec) {
	job.setProgress(agentBuildStatusPending, 10, "准备编译环境")

	outputDir := filepath.Join(component.EnvClient.RootPath, "tmp", "agent_builds", job.ID)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		job.setFailure("创建编译目录失败: " + err.Error())
		return
	}

	outputPath := filepath.Join(outputDir, spec.FileName)
	job.setProgress(agentBuildStatusBuilding, 35, "正在编译客户端")

	buildCommand := exec.Command("go", "build", "-trimpath", "-ldflags", buildAgentBuildLdflags(job.Host, getSmartLinkConfig().ClientVersion), "-o", outputPath, "./cmd/dtool-agent")
	buildCommand.Dir = component.EnvClient.RootPath
	// 固定 GOARCH=amd64 和 CGO_ENABLED=0，降低跨平台编译对宿主环境依赖。
	// Keep GOARCH=amd64 and CGO_ENABLED=0 fixed to reduce host-environment coupling during cross compilation.
	buildCommand.Env = append(os.Environ(),
		"GOOS="+spec.Goos,
		"GOARCH=amd64",
		"CGO_ENABLED=0",
	)
	outputBytes, err := buildCommand.CombinedOutput()
	if err != nil {
		errText := strings.TrimSpace(string(outputBytes))
		// 优先返回 go build 原始输出，方便前端直接看到真实编译错误。
		// Prefer raw go build output so the frontend can surface the real compilation failure.
		if errText == "" {
			errText = err.Error()
		}
		job.setFailure("编译失败: " + errText)
		return
	}

	job.setProgress(agentBuildStatusBuilding, 80, "正在校验编译产物")
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		job.setFailure("校验编译产物失败: " + err.Error())
		return
	}
	// 空文件通常意味着构建链路异常，即便命令退出成功也要阻断下载。
	// Empty artifacts usually indicate a broken build pipeline, so downloads must be blocked even if the command exited successfully.
	if fileInfo.Size() <= 0 {
		job.setFailure("编译产物为空文件")
		return
	}

	job.setReady(spec.FileName, outputPath)
}
