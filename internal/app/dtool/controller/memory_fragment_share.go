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
<title>知识片段分享</title><style>body{font-family:system-ui,sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;margin:0;background:#f5f7f2;color:#5f7059;font-size:14px;}</style>
</head><body><p>%s</p></body></html>`, template.HTMLEscapeString(msg))
}

func buildShareHTML(title, updateTime, expireAt, bodyHTML string) string {
	const tpl = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{min-height:100vh;background:#f5f7f2;color:#2f3c2b;font-family:system-ui,-apple-system,sans-serif}
.shell{width:min(960px,calc(100%% - 32px));margin:0 auto;padding:32px 0}
.viewer{min-height:calc(100vh - 64px);border:1px solid #e2e8d8;border-radius:12px;background:#fff;box-shadow:0 8px 24px rgba(54,74,54,.08);overflow:hidden}
.header{padding:24px 28px 18px;border-bottom:1px solid #e8eee0;background:#f8faf5}
.header h1{color:#263523;font-size:24px;line-height:1.35;font-weight:700;word-break:break-word}
.meta{display:flex;gap:12px;flex-wrap:wrap;margin-top:10px;color:#687762;font-size:13px}
.content{padding:22px 28px 32px;font-size:14px;color:#33422f;line-height:1.7}
.content h1,.content h2,.content h3,.content h4,.content h5,.content h6{color:#263523;margin:24px 0 12px;line-height:1.35}
.content h1{font-size:22px;border-bottom:1px solid #e2e8d8;padding-bottom:8px}
.content h2{font-size:19px}
.content h3{font-size:16px}
.content p{margin:10px 0}
.content ul,.content ol{padding-left:24px;margin:10px 0}
.content li{margin:4px 0}
.content code{background:#f0f4ec;padding:2px 6px;border-radius:4px;font-size:13px;font-family:Menlo,Consolas,monospace}
.content pre{background:#f6f8f3;border:1px solid #e2e8d8;border-radius:8px;padding:16px;overflow-x:auto;margin:14px 0}
.content pre code{background:none;padding:0}
.content blockquote{border-left:4px solid #b5c7ad;padding:8px 16px;margin:14px 0;color:#5f7059;background:#f8faf5;border-radius:0 8px 8px 0}
.content table{border-collapse:collapse;width:100%%;margin:14px 0}
.content th,.content td{border:1px solid #e2e8d8;padding:8px 12px;text-align:left}
.content th{background:#f8faf5;font-weight:600}
.content img{max-width:100%%;border-radius:6px}
.content a{color:#3d7a3a;text-decoration:none}
.content a:hover{text-decoration:underline}
.content hr{border:none;border-top:1px solid #e2e8d8;margin:20px 0}
@media(max-width:720px){
  .shell{width:calc(100%% - 20px);padding:10px 0}
  .viewer{min-height:calc(100vh - 20px)}
  .header,.content{padding-left:16px;padding-right:16px}
}
</style>
</head>
<body>
<main class="shell">
<article class="viewer">
<header class="header">
<h1>%s</h1>
<div class="meta">
%s
</div>
</header>
<section class="content">
%s
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

	return fmt.Sprintf(tpl,
		template.HTMLEscapeString(title),
		template.HTMLEscapeString(title),
		metaStr,
		bodyHTML,
	)
}
