package memory

import (
	"fmt"
	"strconv"
	"strings"
)

// NormalizeFragmentTitle 返回片段标题，优先使用 front matter，其次回退到第一个一级标题。
func NormalizeFragmentTitle(frontMatterTitle, content string) string {
	title := strings.TrimSpace(frontMatterTitle)
	if title != `` {
		return title
	}
	title = ExtractFirstH1(content)
	if title != `` {
		return title
	}
	return `未命名片段`
}

// ExtractFirstH1 提取正文中的第一个 Markdown 一级标题。
func ExtractFirstH1(content string) string {
	for _, line := range strings.Split(normalizeLineBreaks(content), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "# ") {
			continue
		}
		return strings.TrimSpace(strings.TrimPrefix(line, "# "))
	}
	return ``
}

// RenderFragmentMarkdown 把片段对象渲染成标准 Markdown 文件内容。
func RenderFragmentMarkdown(fragment Fragment) (string, error) {
	content := strings.TrimSpace(normalizeLineBreaks(fragment.Content))
	title := NormalizeFragmentTitle(fragment.Title, content)
	meta := FrontMatter{
		Title:      title,
		FolderName: NormalizeFolderName(fragment.FolderName),
		CreatedAt:  fragment.CreatedAt.Format(timeLayout),
		UpdatedAt:  fragment.UpdatedAt.Format(timeLayout),
	}
	return fmt.Sprintf(
		"---\ntitle: %s\nfolder_name: %s\ncreated_at: %s\nupdated_at: %s\n---\n\n%s\n",
		renderFrontMatterTitle(meta.Title),
		renderFrontMatterTitle(meta.FolderName),
		meta.CreatedAt,
		meta.UpdatedAt,
		content,
	), nil
}

func normalizeLineBreaks(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return content
}

func renderFrontMatterTitle(title string) string {
	if title == `` {
		return `""`
	}
	if strings.ContainsAny(title, ":\n\r\"'{}[]#,") {
		return strconv.Quote(title)
	}
	return title
}
