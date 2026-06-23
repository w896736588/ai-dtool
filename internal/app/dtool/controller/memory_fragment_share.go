package controller

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"dev_tool/internal/app/dtool/component"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// tocItem 表示右侧目录导航中的一个标题条目
type tocItem struct {
	Level int    // 标题层级 1-4
	Text  string // 标题纯文本
	ID    string // HTML 锚点 ID
}

// headingRe 匹配 goldmark 渲染出的 <h1>~<h4> 标签
var headingRe = regexp.MustCompile(`<h([1-4])>([\s\S]*?)</h[1-4]>`)

// stripTagRe 用于去除 HTML 标签，提取纯文本
var stripTagRe = regexp.MustCompile(`<[^>]*>`)

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
// URL 格式：/share/:id/:token，其中 id 为知识片段 ID，token 为分享凭证。
// fragmentID 可能是 DB 整型 ID 或文件路径末段（如 "116" 或 "6f351897-xxx"），
// 通过 path.Base 统一提取末段后与 share 记录中的 FragmentID 末段做匹配校验。
func MemoryFragmentSharePage(c *gin.Context) {
	fragmentID := strings.TrimSpace(c.Param(`id`))
	token := strings.TrimSpace(c.Param(`token`))
	if fragmentID == `` {
		c.HTML(http.StatusBadRequest, ``, templateHTML(`分享链接缺少片段 ID`))
		return
	}
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
	// 校验 URL 中的 fragmentID 与 token 对应的片段一致，防止跨片段分享链接滥用。
	// 使用 path.Base 提取末段比较，兼容 fragmentID 含路径前缀的情况（如 "fragments/xxx-uuid"）。
	if path.Base(share.FragmentID) != path.Base(fragmentID) {
		c.HTML(http.StatusNotFound, ``, templateHTML(`分享链接与片段不匹配`))
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
	// 给渲染后的 HTML 标题添加 ID，并提取目录数据
	annotatedBody, tocItems := annotateHeadings(buf.String())
	_, _ = c.Writer.Write([]byte(buildShareHTML(title, updateTimeDesc, expireAtDesc, annotatedBody, tocItems)))
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

// buildShareHTML 拼接完整的分享页面 HTML，包含右侧目录导航（如有标题）。
func buildShareHTML(title, updateTime, expireAt, bodyHTML string, tocItems []tocItem) string {
	const tpl = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{TITLE}}</title>
<style>
body{font-size:14px;line-height:1.7;max-width:1200px;margin:0 auto;padding:24px 16px;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;background:#f5f7f2;color:#2f3c2b}
h1{font-size:20px}h2{font-size:18px}h3{font-size:16px}h4{font-size:15px}
table{border-collapse:collapse;width:100%;margin:1em 0}th,td{border:1px solid #d0d7de;padding:6px 13px;text-align:left}th{font-weight:600;background-color:#f6f8fa}tr:nth-child(2n){background-color:#f6f8fa}
blockquote{border-left:4px solid #d0d7de;padding:0 1em;color:#656d76;margin:0}
code{background-color:#f6f8fa;padding:2px 6px;border-radius:4px;font-size:13px}
pre{background-color:#f6f8fa;padding:16px;border-radius:6px;overflow-x:auto}pre code{background:none;padding:0}
hr{border:none;border-top:1px solid #d0d7de;margin:24px 0}
.share-main section img{max-width:100%;height:auto;border-radius:6px;cursor:zoom-in;display:block;margin:8px 0}
.img-overlay{display:none;position:fixed;inset:0;z-index:9999;background:rgba(0,0,0,.85);justify-content:center;align-items:center;cursor:zoom-out}
.img-overlay.show{display:flex}
.img-overlay img{max-width:95vw;max-height:95vh;border-radius:6px;box-shadow:0 4px 32px rgba(0,0,0,.5)}
.share-layout{display:flex;gap:24px;align-items:flex-start}
.share-main{flex:1;min-width:0;background:#fff;border:1px solid #e2e8d8;border-radius:12px;box-shadow:0 8px 24px rgba(54,74,54,0.08);overflow:hidden}
.share-main>header{padding:24px 28px 18px;border-bottom:1px solid #e8eee0;background:#f8faf5}
.share-main>header>h1{margin:0;color:#263523;font-size:24px;font-weight:700}
.share-main .meta{display:flex;gap:12px;flex-wrap:wrap;margin-top:10px;color:#687762;font-size:13px}
.share-main>section{padding:22px 28px 32px}
.share-toc{position:sticky;top:24px;width:200px;min-width:200px;max-height:calc(100vh - 48px);overflow-y:auto;border:1px solid #e2e8d8;border-radius:8px;background:#fff;box-shadow:0 2px 8px rgba(54,74,54,0.06);padding:16px 0}
.toc-title{padding:0 16px 12px;font-size:14px;font-weight:600;color:#263523;border-bottom:1px solid #e8eee0}
.toc-nav{padding:8px 0}
.toc-link{display:block;padding:5px 16px;font-size:13px;line-height:1.5;color:#5f7059;text-decoration:none;cursor:pointer;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;transition:color .2s,background .2s}
.toc-link:hover{color:#263523;background:#f0f4eb}
.toc-link.active{color:#2d6a2e;font-weight:500;background:#e8f0e0}
.toc-h1{padding-left:16px}.toc-h2{padding-left:24px}.toc-h3{padding-left:32px}.toc-h4{padding-left:40px}
.share-toc::-webkit-scrollbar{width:4px}.share-toc::-webkit-scrollbar-thumb{background:#c8d4be;border-radius:2px}
@media(max-width:720px){.share-layout{flex-direction:column}.share-toc{display:none}.share-main>header,.share-main>section{padding-left:16px;padding-right:16px}}
</style>
</head>
<body>
<main class="share-layout">
<article class="share-main">
<header>
<h1>{{TITLE}}</h1>
<div class="meta">
{{META}}
</div>
</header>
<section>
{{BODY}}
</section>
</article>
{{TOC}}
</main>
<div class="img-overlay" id="imgOverlay"><img id="imgOverlaySrc" src="" alt=""></div>
<script>
(function(){
var links=document.querySelectorAll('.toc-link');
links.forEach(function(link){link.addEventListener('click',function(e){
e.preventDefault();var id=this.getAttribute('data-id');var t=document.getElementById(id);
if(t){t.scrollIntoView({behavior:'smooth',block:'start'});links.forEach(function(el){el.classList.remove('active')});this.classList.add('active')}
})});
var hs=document.querySelectorAll('h1[id],h2[id],h3[id],h4[id]');
if(hs.length>0){var obs=new IntersectionObserver(function(entries){entries.forEach(function(en){
if(en.isIntersecting){links.forEach(function(el){el.classList.remove('active')});
var a=document.querySelector('.toc-link[data-id="'+en.target.id+'"]');if(a)a.classList.add('active')}
})},{rootMargin:'-80px 0px -60% 0px',threshold:0.1});hs.forEach(function(h){obs.observe(h)})}
// 图片点击放大
var overlay=document.getElementById('imgOverlay');
var overlayImg=document.getElementById('imgOverlaySrc');
document.querySelectorAll('.share-main section img').forEach(function(img){
img.addEventListener('click',function(){overlayImg.src=this.src;overlay.classList.add('show')})
});
overlay.addEventListener('click',function(){overlay.classList.remove('show')});
})();
</script>
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
		`{{TOC}}`, buildTocHTML(tocItems),
	)
	return r.Replace(tpl)
}

// buildTocHTML 根据目录条目列表生成右侧导航栏的 HTML 片段
func buildTocHTML(items []tocItem) string {
	if len(items) == 0 {
		return ``
	}
	var sb strings.Builder
	sb.WriteString(`<aside class="share-toc"><div class="toc-title">目录</div><nav class="toc-nav">`)
	for _, item := range items {
		sb.WriteString(fmt.Sprintf(
			`<a class="toc-link toc-h%d" data-id="%s">%s</a>`,
			item.Level,
			template.HTMLEscapeString(item.ID),
			template.HTMLEscapeString(item.Text),
		))
	}
	sb.WriteString(`</nav></aside>`)
	return sb.String()
}

// annotateHeadings 给 HTML 中的 h1~h4 标签添加递增 ID（h-1, h-2, ...），
// 同时返回目录条目列表用于生成右侧导航。
func annotateHeadings(htmlStr string) (string, []tocItem) {
	var items []tocItem
	counter := 0

	annotated := string(headingRe.ReplaceAllFunc([]byte(htmlStr), func(match []byte) []byte {
		sub := headingRe.FindSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		counter++
		id := fmt.Sprintf("h-%d", counter)
		level := sub[1]
		inner := string(sub[2])
		// 去掉内嵌 HTML 标签，提取纯文本用于目录显示
		text := stripTagRe.ReplaceAllString(inner, "")

		items = append(items, tocItem{
			Level: int(level[0] - '0'),
			Text:  text,
			ID:    id,
		})

		return []byte(fmt.Sprintf(`<h%s id="%s">%s</h%s>`, level, id, inner, level))
	}))

	return annotated, items
}
