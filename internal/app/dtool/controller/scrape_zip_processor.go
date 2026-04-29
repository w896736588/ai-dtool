package controller

import (
	"archive/zip"
	"dev_tool/internal/app/dtool/component"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

// resolveDownloadFilePath 将 /api/download/xxx.zip 转为磁盘上的物理路径。
func resolveDownloadFilePath(downloadURL string) string {
	fileName := strings.TrimPrefix(downloadURL, "/api/download/")
	if fileName == downloadURL {
		return ""
	}
	return buildWebDownloadFilePath(fileName)
}

// saveScrapeImagesToMemoryDir 将 ZIP 中 images/ 目录下的图片保存到 memory images 目录。
// 返回 zip 中原始路径 → /memory/images/{newName} 的映射。
func saveScrapeImagesToMemoryDir(zipReader *zip.Reader, memoryDir string) (map[string]string, error) {
	imageDir := filepath.Join(memoryDir, "images")
	if err := gstool.DirCreatePath(imageDir); err != nil {
		return nil, fmt.Errorf("创建图片目录失败: %w", err)
	}

	pathMapping := make(map[string]string)
	for _, f := range zipReader.File {
		if !strings.HasPrefix(f.Name, "images/") || f.Name == "images/" {
			continue
		}
		if f.Mode().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}
		content, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			continue
		}

		ext := strings.ToLower(filepath.Ext(f.Name))
		newName := fmt.Sprintf("%d%s", time.Now().UnixMicro(), ext)
		dst := filepath.Join(imageDir, newName)
		if writeErr := os.WriteFile(dst, content, 0o644); writeErr != nil {
			continue
		}

		pathMapping[f.Name] = "/memory/images/" + newName
		baseName := path.Base(f.Name)
		pathMapping[baseName] = "/memory/images/" + newName
	}
	return pathMapping, nil
}

// rewriteScrapeImagePaths 将 markdown 中的 images/xxx 路径替换为 /memory/images/{newName}。
func rewriteScrapeImagePaths(markdown string, pathMapping map[string]string) string {
	result := markdown
	for oldPath, newPath := range pathMapping {
		result = strings.ReplaceAll(result, "("+oldPath+")", "("+newPath+")")
	}
	return result
}

// memoryImageBaseURL 返回记忆库图片服务的完整基地址（如 http://localhost:17170）。
func memoryImageBaseURL() string {
	host := "localhost"
	port := "17170"
	if component.EnvClient != nil && len(component.EnvClient.Ports) > 0 {
		p := strings.TrimSpace(component.EnvClient.Ports[0])
		if p != "" {
			port = p
		}
	}
	return "http://" + host + ":" + port
}

// prefixMemoryImagePaths 将 markdown 中的 /memory/images/ 相对路径替换为带完整服务端地址的绝对 URL。
func prefixMemoryImagePaths(markdown string) string {
	return strings.ReplaceAll(markdown, "(/memory/images/", "("+memoryImageBaseURL()+"/memory/images/")
}

// prefixRelativeURL 将以 / 开头的相对路径转为带完整服务端地址的绝对 URL；已是完整 URL 则原样返回。
func prefixRelativeURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "/") {
		return memoryImageBaseURL() + rawURL
	}
	return rawURL
}

// prependDownloadURLToMarkdown 在 markdown 头部（第一个标题行之后）插入下载链接。
func prependDownloadURLToMarkdown(markdown, downloadURL string) string {
	if strings.TrimSpace(downloadURL) == "" {
		return markdown
	}
	downloadLine := fmt.Sprintf("[下载原始ZIP](%s)", downloadURL)

	lines := strings.Split(markdown, "\n")
	insertIdx := 0
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			insertIdx = i + 1
			break
		}
	}

	if insertIdx == 0 {
		return downloadLine + "\n\n" + markdown
	}

	insertLines := []string{"", downloadLine, ""}
	resultLines := make([]string, 0, len(lines)+len(insertLines))
	resultLines = append(resultLines, lines[:insertIdx]...)
	resultLines = append(resultLines, insertLines...)
	resultLines = append(resultLines, lines[insertIdx:]...)
	return strings.Join(resultLines, "\n")
}

// processScrapeZipResult 解压 ZIP、保存图片到 memory 目录、重写路径、插入 download_url。
func processScrapeZipResult(downloadURL string, memoryFragmentID string) (map[string]any, error) {
	if component.MemoryRuntime == nil {
		return nil, fmt.Errorf("记忆库未配置")
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return nil, err
	}
	memoryDir := component.MemoryRuntime.Config().Dir

	zipPath := resolveDownloadFilePath(downloadURL)
	if zipPath == "" {
		return nil, fmt.Errorf("无法定位 ZIP 文件: %s", downloadURL)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("打开 ZIP 文件失败: %w", err)
	}
	defer reader.Close()

	var markdownContent string
	for _, f := range reader.File {
		if f.Name == "content.md" {
			rc, openErr := f.Open()
			if openErr != nil {
				return nil, fmt.Errorf("打开 content.md 失败: %w", openErr)
			}
			content, readErr := io.ReadAll(rc)
			_ = rc.Close()
			if readErr != nil {
				return nil, fmt.Errorf("读取 content.md 失败: %w", readErr)
			}
			markdownContent = string(content)
			break
		}
	}
	if markdownContent == "" {
		return nil, fmt.Errorf("ZIP 中未找到 content.md")
	}

	pathMapping, imgErr := saveScrapeImagesToMemoryDir(&reader.Reader, memoryDir)
	if imgErr != nil {
		return nil, fmt.Errorf("保存图片失败: %w", imgErr)
	}
	imageCount := 0
	for k := range pathMapping {
		if strings.HasPrefix(k, "images/") {
			imageCount++
		}
	}

	markdownContent = rewriteScrapeImagePaths(markdownContent, pathMapping)
	markdownContent = prefixMemoryImagePaths(markdownContent)
	markdownContent = prependDownloadURLToMarkdown(markdownContent, prefixRelativeURL(downloadURL))

	return map[string]any{
		"markdown":     markdownContent,
		"fragment_id":  memoryFragmentID,
		"image_count":  imageCount,
		"download_url": downloadURL,
	}, nil
}
