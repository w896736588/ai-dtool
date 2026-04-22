package controller

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dev_tool/internal/app/dtool/component"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
)

// agentDownloadSpec 定义客户端下载时需要的目标平台和文件名。
// agentDownloadSpec describes the target platform and artifact filename for client downloads.
type agentDownloadSpec struct {
	Goos     string
	FileName string
	IsZip    bool
}

// buildAgentDownloadBinaryFunc 允许测试替换下载构建逻辑，避免单测真实编译二进制。
var buildAgentDownloadBinaryFunc = buildAgentDownloadBinary

// resolveAgentDownloadSpec 根据 os 查询参数解析目标平台，避免下载请求打到未知平台。
// resolveAgentDownloadSpec maps the os query parameter to a supported download target.
func resolveAgentDownloadSpec(rawOS string) (agentDownloadSpec, bool) {
	switch strings.ToLower(strings.TrimSpace(rawOS)) {
	case "windows":
		return agentDownloadSpec{
			Goos:     "windows",
			FileName: "dtool-agent.exe",
			IsZip:    false,
		}, true
	case "darwin":
		return agentDownloadSpec{
			Goos:     "darwin",
			FileName: "dtool-agent",
			IsZip:    false,
		}, true
	case "linux":
		return agentDownloadSpec{
			Goos:     "linux",
			FileName: "dtool-agent",
			IsZip:    false,
		}, true
	default:
		return agentDownloadSpec{}, false
	}
}

// buildAgentDefaultServerURL 根据下载请求推导客户端默认回连地址，避免远程访问时仍写死到 localhost。
func buildAgentDefaultServerURL(req *http.Request) string {
	proto := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto"))
	if proto == "" {
		if req.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	host := strings.TrimSpace(req.Host)
	if host == "" {
		host = "localhost:17170"
	}

	return proto + "://" + host
}

// buildAgentDownloadBinary 按目标平台现编 dtool-agent，确保下载入口至少能返回可执行文件。
// buildAgentDownloadBinary builds the dtool-agent binary on demand so the download endpoint can return a runnable artifact.
func buildAgentDownloadBinary(spec agentDownloadSpec, defaultServerURL string) (string, error) {
	rootPath := component.EnvClient.RootPath
	outputDir := filepath.Join(rootPath, "tmp", "agent_download")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}

	binaryName := "dtool-agent"
	if spec.Goos == "windows" {
		binaryName = "dtool-agent.exe"
	}
	outputPath := filepath.Join(outputDir, binaryName)
	ldflags := fmt.Sprintf("-s -w -X main.defaultServerURL=%s", defaultServerURL)
	buildCmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", outputPath, "./cmd/dtool-agent")
	buildCmd.Dir = rootPath
	buildCmd.Env = append(
		os.Environ(),
		"GOOS="+spec.Goos,
		"GOARCH=amd64",
		"CGO_ENABLED=0",
	)

	buildOutput, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		return "", fmt.Errorf("build dtool-agent failed: %w, output: %s", buildErr, string(buildOutput))
	}

	// Windows 平台打包成 zip
	if spec.IsZip {
		zipPath := filepath.Join(outputDir, spec.FileName)
		if err := createZipFile(zipPath, binaryName, outputPath); err != nil {
			return "", fmt.Errorf("create zip failed: %w", err)
		}
		// 删除原始 exe 文件，只保留 zip
		os.Remove(outputPath)
		return zipPath, nil
	}

	return outputPath, nil
}

// createZipFile 创建 zip 文件，将 sourcePath 文件打包到 zip 中的 entryName
func createZipFile(zipPath, entryName, sourcePath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}
	header.Name = entryName
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, sourceFile)
	return err
}

// AgentDownload 提供本地客户端安装包下载，当前实现为按目标平台即时构建二进制。
// AgentDownload serves the local client installer by building the agent binary for the requested platform on demand.
func AgentDownload(c *gin.Context) {
	spec, ok := resolveAgentDownloadSpec(c.Query("os"))
	if !ok {
		respondAgentDownloadError(c, http.StatusBadRequest, "unsupported os, expected windows, darwin or linux")
		return
	}

	binaryPath, err := buildAgentDownloadBinaryFunc(spec, buildAgentDefaultServerURL(c.Request))
	if err != nil {
		respondAgentDownloadError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.FileAttachment(binaryPath, spec.FileName)
}

// respondAgentDownloadError 返回结构化 JSON 错误，避免前端把错误文本误存成 exe。
func respondAgentDownloadError(c *gin.Context, statusCode int, errMsg string) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Status(statusCode)
	c.Writer.WriteHeaderNow()

	responseBody, _ := json.Marshal(map[string]any{
		"ErrCode": gsgin.CodeError,
		"ErrMsg":  errMsg,
		"Data":    nil,
	})
	_, _ = c.Writer.Write(responseBody)
}
