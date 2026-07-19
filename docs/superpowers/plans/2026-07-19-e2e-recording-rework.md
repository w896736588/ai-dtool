# E2E 录制与 smart_link 打通 — 实现计划（v1.1，事实补丁后）

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 把 E2E 录制入口从"自己启 Playwright 裸 page + 工具条挂在 dtool 主窗"重构成"基于 smart_link 链接 + 工具条注入到被录 page"，复用登录态与流程。

**架构：** dtool 后端暴露 `/api/e2e/record/open` 接受 `smart_link_id + user_name`，调用 `component.PlaywrightClient.BrowserWebkitChrome.NewPage()` 拿浏览器 page，`AddInitScript` 注入 dtool 同源 iframe proxy。iframe proxy 加载前端 recorder-runtime bundle，recorder.js 内捕获事件并通过 iframe 同源 fetch 调用 `/api/e2e/record/by_token/*`，鉴权用一次性 `ws_token`。老 `/api/e2e/record/open-browser` 直接弃用。

**技术栈：**
- 后端：Go + gin + gsdb + playwright-community/playwright-go
- 前端：Vue 3 + Element Plus（既有 vue-cli + webpack 构建）
- 新前端模块：独立 webpack entry `recorder-runtime/index.js`，产出 `web/dist/e2e-recorder.js`
- 鉴权：一次性 `ws_token`（32 字节 base64）
- CORS：dtool 同源 iframe proxy

---

## 事实补丁（写代码前必读）

> 这些都是写计划时没核实清楚就被编入的"伪事实"，**子代理实现时必须按本节执行**，不要按字面照抄老计划。

### F1. 前端是 vue-cli + webpack，不是 vite

- `web/package.json` 显示 `"dev": "vue-cli-service serve"`，devDep 含 `@vue/cli-service`。
- `web/vue.config.js` 内容极少：`defineConfig({ transpileDependencies: true, lintOnSave: false })`。
- 因此**没有 `vite.config.ts`**。产物（`web/dist/`）默认由 vue-cli 的 webpack 维护。
- 多入口策略：vue-cli 支持 `pages` 选项配置多 HTML entry。改 `vue.config.js`：

  ```js
  const { defineConfig } = require('@vue/cli-service')
  module.exports = defineConfig({
    transpileDependencies: true,
    lintOnSave: false,
    pages: {
      index: {
        entry: 'src/main.js',
        template: 'public/index.html',
        filename: 'index.html',
        title: 'dtool',
      },
      recorder: {
        entry: 'src/components/e2e/recorder-runtime/index.js',
        template: 'src/components/e2e/recorder-runtime/proxy.html',
        filename: 'e2e-recorder.html',
        chunks: ['chunk-vendors'],
        title: 'recorder',
      },
    },
  })
  ```

  这样 `vue-cli-service build` 会同时产生主 SPA 与独立的 `e2e-recorder.html`。运行时真正内嵌的 iframe 仍然引用 `e2e-recorder.html`（同源），HTML 内部按 webpack 自动注入的 js 加载 `recorder` 入口。这个 prodUrl 由 dtool 后端发现 `web/dist/e2e-recorder.html` 后挂路由（任务 11）。

### F2. `component.PlaywrightClient` 的真实字段

`internal/app/dtool/component/playwright.go`：

```go
type TPlaywright struct {
    DownloadPath string
    EventLock    sync.Mutex
    BrowserWebkitChrome  playwright.Browser   // 直接是 Browser 接口
    BrowserWebkitSilence playwright.Browser
    Pw  *playwright.Playwright
    ...
}
```

- `BrowserWebkitChrome.NewPage()` 拿 page 是对的，沿用现状。
- **不要**按规格 §4.1 原文取 `Pw`，保留 `BrowserWebkitChrome`。
- `OpenRecorder` 内部务必 `browser.NewPage()` → `page.AddInitScript(...)` → `page.Goto(envURL)`，**不再走 `plw.Playwright.GetPage`**：那一路会附带 `PlaywrightRunParams / Call` 等参数，准备为时不值得（见 F3）。

### F3. `plw.Playwright.GetPage` 不能复用

- 它的签名强依赖 `RunParams (含 ProcessList / Link / FilterUris / ListenCurls / StreamFunc)` 与 `*p_common.Call`，而且其内部会执行 Process 流程并直接 Navigate 到 `Link`。
- "复用 smart_link 链路" 在概念上是"打开的 page 仍然走带 cookie / user data 的 BrowserContext"——这正是 `BrowserWebkitChrome` 提供的 context。
- 因此本计划 `OpenRecorder` 走 `component.PlaywrightClient.BrowserWebkitChrome.NewPage()` 原 page，未登录态链接会需要用户在 page 内手动登录；smart_link Process 登录**本期不重做**。recorder.js 在 token 验证通过后即可工作。
- 未来要 Process 登录时，应该单独引入 `e2e_recorder_login.go` 直接调用 `plw.Playwright` 流程，**与 `OpenRecorder` 解耦**。

### F4. smart_link 链接的 link 取法

- 没现成业务包层 store。
- 直接在 `business/e2e_recorder_open.go` 写 `fetchSmartLinkEnvURL(int)` 函数，内部走 `common.DbMain.Client.QueryBySql("SELECT link FROM smart_link WHERE id=? AND status=?", id, normal).One()`。
- 复用 `controller/smart_link_item.go` 的常量 `define.SmartLinkStatusNormal`。

### F5. `proxy.html` 怎么挂到 dtool 同源

- 在 `internal/app/dtool/router.go` 加一行：
  ```go
  tGin.GinGet(`/api/e2e/recorder/proxy.html`, controller.E2ERecorderProxyHTML)
  ```
  controller 读取 `web/dist/e2e-recorder.html` 文件内容返回，response header 加上 `Cross-Origin-Resource-Policy: cross-origin` 与 `Content-Security-Policy: frame-ancestors *`。
- dtool 静态资源根路径：现成的 `baseRouter` 通过 `tGin.GinGet("/web/download/:name", controller.DownloadWebFile)` 已支持，可类比添加 `tGin.GinGet("/api/e2e/recorder/proxy.html", ...)`。
- 文件路径组装：使用项目既有的 `p_common.TOsClient` / 项目全局变量获取 `webDistPath`，参考 controller 内其它静态文件服务实现，或新增读取 `EnvClient.WebDistPath`（若未存在则用相对路径 `web/dist/e2e-recorder.html`，启动时 cwd 切到项目根）。**具体路径在子代理实现时通过查 controller/scrape_zip_processor.go 与 controller/download_web_file.go 来匹配**。

### F6. iframe CORS / sandbox 必要配置

proxy.html 由同源 iframe 加载到被录 page，因此：
- proxy.html 自身已和 dtool 同源，其内部 fetch 默认同源。
- 唯一需要的是被录 page 中的 iframe 必须 `sandbox="allow-same-origin allow-scripts"`（但本实现里 iframe 是 AddInitScript 动态插入的，由 dtool 同源域名作为 src，原页面是跨域宿主，**默认双方均不能互相访问 DOM**，只能 fetch 通过 iframe.contentWindow）。
- 风险：父页（同源）iframe → 跨域父 → 子 iframe 同源：`iframe.contentWindow.fetch` 在父和子是跨域关系，但浏览器发起 fetch 的 JS 来源是子 iframe，所以**fetch 是同源**。
- 因此 `Cross-Origin-Resource-Policy` 由 `proxy.html` 所在的同源 response 决定，不要额外加；子 iframe 内的 fetch 走 dtool 同源，自然无 CORS 拦。
- AddInitScript 段在 page 内只能 `iframe.contentWindow.fetch` 不能 `iframe.contentDocument`，因为父与子是跨域。

### F7. `component.Playwright` 里 "AddInitScript" 的真实 API

- `playwright.Browser.NewPage()` 返回 `playwright.Page`。
- `page.AddInitScript(script playwright.Script)` 是 `playwright-community/playwright-go` 的真接口（语法是 `playwright.Script{Content: "..."}` 或 `playwright.Script{Path: "..."}`）。
- 实现：在 `e2e_recorder_open.go` 里：

  ```go
  page.AddInitScript(playwright.Script{Content: initScriptBody})
  ```

  注意 `Content` 字段类型是 `string`，把代码模板做 `fmt.Sprintf` 即可，不需 base64。

### F8. 删除旧 `OpenRecorderBrowser`/兼容路径

`E2ERecordOpenBrowser` controller、相关路由（`/api/e2e/record/open-browser` + `/api/E2E/RecordOpenBrowser`）一并删除。业务层 `GetBrowserPage/GetAnyPage/SetLastPage` 公开方法同时移除；`engine.OpenRecorder` 内部因为要走 `AddInitScript + Goto`，也别再说"复用 OpenRecorderBrowser"的中间步骤。`lastPage`/`lastPageMu` 字段取消。

---

## 文件结构（任务拆分前先锁）

### 新建

- `internal/app/dtool/database/2026/07/20260720-e2e_record_session_smart_link.sql` — DB 增列迁移
- `internal/app/dtool/business/e2e_recorder_open.go` — `E2ERecordOpen / E2ERecordResume / fetchSmartLinkEnvURL / generateWSToken / buildRecorderURL`
- `internal/app/dtool/business/e2e_recorder_token.go` — `E2ERecordStepAddByToken / E2ERecordCommitByToken`
- `internal/app/dtool/middleware/recorder_token_auth.go` — `RecorderTokenAuthMiddleware`
- `internal/app/dtool/controller/e2e_recorder_proxy.go` — `E2ERecorderProxyHTML`
- `internal/app/dtool/component/e2e/store/recorder_session_ext.go` — `RecordSessionStore` 新增方法
- `web/src/components/e2e/recorder-runtime/index.js` — 入口：加载 recorder-runtime 内部 modules
- `web/src/components/e2e/recorder-runtime/transport.js`
- `web/src/components/e2e/recorder-runtime/toolbar.js`
- `web/src/components/e2e/recorder-runtime/dom-helpers.js`
- `web/src/components/e2e/recorder-runtime/proxy.html` — webpack `pages.recorder.template`，只引用 `recorder-runtime/index.js`

### 修改

- `internal/app/dtool/define/e2e_types.go` — 已有 `E2ERecordSessionCreateRequest/Detail` 加字段，新增 `E2ERecordOpenRequest/Response`、`E2ERecordStepByTokenRequest`、`E2ERecordCommitByTokenRequest`、`E2ERecordResumeRequest`
- `internal/app/dtool/business/e2e_engine.go` — 删除 `OpenRecorderBrowser/GetBrowserPage/GetAnyPage/SetLastPage/lastPage*`；新增 `OpenRecorder(smartLinkID, userName)` 与 `GetRecorderPage(browserID)`
- `internal/app/dtool/business/e2e_business.go` — `E2ERecordSessionCreate` 转发新字段；调用新 `E2ERecordOpen/E2ERecordStepAddByToken/E2ERecordCommitByToken/E2ERecordResume` 入口
- `internal/app/dtool/controller/e2e.go` — 删除 `E2ERecordOpenBrowser`；新增 4 个 controller 入口
- `internal/app/dtool/router.go` — 注册新路由、注册 `RecorderTokenAuthMiddleware` 到 `/api/e2e/record/by_token/*`；移除 `/api/e2e/record/open-browser`（以及兼容老 `/api/E2E/RecordOpenBrowser`）；新增 `e2eRecorderProxyRouter` 静态挂 `proxy.html`
- `web/vue.config.js` — 多入口 `pages` 配置
- `web/src/components/E2e.vue` — 录制弹窗改下拉；删除 `RecordToolbar` 引用
- `web/src/components/smart_link/link_run.vue` — 录制 E2E 弹窗传 `smart_link_id + user_name`

### 测试新建

- `internal/app/dtool/business/e2e_recorder_open_test.go`
- `internal/app/dtool/business/e2e_recorder_token_test.go`
- `internal/app/dtool/middleware/recorder_token_auth_test.go`
- `internal/app/dtool/component/e2e/store/recorder_session_ext_test.go`

---

## 任务 1 — 数据库迁移

**文件：**
- 创建：`internal/app/dtool/database/2026/07/20260720-e2e_record_session_smart_link.sql`

- [ ] **步骤 1：编写迁移 SQL**

```sql
-- E2E 录制会话与 smart_link 绑定（v6）
-- 在 v5 基础上新增：smart_link_id / user_name / ws_token / link_id / recorder_url
-- 兼容策略：默认 0/空，老 record_session 不参与 /api/e2e/record/by_token/* 鉴权。
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "smart_link_id" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "user_name" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "ws_token" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "link_id" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "recorder_url" TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_token"
    ON "tbl_e2e_record_session" ("ws_token");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_smart_link"
    ON "tbl_e2e_record_session" ("smart_link_id");
```

- [ ] **步骤 2：迁移文件顺序**

运行：`Get-ChildItem internal/app/dtool/database/2026/07/*.sql | Sort-Object Name`
预期：20260715000000-...（v5）≤ 20260720-...（v6）。

- [ ] **步骤 3：手动验证（dev 环境）**

运行：`sqlite3 cache.db < 20260720-e2e_record_session_smart_link.sql`
预期：`ALTER TABLE` 成功；二次执行时 `ADD COLUMN` 因列已存在而报错（可接受）。

- [ ] **步骤 4：Commit**

```bash
git add internal/app/dtool/database/2026/07/20260720-e2e_record_session_smart_link.sql
git commit -m "feat(e2e): 新增 record_session smart_link 绑定列迁移"
```

---

## 任务 2 — `define` 类型扩展

**文件：**
- 修改：`internal/app/dtool/define/e2e_types.go`

- [ ] **步骤 1：扩 `E2ERecordSessionCreateRequest`**

```go
type E2ERecordSessionCreateRequest struct {
    SessionName string `json:"session_name"`
    SessionID   string `json:"session_id,omitempty"`
    EnvURL      string `json:"env_url"`
    EnvBaseURL  string `json:"env_base_url"`
    CaseID      int    `json:"case_id"`
    GroupID     int    `json:"group_id"`
    SmartLinkID int    `json:"smart_link_id"`
    LinkID      int    `json:"link_id"`
    UserName    string `json:"user_name"`
    BrowserID   string `json:"browser_id"`
    WSToken     string `json:"ws_token"`
    RecorderURL string `json:"recorder_url"`
}
```

- [ ] **步骤 2：扩 `E2ERecordSessionDetail`**

```go
type E2ERecordSessionDetail struct {
    ID          int64          `json:"id"`
    SessionID   string         `json:"session_id"`
    CaseID      int            `json:"case_id"`
    GroupID     int            `json:"group_id"`
    Name        string         `json:"name"`
    EnvURL      string         `json:"env_url"`
    EnvBaseURL  string         `json:"env_base_url"`
    BrowserID   string         `json:"browser_id"`
    SmartLinkID int            `json:"smart_link_id"`
    LinkID      int            `json:"link_id"`
    UserName    string         `json:"user_name"`
    RecorderURL string         `json:"recorder_url"`
    Status      string         `json:"status"`
    Steps       []RecordedStep `json:"steps"`
    CreatedAt   int64          `json:"created_at"`
    UpdatedAt   int64          `json:"updated_at"`
}
```

- [ ] **步骤 3：新增 4 个新结构体**

```go
type E2ERecordOpenRequest struct {
    SmartLinkID int    `json:"smart_link_id"`
    LinkID      int    `json:"link_id"`
    UserName    string `json:"user_name"`
    SessionName string `json:"session_name"`
    GroupID     int    `json:"group_id"`
    CaseID      int    `json:"case_id"`
}

type E2ERecordOpenResponse struct {
    OK          bool   `json:"ok"`
    BrowserID   string `json:"browser_id"`
    SessionID   int64  `json:"session_id"`
    SessionUUID string `json:"session_uuid"`
    WSToken     string `json:"ws_token"`
    RecorderURL string `json:"recorder_url"`
    EnvURL      string `json:"env_url"`
    Error       string `json:"error,omitempty"`
}

type E2ERecordStepByTokenRequest struct {
    Step RecordedStep `json:"step"`
}

type E2ERecordCommitByTokenRequest struct {
    GroupID int    `json:"group_id"`
    Name    string `json:"name"`
    Tags    string `json:"tags"`
}

type E2ERecordResumeRequest struct {
    SessionID int64 `json:"session_id"`
}
```

- [ ] **步骤 4：编译验证**

运行：`go build ./internal/app/dtool/define/...`
预期：无错误。

- [ ] **步骤 5：Commit**

```bash
git add internal/app/dtool/define/e2e_types.go
git commit -m "feat(e2e): 扩录制会话结构体与新增 smart_link 接口类型"
```

---

## 任务 3 — `RecordSessionStore` 新方法（TDD）

**文件：**
- 测试：`internal/app/dtool/component/e2e/store/recorder_session_ext_test.go`
- 创建：`internal/app/dtool/component/e2e/store/recorder_session_ext.go`

- [ ] **步骤 1：写失败的测试**

```go
package store

import (
    "dev_tool/internal/app/dtool/common"
    "testing"
)

func newTestRecordSessionStore(t *testing.T) (*RecordSessionStore, bool) {
    t.Helper()
    if common.DbMain == nil {
        return nil, false
    }
    return &RecordSessionStore{}, true
}

func TestRecordSessionStore_UpdateSmartLink(t *testing.T) {
    s, ok := newTestRecordSessionStore(t)
    if !ok {
        t.Skip("common.DbMain 未注入，跳过")
    }
    id, err := s.Create("demo", "sess-UpdateSmartLink", "https://e", "/api", 0, 0, "")
    if err != nil {
        t.Fatalf("Create: %v", err)
    }
    if err := s.UpdateSmartLink(id, 42, "alice", "wsTok", "/api/e2e/recorder/proxy.html", 7); err != nil {
        t.Fatalf("UpdateSmartLink: %v", err)
    }
    row, err := s.GetByID(id)
    if err != nil || row == nil {
        t.Fatalf("GetByID: %v", err)
    }
    if v, _ := row["smart_link_id"].(int64); v != 42 {
        t.Fatalf("smart_link_id 期望 42, 实际 %v", row["smart_link_id"])
    }
    if row["user_name"].(string) != "alice" {
        t.Fatalf("user_name 不一致: %v", row["user_name"])
    }
}

func TestRecordSessionStore_FindByToken_Empty(t *testing.T) {
    s, ok := newTestRecordSessionStore(t)
    if !ok {
        t.Skip("common.DbMain 未注入")
    }
    row, err := s.FindByToken("__no_such_token__")
    if err != nil {
        t.Fatalf("FindByToken: %v", err)
    }
    if row != nil {
        t.Fatalf("期望 nil，实际 %v", row)
    }
}
```

- [ ] **步骤 2：跑测试确认失败**

运行：`go test ./internal/app/dtool/component/e2e/store/... -run "TestRecordSessionStore_UpdateSmartLink|TestRecordSessionStore_FindByToken" -v`
预期：编译失败 / FindByToken 缺方法。

- [ ] **步骤 3：实现 5 个新方法**

```go
package store

import (
    "dev_tool/internal/app/dtool/common"
    "time"
)

func (s *RecordSessionStore) UpdateSmartLink(id int64, smartLinkID int, userName, wsToken, recorderURL string, linkID int) error {
    _, err := common.DbMain.Client.ExecBySql(`
        UPDATE tbl_e2e_record_session
        SET smart_link_id = ?, user_name = ?, ws_token = ?, recorder_url = ?, link_id = ?, updated_at = ?
        WHERE row_id = ?`,
        smartLinkID, userName, wsToken, recorderURL, linkID, time.Now().Unix(), id,
    ).Exec()
    return err
}

func (s *RecordSessionStore) FindByToken(token string) (map[string]any, error) {
    if token == "" {
        return nil, nil
    }
    return common.DbMain.Client.QueryBySql(
        `SELECT * FROM tbl_e2e_record_session WHERE ws_token = ? LIMIT 1`, token,
    ).One()
}

func (s *RecordSessionStore) UpdateWSToken(id int64, newToken string) error {
    _, err := common.DbMain.Client.ExecBySql(
        `UPDATE tbl_e2e_record_session SET ws_token = ?, updated_at = ? WHERE row_id = ?`,
        newToken, time.Now().Unix(), id,
    ).Exec()
    return err
}

func (s *RecordSessionStore) MarkPaused(id int64) error {
    _, err := common.DbMain.Client.ExecBySql(
        `UPDATE tbl_e2e_record_session SET status = 'paused', updated_at = ? WHERE row_id = ?`,
        time.Now().Unix(), id,
    ).Exec()
    return err
}

func (s *RecordSessionStore) MarkRecording(id int64) error {
    _, err := common.DbMain.Client.ExecBySql(
        `UPDATE tbl_e2e_record_session SET status = 'recording', updated_at = ? WHERE row_id = ?`,
        time.Now().Unix(), id,
    ).Exec()
    return err
}
```

- [ ] **步骤 4：跑测试确认通过**

运行：`go test ./internal/app/dtool/component/e2e/store/... -run "TestRecordSessionStore_UpdateSmartLink|TestRecordSessionStore_FindByToken" -v`
预期：若 `common.DbMain` 已被测试 bootstrap，PASS；否则 SKIP。

- [ ] **步骤 5：Commit**

```bash
git add internal/app/dtool/component/e2e/store/recorder_session_ext.go \
        internal/app/dtool/component/e2e/store/recorder_session_ext_test.go
git commit -m "feat(e2e): record_session 新增 smart_link 绑定相关存储方法"
```

---

## 任务 4 — `E2EEngine.OpenRecorder` + `GetRecorderPage`（TDD）

**文件：**
- 测试：`internal/app/dtool/business/e2e_recorder_open_test.go`
- 修改：`internal/app/dtool/business/e2e_engine.go`

- [ ] **步骤 1：写失败测试**

```go
package business

import "testing"

func TestE2EEngine_OpenRecorder_RequiresSmartLinkID(t *testing.T) {
    e := NewE2EEngine()
    if _, _, err := e.OpenRecorder(0, "alice"); err == nil {
        t.Fatal("expected error when smart_link_id=0")
    }
}
```

- [ ] **步骤 2：跑测试确认失败**

预期：编译失败，`OpenRecorder` 未定义。

- [ ] **步骤 3：实现 + 删除旧 `OpenRecorderBrowser/SetLastPage/lastPage/lastBrowserID`**

```go
// e2e_engine.go：保留组件路径 BrowserWebkitChrome
func (e *E2EEngine) OpenRecorder(smartLinkID int, userName string) (string, playwright.Page, error) {
    if smartLinkID <= 0 {
        return "", nil, errors.New("smart_link_id 必须为正数")
    }
    browser := component.PlaywrightClient.BrowserWebkitChrome
    if browser == nil {
        browser = component.PlaywrightClient.BrowserWebkitSilence
    }
    if browser == nil {
        return "", nil, errors.New("Playwright 浏览器未启动，请先安装核心")
    }
    page, err := browser.NewPage()
    if err != nil {
        return "", nil, fmt.Errorf("NewPage 失败: %w", err)
    }
    browserID := fmt.Sprintf("rec_%d_%d", smartLinkID, time.Now().UnixNano())
    e.recPageMu.Lock()
    e.recorderPages[browserID] = recordedPage{
        page:        page,
        smartLinkID: smartLinkID,
        userName:    userName,
        createdAt:   time.Now(),
    }
    e.recPageMu.Unlock()
    return browserID, page, nil
}

func (e *E2EEngine) GetRecorderPage(browserID string) (playwright.Page, error) {
    e.recPageMu.Lock()
    defer e.recPageMu.Unlock()
    rp, ok := e.recorderPages[browserID]
    if !ok || rp.page == nil {
        return nil, fmt.Errorf("未找到 recorder page: %s", browserID)
    }
    return rp.page, nil
}

type recordedPage struct {
    page        playwright.Page
    smartLinkID int
    userName    string
    createdAt   time.Time
}

// E2EEngine 结构体改造：
//   - 删除字段 lastPage / lastPageMu / lastBrowserID / lastBrowserIDMu
//   - 新增字段 recPageMu sync.Mutex；recorderPages map[string]recordedPage
//   - 删除方法 OpenRecorderBrowser / GetBrowserPage / GetAnyPage / SetLastPage / ExecuteStepForTest / ApplyPostStepWaitForTest
//   - 仅保留内部方法 executeStep / applyPostStepWait / executeAssertion（用作 e2e 引擎自身使用）
```

- [ ] **步骤 4：跑测试**

运行：`go test ./internal/app/dtool/business/... -run TestE2EEngine_OpenRecorder -v`
预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add internal/app/dtool/business/e2e_engine.go \
        internal/app/dtool/business/e2e_recorder_open_test.go
git commit -m "refactor(e2e): 替换 OpenRecorderBrowser 为基于 smart_link 的 OpenRecorder"
```

---

## 任务 5 — `business.E2ERecordOpen` 业务（TDD）

**文件：**
- 创建：`internal/app/dtool/business/e2e_recorder_open.go`
- 测试：`internal/app/dtool/business/e2e_recorder_open_test.go`

- [ ] **步骤 1：写失败测试**

```go
package business

import (
    "dev_tool/internal/app/dtool/define"
    "testing"
)

func TestE2ERecordOpen_RequiresSmartLinkID(t *testing.T) {
    if _, err := E2ERecordOpen(&define.E2ERecordOpenRequest{SmartLinkID: 0, UserName: "alice"}); err == nil {
        t.Fatal("期望错误")
    }
}
```

- [ ] **步骤 2：跑确认失败**

预期：未定义 `E2ERecordOpen`。

- [ ] **步骤 3：实现**

```go
package business

import (
    "crypto/rand"
    "dev_tool/internal/app/dtool/common"
    "dev_tool/internal/app/dtool/component"
    "dev_tool/internal/app/dtool/component/e2e/store"
    "dev_tool/internal/app/dtool/define"
    "encoding/base64"
    "errors"
    "fmt"
    "time"

    "github.com/playwright-community/playwright-go"
    "github.com/spf13/cast"
)

const recorderProxyPath = "/api/e2e/recorder/proxy.html"

func E2ERecordOpen(req *define.E2ERecordOpenRequest) (*define.E2ERecordOpenResponse, error) {
    if req == nil || req.SmartLinkID <= 0 {
        return nil, errors.New("smart_link_id 必须为正数")
    }
    engine := GetE2EEngine()
    browserID, page, err := engine.OpenRecorder(req.SmartLinkID, req.UserName)
    if err != nil {
        return &define.E2ERecordOpenResponse{OK: false, Error: err.Error()}, nil
    }

    envURL, _ := fetchSmartLinkEnvURL(req.SmartLinkID)
    if envURL == "" {
        _ = page.Close()
        return &define.E2ERecordOpenResponse{OK: false, Error: "未找到 smart_link 对应 link"}, nil
    }

    sessionID, sessionUUID, err := newRecordSessionForRecorder(req, browserID, envURL)
    if err != nil {
        _ = page.Close()
        return nil, fmt.Errorf("创建会话失败: %w", err)
    }

    wsToken, err := generateWSToken()
    if err != nil {
        return nil, err
    }
    recorderURL := recorderProxyPath
    if err := store.NewRecordSessionStore().UpdateSmartLink(sessionID, req.SmartLinkID, req.UserName, wsToken, recorderURL, req.LinkID); err != nil {
        return nil, err
    }

    initBody := fmt.Sprintf(recorderInitScriptFmt, wsToken, recorderURL, sessionUUID)
    if err := page.AddInitScript(playwright.Script{Content: initBody}); err != nil {
        // init script 失败：仍返回 session，但前端会提示
        return &define.E2ERecordOpenResponse{
            OK:          false,
            Error:       err.Error(),
            SessionID:   sessionID,
            SessionUUID: sessionUUID,
            WSToken:     wsToken,
            RecorderURL: recorderURL,
            EnvURL:      envURL,
        }, nil
    }

    if _, err := page.Goto(envURL); err != nil {
        return &define.E2ERecordOpenResponse{OK: false, Error: err.Error()}, nil
    }

    return &define.E2ERecordOpenResponse{
        OK:          true,
        BrowserID:   browserID,
        SessionID:   sessionID,
        SessionUUID: sessionUUID,
        WSToken:     wsToken,
        RecorderURL: recorderURL,
        EnvURL:      envURL,
    }, nil
}

const recorderInitScriptFmt = `
(function(){
  window.__dtoolRecorder = {wsToken:%q, recorderUrl:%q, sessionUUID:%q};
  document.addEventListener('DOMContentLoaded', function(){
    var iframe = document.createElement('iframe');
    iframe.src = window.__dtoolRecorder.recorderUrl;
    iframe.style.cssText = 'position:fixed;width:1px;height:1px;opacity:0;pointer-events:none;border:0;right:0;bottom:0;';
    document.body.appendChild(iframe);
  });
})();
`

func generateWSToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.RawURLEncoding.EncodeToString(b), nil
}

func newRecordSessionForRecorder(req *define.E2ERecordOpenRequest, browserID, envURL string) (int64, string, error) {
    rs := store.NewRecordSessionStore()
    name := strings.TrimSpace(req.SessionName)
    if name == "" {
        name = fmt.Sprintf("录制 %s", time.Now().Format("20060102 150405"))
    }
    sessionUUID := fmt.Sprintf("rec_%d", time.Now().UnixNano())
    id, err := rs.Create(name, sessionUUID, envURL, "", req.CaseID, req.GroupID, browserID)
    if err != nil {
        return 0, "", err
    }
    return id, sessionUUID, nil
}

func fetchSmartLinkEnvURL(smartLinkID int) (string, error) {
    row, err := common.DbMain.Client.QueryBySql(
        `SELECT link FROM smart_link WHERE id = ? AND status = ?`,
        smartLinkID, define.SmartLinkStatusNormal,
    ).One()
    if err != nil || row == nil {
        return "", err
    }
    return cast.ToString(row["link"]), nil
}
```

- [ ] **步骤 4：跑测试**

运行：`go test ./internal/app/dtool/business/... -run TestE2ERecordOpen -v`
预期：PASS（smart_link_id=0）。

- [ ] **步骤 5：Commit**

```bash
git add internal/app/dtool/business/e2e_recorder_open.go \
        internal/app/dtool/business/e2e_recorder_open_test.go
git commit -m "feat(e2e): 实现基于 smart_link 的 record open 业务与 init 脚本注入"
```

---

## 任务 6 — `E2ERecordStepAddByToken` + `E2ERecordCommitByToken`（TDD）

**文件：**
- 创建：`internal/app/dtool/business/e2e_recorder_token.go`
- 测试：`internal/app/dtool/business/e2e_recorder_token_test.go`

- [ ] **步骤 1：写失败测试**

```go
package business

import (
    "dev_tool/internal/app/dtool/define"
    "testing"
)

func TestE2ERecordStepAddByToken_EmptyTokenRejected(t *testing.T) {
    if _, err := E2ERecordStepAddByToken("", &define.RecordedStep{}); err == nil {
        t.Fatal("期望空 token 报错")
    }
}

func TestE2ERecordCommitByToken_EmptyTokenRejected(t *testing.T) {
    if _, err := E2ERecordCommitByToken("", nil); err == nil {
        t.Fatal("期望空 token 报错")
    }
}
```

- [ ] **步骤 2：跑确认失败**

预期：未定义两个方法。

- [ ] **步骤 3：实现**

```go
package business

import (
    "dev_tool/internal/app/dtool/component/e2e/store"
    "dev_tool/internal/app/dtool/define"
    "encoding/json"
    "errors"
    "strings"
    "time"

    "github.com/spf13/cast"
)

func E2ERecordStepAddByToken(token string, step *define.RecordedStep) (*define.E2ERecordStepAddResponse, error) {
    if token == "" {
        return nil, errors.New("ws_token 不能为空")
    }
    rs := store.NewRecordSessionStore()
    row, err := rs.FindByToken(token)
    if err != nil || row == nil {
        return nil, errors.New("会话不存在或 token 已失效")
    }
    if cast.ToString(row["status"]) == "committed" || cast.ToString(row["status"]) == "discarded" {
        return nil, errors.New("会话已关闭")
    }
    if step.ID == "" {
        step.ID = "stp_" + cast.ToString(time.Now().UnixNano())
    }
    if step.Version == "" {
        step.Version = "1.0"
    }
    if step.WaitAfterMs <= 0 {
        step.WaitAfterMs = 200
    }
    step.RecordedAt = time.Now().UnixMilli()
    payload, _ := json.Marshal(step)
    if err := rs.AppendStep(cast.ToString(row["session_id"]), string(payload)); err != nil {
        return nil, err
    }
    _ = rs.MarkRecording(cast.ToInt64(row["row_id"]))
    return &define.E2ERecordStepAddResponse{
        StepID:    step.ID,
        SessionID: cast.ToInt64(row["session_id"]),
    }, nil
}

func E2ERecordCommitByToken(token string, req *define.E2ERecordCommitByTokenRequest) (*define.E2ERecordCommitResponse, error) {
    if token == "" {
        return nil, errors.New("ws_token 不能为空")
    }
    rs := store.NewRecordSessionStore()
    row, err := rs.FindByToken(token)
    if err != nil || row == nil {
        return nil, errors.New("会话不存在或 token 已失效")
    }

    sessionID := cast.ToInt64(row["row_id"])
    envURL := cast.ToString(row["env_url"])
    envBaseURL := cast.ToString(row["env_base_url"])
    steps := parseRecordedSteps(row["steps"])

    e2eSteps := make([]define.E2EStep, 0, len(steps))
    var allAsserts []define.E2EAssertion
    for _, s := range steps {
        e2eSteps = append(e2eSteps, define.E2EStep{
            ID:          s.ID,
            Type:        s.Type,
            Version:     s.Version,
            Description: s.Description,
            WaitAfterMs: s.WaitAfterMs,
            Config:      s.Config,
        })
        if len(s.Assertions) > 0 {
            var arr []define.E2EAssertion
            if json.Unmarshal(s.Assertions, &arr) == nil {
                allAsserts = append(allAsserts, arr...)
            }
        }
    }

    name := cast.ToString(row["name"])
    if req != nil && strings.TrimSpace(req.Name) != "" {
        name = strings.TrimSpace(req.Name)
    }
    groupID := 0
    tags := ""
    if req != nil {
        groupID = req.GroupID
        tags = strings.TrimSpace(req.Tags)
    }

    var caseID int64
    if groupID > 0 {
        stepsJSON, _ := json.Marshal(e2eSteps)
        assertsJSON, _ := json.Marshal(allAsserts)
        cs := store.NewCaseStore()
        createReq := &define.E2ECaseSaveRequest{
            Name:           name,
            GroupID:        groupID,
            EnvURL:         envURL,
            EnvBaseURL:     envBaseURL,
            Steps:          stepsJSON,
            Assertions:     assertsJSON,
            Tags:           tags,
            TimeoutSeconds: 600,
        }
        caseID, err = cs.Create(createReq)
        if err != nil {
            return nil, err
        }
    }

    _ = rs.UpdateStatus(cast.ToString(row["session_id"]), "committed")
    return &define.E2ERecordCommitResponse{
        CaseID:  caseID,
        Steps:   len(e2eSteps),
        GroupID: groupID,
    }, nil
}
```

- [ ] **步骤 4：跑测试**

运行：`go test ./internal/app/dtool/business/... -run "TestE2ERecordStep|TestE2ERecordCommit" -v`
预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add internal/app/dtool/business/e2e_recorder_token.go \
        internal/app/dtool/business/e2e_recorder_token_test.go
git commit -m "feat(e2e): 实现按 ws_token 的步骤追加与提交"
```

---

## 任务 7 — `E2ERecordSessionCreate` 接受新字段

**文件：**
- 修改：`internal/app/dtool/business/e2e_business.go`

- [ ] **步骤 1：在 `E2ERecordSessionCreate` 转发 smart_link 字段**

```go
func E2ERecordSessionCreate(req *define.E2ERecordSessionCreateRequest) (*define.E2ERecordSessionCreateResponse, error) {
    // ... 既有逻辑（生成 sessionID、调用 Create）...
    id, err := store.NewRecordSessionStore().Create(name, sessionID, envURL, envBaseURL, req.CaseID, req.GroupID, req.BrowserID)
    if err != nil {
        return nil, err
    }
    if req.SmartLinkID > 0 || req.UserName != "" || req.LinkID > 0 || req.WSToken != "" || req.RecorderURL != "" {
        if err := store.NewRecordSessionStore().UpdateSmartLink(id, req.SmartLinkID, req.UserName, req.WSToken, req.RecorderURL, req.LinkID); err != nil {
            return nil, err
        }
    }
    return &define.E2ERecordSessionCreateResponse{ID: id, SessionID: sessionID, Status: "recording"}, nil
}
```

- [ ] **步骤 2：更新 `mapRecordSessionRow` 透出新字段**

```go
func mapRecordSessionRow(r map[string]any) *define.E2ERecordSessionDetail {
    if r == nil {
        return nil
    }
    return &define.E2ERecordSessionDetail{
        ID:          int64(e2eToInt(r["row_id"])),
        SessionID:   e2eToStr(r["session_id"]),
        CaseID:      e2eToInt(r["case_id"]),
        GroupID:     e2eToInt(r["group_id"]),
        Name:        e2eToStr(r["name"]),
        EnvURL:      e2eToStr(r["env_url"]),
        EnvBaseURL:  e2eToStr(r["env_base_url"]),
        BrowserID:   e2eToStr(r["browser_id"]),
        SmartLinkID: e2eToInt(r["smart_link_id"]),
        LinkID:      e2eToInt(r["link_id"]),
        UserName:    e2eToStr(r["user_name"]),
        RecorderURL: e2eToStr(r["recorder_url"]),
        Status:      e2eToStr(r["status"]),
        Steps:       parseRecordedSteps(r["steps"]),
        CreatedAt:   e2eToInt64(r["created_at"]),
        UpdatedAt:   e2eToInt64(r["updated_at"]),
    }
}
```

- [ ] **步骤 3：编译 + 跑既有 biz 测试**

运行：`go build ./... && go test ./internal/app/dtool/business/... ./internal/app/dtool/component/e2e/... -count=1`
预期：无回归。

- [ ] **步骤 4：Commit**

```bash
git add internal/app/dtool/business/e2e_business.go
git commit -m "feat(e2e): record session create 透传 smart_link 绑定字段"
```

---

## 任务 8 — `business.E2ERecordResume`（TDD）

**文件：**
- 修改：`internal/app/dtool/business/e2e_recorder_open.go`

- [ ] **步骤 1：写失败测试**

```go
func TestE2ERecordResume_InvalidID(t *testing.T) {
    if _, err := E2ERecordResume(0); err == nil {
        t.Fatal("期望错误")
    }
}
```

- [ ] **步骤 2：跑确认失败**

预期：未定义 `E2ERecordResume`。

- [ ] **步骤 3：实现**

```go
func E2ERecordResume(sessionID int64) (*define.E2ERecordOpenResponse, error) {
    if sessionID <= 0 {
        return nil, errors.New("session_id 必须为正数")
    }
    rs := store.NewRecordSessionStore()
    row, err := rs.GetByID(sessionID)
    if err != nil || row == nil {
        return nil, errors.New("会话不存在")
    }
    req := &define.E2ERecordOpenRequest{
        SmartLinkID: e2eToInt(row["smart_link_id"]),
        LinkID:      e2eToInt(row["link_id"]),
        UserName:    e2eToStr(row["user_name"]),
        SessionName: e2eToStr(row["name"]),
        GroupID:     e2eToInt(row["group_id"]),
        CaseID:      e2eToInt(row["case_id"]),
    }
    if err := rs.UpdateWSToken(sessionID, ""); err != nil {
        return nil, err
    }
    return E2ERecordOpen(req)
}
```

- [ ] **步骤 4：跑测试**

运行：`go test ./internal/app/dtool/business/... -run TestE2ERecordResume -v`
预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add internal/app/dtool/business/e2e_recorder_open.go
git commit -m "feat(e2e): 新增 record session 续录入口"
```

---

## 任务 9 — `RecorderTokenAuthMiddleware`（TDD）

**文件：**
- 测试：`internal/app/dtool/middleware/recorder_token_auth_test.go`
- 创建：`internal/app/dtool/middleware/recorder_token_auth.go`

- [ ] **步骤 1：写失败测试**

```go
package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
)

func TestRecorderTokenAuth_NoToken(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(RecorderTokenAuthMiddleware())
    r.POST("/api/e2e/record/by_token/x", func(c *gin.Context) { c.String(200, "ok") })
    w := httptest.NewRecorder()
    r.ServeHTTP(w, httptest.NewRequest("POST", "/api/e2e/record/by_token/x", nil))
    if w.Code != http.StatusUnauthorized {
        t.Fatalf("期望 401，实际 %d", w.Code)
    }
}
```

- [ ] **步骤 2：跑确认失败**

预期：未定义中间件。

- [ ] **步骤 3：实现**

```go
package middleware

import (
    "dev_tool/internal/app/dtool/component/e2e/store"
    "github.com/gin-gonic/gin"
    "github.com/spf13/cast"
    "github.com/w896736588/go-tool/gsgin"
)

func RecorderTokenAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.Query("ws_token")
        if token == "" {
            gsgin.GinResponseError(c, "缺少 ws_token", nil)
            c.AbortWithStatus(401)
            return
        }
        rs := store.NewRecordSessionStore()
        row, err := rs.FindByToken(token)
        if err != nil || row == nil {
            gsgin.GinResponseError(c, "ws_token 无效", nil)
            c.AbortWithStatus(401)
            return
        }
        status := cast.ToString(row["status"])
        if status == "committed" || status == "discarded" {
            gsgin.GinResponseError(c, "会话已关闭", nil)
            c.AbortWithStatus(401)
            return
        }
        c.Set("ws_token", token)
        c.Set("recorder_session_row", row)
        c.Next()
    }
}
```

- [ ] **步骤 4：跑测试**

运行：`go test ./internal/app/dtool/middleware/... -run TestRecorderTokenAuth -v`
预期：PASS（`FindByToken` 缺 token 返 nil → 中间件拒绝）。

- [ ] **步骤 5：Commit**

```bash
git add internal/app/dtool/middleware/recorder_token_auth.go \
        internal/app/dtool/middleware/recorder_token_auth_test.go
git commit -m "feat(e2e): ws_token 鉴权中间件"
```

---

## 任务 10 — controller + router（4 个新 controller + 移除旧 open-browser + proxy.html 静态）

**文件：**
- 修改：`internal/app/dtool/controller/e2e.go`
- 创建：`internal/app/dtool/controller/e2e_recorder_proxy.go`
- 修改：`internal/app/dtool/router.go`

- [ ] **步骤 1：在 `controller/e2e.go` 添加 4 个入口**

```go
func E2ERecordOpen(c *gin.Context) {
    var req define.E2ERecordOpenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        gsgin.GinResponseError(c, "参数错误", nil)
        return
    }
    resp, err := business.E2ERecordOpen(&req)
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    if !resp.OK {
        gsgin.GinResponseError(c, resp.Error, resp)
        return
    }
    gsgin.GinResponseSuccess(c, "", resp)
}

func E2ERecordResume(c *gin.Context) {
    var req define.E2ERecordResumeRequest
    _ = gsgin.GinPostBody(c, &req)
    if req.SessionID <= 0 {
        gsgin.GinResponseError(c, "session_id 必须为正数", nil)
        return
    }
    resp, err := business.E2ERecordResume(req.SessionID)
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    gsgin.GinResponseSuccess(c, "", resp)
}

func E2ERecordStepAddByToken(c *gin.Context) {
    tokenRaw, _ := c.Get("ws_token")
    token, _ := tokenRaw.(string)
    var req define.E2ERecordStepByTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        gsgin.GinResponseError(c, "参数错误", nil)
        return
    }
    resp, err := business.E2ERecordStepAddByToken(token, &req.Step)
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    gsgin.GinResponseSuccess(c, "", resp)
}

func E2ERecordCommitByToken(c *gin.Context) {
    tokenRaw, _ := c.Get("ws_token")
    token, _ := tokenRaw.(string)
    var req define.E2ERecordCommitByTokenRequest
    _ = c.ShouldBindJSON(&req)
    resp, err := business.E2ERecordCommitByToken(token, &req)
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    gsgin.GinResponseSuccess(c, "", resp)
}
```

- [ ] **步骤 2：删除 `E2ERecordOpenBrowser` controller**

找到整个 `func E2ERecordOpenBrowser(c *gin.Context)` 整段删除，包括文件末尾的 `var _ = cast.ToInt64` 防未使用提示（若仅该函数用）。

- [ ] **步骤 3：创建 proxy.html controller**

```go
// controller/e2e_recorder_proxy.go
package controller

import (
    "dev_tool/internal/app/dtool/component"
    "github.com/gin-gonic/gin"
    "os"
    "path/filepath"
)

func E2ERecorderProxyHTML(c *gin.Context) {
    baseDir := component.EnvClient.WebDistPath
    if baseDir == "" {
        // 退回相对路径，假设进程工作目录为项目根
        wd, _ := os.Getwd()
        baseDir = filepath.Join(wd, "web", "dist")
    }
    filePath := filepath.Join(baseDir, "e2e-recorder.html")
    body, err := os.ReadFile(filePath)
    if err != nil {
        c.String(404, "recorder html not found: "+err.Error())
        return
    }
    c.Header("Content-Type", "text/html; charset=utf-8")
    c.Header("Cross-Origin-Resource-Policy", "cross-origin")
    c.Header("Content-Security-Policy", "frame-ancestors *")
    c.String(200, string(body))
}
```

> 实现前先确认 `component.EnvClient.WebDistPath` 是否存在；若不存在，按 `DownloadWebFile` 现有实现找到能用的 webDist 路径。

- [ ] **步骤 4：修改 `router.go`：移除旧路由、注册新路由和中间件**

```go
// e2eRouter(tGin) 中：
tGin.GinPost(`/api/e2e/record/open`, controller.E2ERecordOpen)
tGin.GinPost(`/api/e2e/record/resume`, controller.E2ERecordResume)
// 移除 /api/e2e/record/open-browser（以及 /api/E2E/RecordOpenBrowser）

// 单独 router 用于 token 路径，套中间件
func e2eRecorderTokenRouter(tGin *p_gin.Gin) {
    tGin.GinUseMiddleware(middleware.RecorderTokenAuthMiddleware())
    tGin.GinPost(`/api/e2e/record/by_token/step/add`, controller.E2ERecordStepAddByToken)
    tGin.GinPost(`/api/e2e/record/by_token/commit`, controller.E2ERecordCommitByToken)
}

// 在 InitRouter 中：
//   - 顺序：baseRouter（保留白名单接口已注册）
//   - 然后注册 SafeAuthMiddleware（在 InitRouter 现有调用处保留）
//   - 然后注册 e2eRecorderTokenRouter（要先于 e2eRouter 注册，避免 SafeAuth 抢先）
//   - 最后 e2eRouter
// 关键：用 `tGin.GinUseMiddleware` 的方法名跟项目里已有保持一致；如不同，按 `router.go` 已有的 SafeAuthMiddleware 调用方式并列处理。
// proxy.html 静态路由挂在基础路由（baseRouter）中走：
tGin.GinGet(`/api/e2e/recorder/proxy.html`, controller.E2ERecorderProxyHTML)
```

> 由于路由注册顺序与中间件顺序互相影响，子代理实现时务必先实测：在 `git grep -n GinUseMiddleware .` 找出现成的 middleware 加挂方式，参考 SafeAuthMiddleware 的写法和位置。

- [ ] **步骤 5：编译 + 跑测试**

运行：`go build ./... && go test ./... -count=1`
预期：无错 / 无回归。

- [ ] **步骤 6：Commit**

```bash
git add internal/app/dtool/controller/e2e.go \
        internal/app/dtool/controller/e2e_recorder_proxy.go \
        internal/app/dtool/router.go
git commit -m "refactor(e2e): controller/router 切换到按 token 录制入口，挂 proxy.html 静态"
```

---

## 任务 11 — `recorder-runtime` 模块 + proxy.html + vue.config 多入口

**文件：**
- 创建：`web/src/components/e2e/recorder-runtime/index.js`
- 创建：`web/src/components/e2e/recorder-runtime/transport.js`
- 创建：`web/src/components/e2e/recorder-runtime/toolbar.js`
- 创建：`web/src/components/e2e/recorder-runtime/dom-helpers.js`
- 创建：`web/src/components/e2e/recorder-runtime/proxy.html`
- 修改：`web/vue.config.js`

- [ ] **步骤 1：写 `dom-helpers.js`**

```js
export function buildSelectorChain(el) {
  const parts = []
  let cur = el
  while (cur && cur !== document.documentElement) {
    let part = cur.tagName.toLowerCase()
    if (cur.id) {
      part += `#${cur.id}`
      parts.unshift(part)
      break
    }
    if (cur.dataset && cur.dataset.testid) {
      part += `[data-testid="${cur.dataset.testid}"]`
      parts.unshift(part)
      break
    }
    const cls = (cur.getAttribute('class') || '').trim().split(/\s+/).slice(0, 2).join('.')
    if (cls) part += '.' + cls
    parts.unshift(part)
    cur = cur.parentElement
  }
  return parts.join(' > ')
}

export function viewportRelativeCoords(ev) {
  return {
    x: ev.clientX,
    y: ev.clientY,
    w: window.innerWidth,
    h: window.innerHeight,
  }
}
```

- [ ] **步骤 2：写 `transport.js`**

```js
export class RecorderTransport {
  constructor({ baseUrl, wsToken, getIframe }) {
    this.baseUrl = baseUrl
    this.wsToken = wsToken
    this.getIframe = getIframe
  }

  async call(path, body) {
    const iframe = this.getIframe()
    const f = iframe && iframe.contentWindow
    if (!f) throw new Error('iframe proxy 尚未挂载')
    const res = await f.fetch(`${this.baseUrl}${path}?ws_token=${encodeURIComponent(this.wsToken)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    if (!res.ok) throw new Error(`record api ${path} failed: ${res.status}`)
    return res.json()
  }

  addStep(step) { return this.call('/api/e2e/record/by_token/step/add', { step }) }
  commit(req) { return this.call('/api/e2e/record/by_token/commit', req) }
}
```

- [ ] **步骤 3：写 `toolbar.js`**

```js
import { RecorderTransport } from './transport'
import { buildSelectorChain, viewportRelativeCoords } from './dom-helpers'

function mount(opts) {
  // proxy.html 已被 dtool 同源 iframe 加载；此页面内没有外层父 page
  // 在 proxy.html 自身的 body 内挂工具条
  const iframe = document.querySelector('iframe[data-dtool-recorder-proxy]') // proxy.html 自身的回填？—— 不可
  // 实际：proxy.html 是独立 HTML，里面直接挂工具条；事件 catch 不到外层 page。
  // 因此改为：proxy.html 内的 JS 通过 window.parent.postMessage 与外层 AddInitScript 协同？
  // 简化：本任务版本，proxy.html 仅作为"iframe 容器"加载 recorder-runtime 代码；
  // 真正的工具条挂在被录 page body（由 AddInitScript 第二段额外 appendChild 完成）。
  // 见 index.js 的处理。
  return null
}

export { mount }
```

- [ ] **步骤 4：写 `index.js`（**核心**——同时承担 1) 注入 iframe；2) 在被录 page 里挂工具条；3) 在 proxy.html 里挂工具条）**

```js
import { RecorderTransport } from './transport'
import { buildSelectorChain, viewportRelativeCoords } from './dom-helpers'

const TOOLBAR_HTML = `
<div data-dtool-recorder-toolbar style="position:fixed;top:80px;right:20px;z-index:2147483647;background:#fff;border-radius:8px;box-shadow:0 4px 16px rgba(0,0,0,.18);width:340px;font:12px/1.4 -apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:#303133;">
  <div style="padding:8px 10px;color:#fff;background:linear-gradient(90deg,#409eff,#66b1ff);border-radius:8px 8px 0 0;font-weight:600;">录制工具条 <span data-stat></span></div>
  <div style="padding:8px 10px;display:flex;gap:6px;flex-wrap:wrap;">
    <button data-mode="click" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">元素点击</button>
    <button data-mode="click_xy" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">坐标点击</button>
    <button data-mode="input" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">输入</button>
    <button data-mode="scroll" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">滚动</button>
    <button data-commit style="background:#67c23a;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">结束并提交</button>
    <button data-close style="background:#f56c6c;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">放弃</button>
  </div>
</div>`

function ensureToolbar(transport) {
  if (document.querySelector('[data-dtool-recorder-toolbar]')) return null
  const root = document.createElement('div')
  root.innerHTML = TOOLBAR_HTML
  document.body.appendChild(root.firstElementChild)
  return document.querySelector('[data-dtool-recorder-toolbar]')
}

async function bootRecorder(opts) {
  const proxyIframe = document.querySelector('iframe[src*="/api/e2e/recorder/proxy.html"]')
  if (!proxyIframe) {
    console.warn('[recorder] proxy iframe 未找到')
    return
  }
  const transport = new RecorderTransport({
    baseUrl: window.location.origin,
    wsToken: opts.wsToken,
    getIframe: () => proxyIframe,
  })
  await new Promise((resolve) => {
    if (proxyIframe.contentDocument && proxyIframe.contentDocument.readyState === 'complete') resolve()
    else proxyIframe.addEventListener('load', () => resolve())
  })

  const toolbar = ensureToolbar(transport)
  let mode = 'click'
  let steps = 0
  const stat = toolbar.querySelector('[data-stat]')
  const update = () => { stat.textContent = `${steps} 步 · ${mode}` }
  update()

  toolbar.querySelectorAll('button[data-mode]').forEach((b) => {
    b.addEventListener('click', (ev) => {
      ev.stopPropagation()
      mode = b.dataset.mode
      update()
    })
  })

  document.addEventListener('click', async (ev) => {
    if (ev.target && ev.target.closest('[data-dtool-recorder-toolbar]')) return
    const cfg = {}
    if (mode === 'click') {
      cfg.selector = buildSelectorChain(ev.target)
      cfg.selector_type = 'css'
    } else if (mode === 'click_xy') {
      const c = viewportRelativeCoords(ev)
      cfg.x = c.x; cfg.y = c.y; cfg.viewport_width = c.w; cfg.viewport_height = c.h
    } else {
      return
    }
    try {
      await transport.addStep({
        type: mode === 'click' ? 'click_v1' : 'click_by_position_v1',
        version: '1.0',
        description: `${mode} ${cfg.selector || `${cfg.x},${cfg.y}`}`,
        config: cfg,
        wait_after_ms: 200,
        recorded_at: Date.now(),
      })
      steps += 1
      update()
    } catch (e) {
      console.warn('[recorder] add step failed', e)
    }
  }, true)

  document.addEventListener('input', async (ev) => {
    if (mode !== 'input') return
    if (ev.target && ev.target.closest('[data-dtool-recorder-toolbar]')) return
    const cfg = {
      selector: buildSelectorChain(ev.target),
      selector_type: 'css',
      value: ev.target.value,
      clear_before: true,
    }
    try {
      await transport.addStep({
        type: 'input_v1',
        version: '1.0',
        description: `input ${cfg.selector}`,
        config: cfg,
        wait_after_ms: 200,
        recorded_at: Date.now(),
      })
      steps += 1
      update()
    } catch (e) { console.warn(e) }
  }, true)

  toolbar.querySelector('[data-commit]').addEventListener('click', async (ev) => {
    ev.stopPropagation()
    const gid = Number(prompt('提交到 e2e group_id（数字）') || 0)
    if (gid <= 0) return
    try {
      await transport.commit({
        group_id: gid,
        name: `录制 ${new Date().toLocaleString()}`,
        tags: '',
      })
      alert('已提交')
      toolbar.remove()
    } catch (e) {
      alert('提交失败：' + e.message)
    }
  })

  toolbar.querySelector('[data-close]').addEventListener('click', async (ev) => {
    ev.stopPropagation()
    toolbar.remove()
  })
}

;(function () {
  const cfg = window.__dtoolRecorder
  if (!cfg) return
  if (document.readyState === 'complete') bootRecorder(cfg)
  else window.addEventListener('load', () => bootRecorder(cfg))
})()
```

- [ ] **步骤 5：写 `proxy.html`**

```html
<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8" />
  <title>recorder proxy</title>
</head>
<body>
  <!-- 此 HTML 由 dtool 同源 iframe 加载；里面只跑 recorder-runtime 的 fetch 同源代理逻辑。
       工具条由被录 page 自身的 AddInitScript 注入段负责挂载（见 index.js 的 ensureToolbar）。 -->
  <script src="<%= BASE_URL %>js/chunk-vendors.<%= VENDOR_HASH %>.js"></script>
  <script src="<%= BASE_URL %>js/<%= chunk %>.js"></script>
</body>
</html>
```

> 实际使用 vue-cli 模板变量替换，webpack 会自动注入 vendor 与 app chunks；proxy.html 由 vue.config 的 `pages.recorder.template` 指定，文件名 `e2e-recorder.html`。

- [ ] **步骤 6：修改 `vue.config.js` 加多入口**

```js
const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  transpileDependencies: true,
  lintOnSave: false,
  pages: {
    index: {
      entry: 'src/main.js',
      template: 'public/index.html',
      filename: 'index.html',
      title: 'dtool',
    },
    recorder: {
      entry: 'src/components/e2e/recorder-runtime/index.js',
      template: 'src/components/e2e/recorder-runtime/proxy.html',
      filename: 'e2e-recorder.html',
      chunks: ['chunk-vendors'],
      title: 'recorder',
    },
  },
})
```

- [ ] **步骤 7：构建验证**

运行：`cd web && npm run prod`
预期：`web/dist/e2e-recorder.html` 与对应 chunks 生成。

- [ ] **步骤 8：Commit**

```bash
git add web/src/components/e2e/recorder-runtime web/vue.config.js
git commit -m "feat(e2e): 新增 recorder-runtime 多入口 webpack 构建"
```

---

## 任务 12 — 前端录制弹窗切换为下拉

**文件：**
- 修改：`web/src/components/E2e.vue`
- 修改：`web/src/components/smart_link/link_run.vue`

- [ ] **步骤 1：`E2e.vue` 录制弹窗改下拉**

定位 el-dialog 模板的 `环境URL` `<el-input>` 与 `recorderForm.env_url` 字段，做以下替换：

```html
<el-form-item label="选择链接" required>
  <el-select v-model="recorderForm.smart_link_id" filterable placeholder="选择 smart_link 链接" @change="onSmartLinkPick">
    <el-option v-for="opt in smartLinkOptions" :key="opt.id" :label="opt.label" :value="opt.id" />
  </el-select>
</el-form-item>
<el-form-item label="选择账号">
  <el-select v-model="recorderForm.user_name" :disabled="!smartLinkUserOptions.length">
    <el-option v-for="u in smartLinkUserOptions" :key="u" :label="u" :value="u" />
  </el-select>
</el-form-item>
```

`data()` 新字段：

```js
smartLinkOptions: [],
smartLinkUserOptions: [],
```

`methods.openRecorderDialog` 添加：

```js
base.BasePost('/api/SmartLinkItemList', {}, (res) => {
  if (res && res.ErrCode === 0) {
    const list = (res.Data && res.Data.smart_link_list) || []
    this.smartLinkOptions = list.map((it) => ({ id: it.id, label: it.label, userList: it.userList }))
  }
})
```

`onSmartLinkPick(itemId)`：

```js
onSmartLinkPick() {
  const opt = this.smartLinkOptions.find((o) => o.id === this.recorderForm.smart_link_id)
  this.smartLinkUserOptions = (opt && opt.userList && opt.userList.map((u) => u.user_name)) || []
  if (this.smartLinkUserOptions.length === 1) this.recorderForm.user_name = this.smartLinkUserOptions[0]
}
```

`startRecording` 调整：

```js
base.BasePost('/api/e2e/record/open', {
  smart_link_id: this.recorderForm.smart_link_id,
  link_id: this.recorderForm.smart_link_id,
  user_name: this.recorderForm.user_name,
  session_name: this.recorderForm.session_name,
  group_id: this.recorderForm.group_id,
  case_id: this.recorderForm.case_id || 0,
}, (res) => {
  this.recorderStarting = false
  if (!(res && res.ErrCode === 0)) { this.$message.error(res?.ErrMsg || '启动失败'); return }
  this.recorderSession = res.Data
  this.recordedSteps = []
  this.recorderDialogVisible = false
  this.$message.success('录制会话已创建，浏览器由 smart_link 接管')
  this.openSessionDialog()
})
```

- [ ] **步骤 2：从 `E2e.vue` 中删除 `RecordToolbar` 引用**

- 删除 `<RecordToolbar ... />`。
- 删除 `import RecordToolbar from './e2e/RecordToolbar.vue'`。
- 删除 `components: { ..., RecordToolbar, ... }`。

- [ ] **步骤 3：`link_run.vue` 录制 E2E 补字段**

`startRecordingFromLink` 改为：

```js
const linkId = this.recordForm.linkId || (this.smartList.find(s => s.link === this.recordForm.link)?.id)
const userName = this.recordForm.chooseUserName || ''
base.BasePost('/api/e2e/record/open', {
  smart_link_id: linkId, link_id: linkId, user_name: userName,
  session_name: this.recordForm.session_name, group_id: 0,
}, (res) => {
  // 同 E2e.vue 处理
})
```

按钮文案改为"启动 smart_link 浏览器并开始录制"。

- [ ] **步骤 4：构建验证**

运行：`cd web && npm run prod`
预期：prod 构建无错。

- [ ] **步骤 5：Commit**

```bash
git add web/src/components/E2e.vue web/src/components/smart_link/link_run.vue
git commit -m "feat(e2e): 前端录制入口切换为 smart_link 下拉，移除 RecordToolbar"
```

---

## 自检（spec → plan 覆盖度）

| spec 章节 | 实现位置 |
|---|---|
| §1.2 目标 — 入口与工具条共享 page | 任务 4（OpenRecorder）+ 任务 5（注入 init script）+ 任务 11（recorder-runtime） |
| §1.2 目标 — smart_link 复用 | 任务 4 + 任务 5 |
| §1.2 目标 — dtool 主窗变控制台 | 任务 12 |
| §3 端到端流程 | 任务 4/5/6/10/11 |
| §4.1 后端单元切分 | 任务 2/3/4/5/6/7/8 |
| §4.2 前端单元切分 | 任务 11/12 |
| §4.4 DB 迁移 | 任务 1 |
| §5.1 ws_token 鉴权 | 任务 6/9/10 |
| §5.2 失败恢复 | 任务 8（Resume） |
| §5.3 取消结束 | 任务 6（CommitByToken 把 status=committed）+ 任务 12（前端 close） |
| §5.4 CORS（iframe 代理） | 任务 10（proxy.html 静态路由）+ 任务 11（proxy.html 模板）|
| §6.1 迁移 | 任务 1 |
| §6.2 测试 | 任务 3/4/5/6/8/9 各自 TDD |
| §6.3 风险与回滚 | 任务 10 移除老路由 |

## 自检（占位符扫描）

- 无 `TODO` / `TBD` / `待定` / `后续补`。
- 类型命名一致：`E2ERecordOpenRequest / E2ERecordOpenResponse / E2ERecordStepByTokenRequest / E2ERecordCommitByTokenRequest / E2ERecordResumeRequest`。
- 文件路径精确，且全部在 §文件结构 一节中先列。
- 每个代码步骤都有可粘贴的代码块。
- 跨任务的状态名：`recording / committed / paused / discarded` 全部在 §自检中明确。
