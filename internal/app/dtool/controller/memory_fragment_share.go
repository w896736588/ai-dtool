package controller

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"dev_tool/internal/app/dtool/component"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// MemoryFragmentShareCreate 创建一个 24 小时有效的知识片段只读分享 token。
func MemoryFragmentShareCreate(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	fragmentID := strings.TrimSpace(cast.ToString(dataMap[`id`]))
	if fragmentID == `` || fragmentID == `0` {
		gsgin.GinResponseError(c, `片段id不能为空`, nil)
		return
	}
	if _, err := memoryDB.MemoryFragmentInfo(fragmentID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, err := shareStore.Create(fragmentID, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, `创建分享链接失败：`+err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, ``, memoryFragmentShareResponse(share))
}

// MemoryFragmentShareInfo 通过分享 token 读取知识片段详情，只返回查看页所需数据。
func MemoryFragmentShareInfo(c *gin.Context) {
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		gsgin.GinResponseError(c, err.Error(), map[string]any{
			`configured`: false,
		})
		return
	}
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	token := strings.TrimSpace(cast.ToString(dataMap[`token`]))
	if token == `` {
		gsgin.GinResponseError(c, `分享链接不能为空`, nil)
		return
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, ok, err := shareStore.Resolve(token, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, `读取分享链接失败：`+err.Error(), nil)
		return
	}
	if !ok {
		gsgin.GinResponseError(c, `分享链接不存在或已过期`, nil)
		return
	}
	info, err := component.MemoryRuntime.DB().MemoryFragmentInfo(share.FragmentID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`fragment`: info,
		`share`:    memoryFragmentShareResponse(share),
	})
}

// MemoryFragmentSharePage 纯 HTML 分享页面，服务端渲染 markdown 为 HTML 后返回完整页面，方便 AI 直接读取。
func MemoryFragmentSharePage(c *gin.Context) {
	token := strings.TrimSpace(c.Param(`token`))
	if token == `` {
		c.HTML(http.StatusBadRequest, ``, templateHTML(`分享链接缺少 token`))
		return
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		c.HTML(http.StatusInternalServerError, ``, templateHTML(`记忆库未配置：`+err.Error()))
		return
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, ok, err := shareStore.Resolve(token, time.Now())
	if err != nil {
		c.HTML(http.StatusInternalServerError, ``, templateHTML(`读取分享链接失败：`+err.Error()))
		return
	}
	if !ok {
		c.HTML(http.StatusNotFound, ``, templateHTML(`分享链接不存在或已过期`))
		return
	}
	info, err := component.MemoryRuntime.DB().MemoryFragmentInfo(share.FragmentID)
	if err != nil {
		c.HTML(http.StatusInternalServerError, ``, templateHTML(err.Error()))
		return
	}

	title := cast.ToString(info[`title`])
	if title == `` {
		title = `未命名片段`
	}
	content := cast.ToString(info[`content`])
	updateTimeDesc := cast.ToString(info[`update_time_desc`])
	expireAtDesc := gstool.TimeUnixToString(share.ExpireAt, `Y-m-d H:i:s`)

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		c.HTML(http.StatusInternalServerError, ``, templateHTML(`Markdown 渲染失败：`+err.Error()))
		return
	}

	c.Header(`Content-Type`, `text/html; charset=utf-8`)
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write([]byte(buildShareHTML(title, updateTimeDesc, expireAtDesc, buf.String())))
}

func memoryFragmentShareResponse(share memoryFragmentShare) map[string]any {
	return map[string]any{
		`token`:          share.Token,
		`fragment_id`:    share.FragmentID,
		`expire_at`:      share.ExpireAt.Unix(),
		`expire_at_desc`: gstool.TimeUnixToString(share.ExpireAt, `Y-m-d H:i:s`),
	}
}

func templateHTML(msg string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>知识片段分享</title>
</head><body><p>%s</p></body></html>`, template.HTMLEscapeString(msg))
}

func buildShareHTML(title, updateTime, expireAt, bodyHTML string) string {
	const tpl = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{TITLE}}</title>
<style>body{font-size:14px;line-height:1.7;max-width:960px;margin:0 auto;padding:16px;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif}h1{font-size:20px}h2{font-size:18px}h3{font-size:16px}h4{font-size:15px}table{border-collapse:collapse;width:100%;margin:1em 0}th,td{border:1px solid #d0d7de;padding:6px 13px;text-align:left}th{font-weight:600;background-color:#f6f8fa}tr:nth-child(2n){background-color:#f6f8fa}blockquote{border-left:4px solid #d0d7de;padding:0 1em;color:#656d76;margin:0}code{background-color:#f6f8fa;padding:2px 6px;border-radius:4px;font-size:13px}pre{background-color:#f6f8fa;padding:16px;border-radius:6px;overflow-x:auto}pre code{background:none;padding:0}hr{border:none;border-top:1px solid #d0d7de;margin:24px 0}</style>
</head>
<body>
<main>
<article>
<header>
<h1>{{TITLE}}</h1>
<div>
{{META}}
</div>
</header>
<section>
{{BODY}}
</section>
</article>
</main>
</body>
</html>`

	var metaParts []string
	if updateTime != `` {
		metaParts = append(metaParts, fmt.Sprintf(`<span>更新：%s</span>`, template.HTMLEscapeString(updateTime)))
	}
	if expireAt != `` {
		metaParts = append(metaParts, fmt.Sprintf(`<span>链接有效期至：%s</span>`, template.HTMLEscapeString(expireAt)))
	}
	metaStr := strings.Join(metaParts, "\n")

	r := strings.NewReplacer(
		`{{TITLE}}`, template.HTMLEscapeString(title),
		`{{META}}`, metaStr,
		`{{BODY}}`, bodyHTML,
	)
	return r.Replace(tpl)
}
