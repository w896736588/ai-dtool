package plw

import (
	"archive/zip"
	"bytes"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"io"
	"mime"
	"net/http"
	neturl "net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/playwright-community/playwright-go"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type ScrapeImageResource struct {
	SourceURL        string
	CandidateURLs    []string
	RawCandidateURLs []string
}

type ScrapeMarkdownResult struct {
	Markdown []byte
	ZipBytes []byte
	FileName string
}

func logScrapeMarkdownStep(log *gstool.GsSlog, runParams *PlaywrightRunParams, step, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if log != nil {
		log.Debugf("[scrape_markdown][%s] %s", step, message)
	}
	if runParams != nil && runParams.StreamFunc != nil {
		runParams.StreamFunc(step, message)
	}
}

// buildScrapeProcessList 复制原流程并在末尾追加跳转节点，保持现有自定义网页执行模型。
func buildScrapeProcessList(processList []map[string]any, jumpURL string) []map[string]any {
	newList := make([]map[string]any, 0, len(processList)+1)
	for _, item := range processList {
		newItem := make(map[string]any, len(item))
		for key, value := range item {
			newItem[key] = value
		}
		newList = append(newList, newItem)
	}
	newList = append(newList, map[string]any{
		"name":              "跳转到抓取页面",
		"type":              string(define.RedirectUri),
		"value":             jumpURL,
		"wait_mills":        0,
		"is_async":          0,
		"is_error_continue": 0,
		"domain_limit":      "",
		"tip":               "跳转到抓取页面",
	})
	return newList
}

// extractImageSources 从 HTML 中提取图片链接，并统一转成绝对地址。
func extractImageSources(htmlText, jumpURL string) ([]ScrapeImageResource, error) {
	root, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return nil, err
	}
	baseURL, err := neturl.Parse(jumpURL)
	if err != nil {
		return nil, err
	}
	result := make([]ScrapeImageResource, 0)
	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "img") {
			rawCandidateURLs := collectRawImageCandidateURLs(node)
			candidateURLs := collectImageCandidateURLs(node, baseURL)
			if len(candidateURLs) > 0 || len(rawCandidateURLs) > 0 {
				sourceURL := ""
				if len(candidateURLs) > 0 {
					sourceURL = candidateURLs[0]
				}
				result = append(result, ScrapeImageResource{
					SourceURL:        sourceURL,
					CandidateURLs:    candidateURLs,
					RawCandidateURLs: rawCandidateURLs,
				})
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return result, nil
}

func pickBestImageSource(node *html.Node, baseURL *neturl.URL) string {
	candidateURLs := collectImageCandidateURLs(node, baseURL)
	if len(candidateURLs) == 0 {
		return ""
	}
	return candidateURLs[0]
}

// collectRawImageCandidateURLs 收集 HTML 原始图片地址，确保 Markdown 替换能命中页面里的相对路径。
func collectRawImageCandidateURLs(node *html.Node) []string {
	attrMap := make(map[string]string)
	for _, attr := range node.Attr {
		attrMap[strings.ToLower(strings.TrimSpace(attr.Key))] = strings.TrimSpace(attr.Val)
	}
	candidates := make([]string, 0, 4)
	if srcsetValue := attrMap["srcset"]; srcsetValue != "" {
		if srcsetURL := parseLargestSrcsetURL(srcsetValue); srcsetURL != "" {
			candidates = append(candidates, srcsetURL)
		}
	}
	for _, key := range []string{"data-src", "data-original", "src"} {
		if attrMap[key] != "" {
			candidates = append(candidates, attrMap[key])
		}
	}
	result := make([]string, 0, len(candidates))
	existsMap := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, exists := existsMap[candidate]; exists {
			continue
		}
		existsMap[candidate] = struct{}{}
		result = append(result, candidate)
	}
	return result
}

// collectImageCandidateURLs 收集图片节点的候选地址，避免 Markdown 保留原始 src 时丢失本地替换。
func collectImageCandidateURLs(node *html.Node, baseURL *neturl.URL) []string {
	candidates := collectRawImageCandidateURLs(node)
	result := make([]string, 0, len(candidates))
	existsMap := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		absoluteURL := resolveAbsoluteURL(baseURL, candidate)
		if absoluteURL == "" {
			continue
		}
		if _, exists := existsMap[absoluteURL]; exists {
			continue
		}
		existsMap[absoluteURL] = struct{}{}
		result = append(result, absoluteURL)
	}
	return result
}

func parseLargestSrcsetURL(srcsetValue string) string {
	type srcsetItem struct {
		url    string
		weight int
	}
	items := make([]srcsetItem, 0)
	for _, part := range strings.Split(srcsetValue, ",") {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) == 0 {
			continue
		}
		weight := 1
		if len(fields) > 1 {
			size := fields[1]
			size = strings.TrimSuffix(size, "x")
			size = strings.TrimSuffix(size, "w")
			fmt.Sscanf(size, "%d", &weight)
		}
		items = append(items, srcsetItem{url: fields[0], weight: weight})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].weight > items[j].weight
	})
	if len(items) == 0 {
		return ""
	}
	return items[0].url
}

func resolveAbsoluteURL(baseURL *neturl.URL, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "data:image/") {
		return raw
	}
	parsed, err := neturl.Parse(raw)
	if err != nil {
		return ""
	}
	return baseURL.ResolveReference(parsed).String()
}

// rewriteMarkdownImageLinks 将已成功下载的远程图片地址替换为本地 images/ 路径。
func rewriteMarkdownImageLinks(markdown string, replacements map[string]string) string {
	result := markdown
	for remoteURL, localPath := range replacements {
		result = strings.ReplaceAll(result, "("+remoteURL+")", "("+localPath+")")
	}
	return result
}

// buildMarkdownImageReplacements 统一生成图片候选地址到本地 images 路径的映射。
func buildMarkdownImageReplacements(resource ScrapeImageResource, localPath string) map[string]string {
	replacements := make(map[string]string)
	for _, candidateURL := range resource.CandidateURLs {
		if candidateURL == "" {
			continue
		}
		replacements[candidateURL] = localPath
	}
	if resource.SourceURL != "" {
		replacements[resource.SourceURL] = localPath
	}
	for _, rawCandidateURL := range resource.RawCandidateURLs {
		if rawCandidateURL == "" {
			continue
		}
		replacements[rawCandidateURL] = localPath
	}
	return replacements
}

func setHTMLAttr(node *html.Node, key, value string) {
	for i := range node.Attr {
		if strings.EqualFold(strings.TrimSpace(node.Attr[i].Key), key) {
			node.Attr[i].Val = value
			return
		}
	}
	node.Attr = append(node.Attr, html.Attribute{Key: key, Val: value})
}

// normalizeHTMLImageSources makes markdown conversion use each image node's best
// lazy-loaded source instead of a shared placeholder src.
func normalizeHTMLImageSources(htmlText string) (string, error) {
	contextNode := &html.Node{Type: html.ElementNode, DataAtom: atom.Div, Data: "div"}
	nodes, err := html.ParseFragment(strings.NewReader(htmlText), contextNode)
	if err != nil {
		return "", err
	}
	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "img") {
			rawCandidateURLs := collectRawImageCandidateURLs(node)
			if len(rawCandidateURLs) > 0 {
				setHTMLAttr(node, "src", rawCandidateURLs[0])
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	for _, node := range nodes {
		walk(node)
	}

	buffer := bytes.NewBuffer(nil)
	for _, node := range nodes {
		if err := html.Render(buffer, node); err != nil {
			return "", err
		}
	}
	return buffer.String(), nil
}

// convertHTMLToMarkdown 将抓取到的 HTML 统一转为 Markdown。
func convertHTMLToMarkdown(htmlText string) (string, error) {
	converter := md.NewConverter("", true, nil)
	normalizedHTML, err := normalizeHTMLImageSources(htmlText)
	if err != nil {
		return "", err
	}
	return converter.ConvertString(normalizedHTML)
}

// buildImageFileName 按下载顺序和响应类型生成稳定文件名。
func buildImageFileName(index int, contentType, rawURL string) string {
	ext := ".bin"
	if contentType != "" {
		if mediaType, _, err := mime.ParseMediaType(contentType); err == nil {
			if exts, _ := mime.ExtensionsByType(mediaType); len(exts) > 0 {
				ext = exts[0]
			}
		}
	}
	if ext == ".bin" {
		urlObj, err := neturl.Parse(rawURL)
		if err == nil {
			if fromURL := filepath.Ext(urlObj.Path); fromURL != "" {
				ext = fromURL
			}
		}
	}
	return fmt.Sprintf("image_%03d%s", index, ext)
}

// buildScrapeZip 打包 HTML、Markdown 与图片资源，供 Agent 上传回服务端。
func buildScrapeZip(markdown, htmlText string, imageFiles map[string][]byte) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	zipWriter := zip.NewWriter(buffer)

	contentWriter, err := zipWriter.Create("content.md")
	if err != nil {
		return nil, err
	}
	if _, err = contentWriter.Write([]byte(markdown)); err != nil {
		return nil, err
	}

	htmlWriter, err := zipWriter.Create("content.html")
	if err != nil {
		return nil, err
	}
	if _, err = htmlWriter.Write([]byte(htmlText)); err != nil {
		return nil, err
	}

	imageKeys := make([]string, 0, len(imageFiles))
	for key := range imageFiles {
		imageKeys = append(imageKeys, key)
	}
	sort.Strings(imageKeys)
	for _, key := range imageKeys {
		writer, createErr := zipWriter.Create(path.Join("images", key))
		if createErr != nil {
			return nil, createErr
		}
		if _, createErr = writer.Write(imageFiles[key]); createErr != nil {
			return nil, createErr
		}
	}

	if err = zipWriter.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// downloadImage 下载图片并返回内容与文件名。
func downloadImage(client *http.Client, imageURL, userAgent, referer string, cookies []*http.Cookie, index int) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	for _, cookie := range cookies {
		if cookie != nil {
			req.AddCookie(cookie)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("download image status=%d", resp.StatusCode)
	}
	content, err := ioReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return content, buildImageFileName(index, resp.Header.Get("Content-Type"), imageURL), nil
}

func ioReadAll(reader io.Reader) ([]byte, error) {
	return io.ReadAll(reader)
}

// RunScrapeToMarkdown 在 Agent 本地浏览器上下文中执行抓取、转 Markdown 和 ZIP 打包。
func RunScrapeToMarkdown(runParams *PlaywrightRunParams, scrapeConfig define.AgentTaskScrapeConfig, log *gstool.GsSlog) (*ScrapeMarkdownResult, error) {
	if runParams == nil {
		return nil, fmt.Errorf("runParams不能为空")
	}
	if scrapeConfig.JumpURL == "" {
		return nil, fmt.Errorf("jump_url不能为空")
	}
	if scrapeConfig.CssSelector == "" {
		return nil, fmt.Errorf("css_selector不能为空")
	}
	if runParams.CombineType != define.CombineTypeNo && runParams.CombineType != define.CombineTypeFix {
		return nil, fmt.Errorf("抓取任务暂不支持当前数据目录复用类型=%d，请改为固定目录或每次打开新的", runParams.CombineType)
	}
	logScrapeMarkdownStep(log, runParams, "抓取任务", "开始 jump_url=%s css_selector=%s combine_type=%d process_count=%d", scrapeConfig.JumpURL, scrapeConfig.CssSelector, runParams.CombineType, len(runParams.ProcessList))

	runParams.ProcessList = buildScrapeProcessList(runParams.ProcessList, scrapeConfig.JumpURL)
	logScrapeMarkdownStep(log, runParams, "抓取任务", "已追加跳转节点 process_count=%d", len(runParams.ProcessList))
	playwrightRunner := NewPlaywright(runParams, log)
	page, err := playwrightRunner.GetPage(nil)
	if err != nil {
		logScrapeMarkdownStep(log, runParams, "抓取任务", "获取页面失败 err=%s", err.Error())
		return nil, err
	}
	logScrapeMarkdownStep(log, runParams, "抓取任务", "获取页面成功 current_url=%s", (*page).URL())
	defer func() {
		_ = (*page).Close()
	}()

	for _, processVal := range runParams.ProcessList {
		runParams.StreamFunc(castToString(processVal["name"]), "按顺序执行")
		boolContinue, runErr := playwrightRunner.ProcessRun(processVal, page)
		if runErr != nil {
			return nil, runErr
		}
		if !boolContinue {
			break
		}
	}

	component.PlaywrightClient.WaitForLoadState(page, runParams.LocatorTimeout)
	if scrapeConfig.WaitSeconds > 0 {
		logScrapeMarkdownStep(log, runParams, "抓取等待", "额外等待 %d 秒", scrapeConfig.WaitSeconds)
		time.Sleep(time.Duration(scrapeConfig.WaitSeconds) * time.Second)
	}

	locator := (*page).Locator(scrapeConfig.CssSelector)
	count, err := locator.Count()
	if err != nil {
		logScrapeMarkdownStep(log, runParams, "定位抓取节点", "统计节点失败 err=%s", err.Error())
		return nil, err
	}
	logScrapeMarkdownStep(log, runParams, "定位抓取节点", "节点数量=%d selector=%s", count, scrapeConfig.CssSelector)
	if count <= 0 {
		return nil, fmt.Errorf("未找到抓取节点: %s", scrapeConfig.CssSelector)
	}
	htmlText, err := locator.First().InnerHTML()
	if err != nil {
		logScrapeMarkdownStep(log, runParams, "抓取HTML", "读取HTML失败 err=%s", err.Error())
		return nil, err
	}
	logScrapeMarkdownStep(log, runParams, "抓取HTML", "读取HTML成功 html_len=%d", len(htmlText))

	imageResources, err := extractImageSources(htmlText, scrapeConfig.JumpURL)
	if err != nil {
		logScrapeMarkdownStep(log, runParams, "解析图片", "提取图片资源失败 err=%s", err.Error())
		return nil, err
	}
	logScrapeMarkdownStep(log, runParams, "解析图片", "提取到图片资源 %d 个", len(imageResources))
	userAgent, _ := readPageUserAgent(page)
	imageFiles := make(map[string][]byte)
	replacements := make(map[string]string)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	skippedDataImageCount := 0
	for index, resource := range imageResources {
		if strings.HasPrefix(resource.SourceURL, "data:image/") {
			skippedDataImageCount++
			continue
		}
		cookies, _ := buildHTTPCookies((*page).Context(), resource.SourceURL)
		content, fileName, downloadErr := downloadImage(httpClient, resource.SourceURL, userAgent, scrapeConfig.JumpURL, cookies, index+1)
		if downloadErr != nil {
			logScrapeMarkdownStep(log, runParams, "下载图片", "跳过失败图片 url=%s err=%s", resource.SourceURL, downloadErr.Error())
			continue
		}
		imageFiles[fileName] = content
		localPath := path.Join("images", fileName)
		for sourceURL, replacePath := range buildMarkdownImageReplacements(resource, localPath) {
			replacements[sourceURL] = replacePath
		}
		logScrapeMarkdownStep(log, runParams, "下载图片", "下载成功 url=%s file_name=%s size=%d", resource.SourceURL, fileName, len(content))
	}
	logScrapeMarkdownStep(log, runParams, "下载图片", "图片下载结束 success_count=%d replacement_count=%d skipped_data_count=%d", len(imageFiles), len(replacements), skippedDataImageCount)

	markdown, err := convertHTMLToMarkdown(htmlText)
	if err != nil {
		logScrapeMarkdownStep(log, runParams, "转换Markdown", "HTML转Markdown失败 err=%s", err.Error())
		return nil, err
	}
	markdown = rewriteMarkdownImageLinks(markdown, replacements)
	logScrapeMarkdownStep(log, runParams, "转换Markdown", "HTML转Markdown成功 markdown_len=%d", len(markdown))
	zipBytes, err := buildScrapeZip(markdown, htmlText, imageFiles)
	if err != nil {
		logScrapeMarkdownStep(log, runParams, "打包ZIP", "打包失败 err=%s", err.Error())
		return nil, err
	}
	fileName := fmt.Sprintf("scrape_result_%d.zip", time.Now().Unix())
	logScrapeMarkdownStep(log, runParams, "打包ZIP", "打包成功 zip_bytes=%d file_name=%s", len(zipBytes), fileName)
	return &ScrapeMarkdownResult{
		Markdown: []byte(markdown),
		ZipBytes: zipBytes,
		FileName: fileName,
	}, nil
}

func readPageUserAgent(page *playwright.Page) (string, error) {
	value, err := (*page).Evaluate("() => navigator.userAgent")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", value), nil
}

func buildHTTPCookies(context playwright.BrowserContext, targetURL string) ([]*http.Cookie, error) {
	items, err := context.Cookies(targetURL)
	if err != nil {
		return nil, err
	}
	result := make([]*http.Cookie, 0, len(items))
	for _, item := range items {
		result = append(result, &http.Cookie{
			Name:     item.Name,
			Value:    item.Value,
			Path:     item.Path,
			Domain:   item.Domain,
			Secure:   item.Secure,
			HttpOnly: item.HttpOnly,
		})
	}
	return result, nil
}

func castToString(value any) string {
	return fmt.Sprintf("%v", value)
}
