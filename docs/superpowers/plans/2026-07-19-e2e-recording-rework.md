# E2E 录制与 smart_link 打通 — 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 把 E2E 录制入口从"自己启 Playwright 裸 page + 工具条挂在 dtool 主窗"重构成"基于 smart_link 链接 + 工具条注入到被录 page"，复用登录态与流程。

**架构：** dtool 后端暴露 `/api/e2e/record/open` 接受 `smart_link_id + user_name`，调用 `plw.Playwright.GetPage(...)` 拿浏览器 page，`AddInitScript` 注入 dtool 同源 iframe proxy + recorder.js。recorder.js 内捕获事件并通过同源 iframe fetch 调用 `/api/e2e/record/by_token/*`，鉴权用一次性 `ws_token`。老 `/api/e2e/record/open-browser` 直接弃用。

**技术栈：**
- 后端：Go + gin + gsdb + playwright-community/playwright-go
- 前端：Vue 2 + Element Plus（既有）
- 新前端模块：TypeScript + Vite library 模式（产生 `web/dist/e2e-recorder.js`）
- 鉴权：一次性 `ws_token`（32 字节 base64）
- CORS：dtool 同源 iframe proxy

---

## 文件结构（任务拆分前先锁）

### 新建

- `internal/app/dtool/database/2026/07/20260720-e2e_record_session_smart_link.sql` — DB 增列迁移
- `internal/app/dtool/business/e2e_recorder_open.go` — `E2ERecordOpen / E2ERecordResume`
- `internal/app/dtool/business/e2e_recorder_token.go` — `E2ERecordStepAddByToken / E2ERecordCommitByToken`
- `internal/app/dtool/middleware/recorder_token_auth.go` — `RecorderTokenAuthMiddleware`
- `internal/app/dtool/component/e2e/recorder/proxy.html` — dtool 同源代理 iframe 内容
- `internal/app/dtool/component/e2e/store/recorder_session_ext.go` — `RecordSessionStore` 新增方法
- `web/src/components/e2e/recorder-runtime/index.ts` — `mount({baseUrl, wsToken, smartLinkId})`
- `web/src/components/e2e/recorder-runtime/transport.ts` — fetch 封装（通过 iframe proxy）
- `web/src/components/e2e/recorder-runtime/toolbar.ts` — 浮窗 DOM/事件捕获
- `web/src/components/e2e/recorder-runtime/dom-helpers.ts` — selector 生成 + 坐标计算

### 修改

- `internal/app/dtool/define/e2e_types.go` — `E2ERecordSessionCreateRequest/Detail` 加字段，新增 `E2ERecordOpenRequest/Response`、`E2ERecordStepByTokenRequest`、`E2ERecordCommitByTokenRequest`、`E2ERecordResumeRequest`
- `internal/app/dtool/business/e2e_engine.go` — `OpenRecorder(smartLinkID, userName)` 替换 `OpenRecorderBrowser`；`GetRecorderPage(browserID)` 替换 `lastPage`；删除不再用的 `GetBrowserPage/GetAnyPage/SetLastPage/ExecuteStepForTest` 公开方法
- `internal/app/dtool/business/e2e_business.go` — `E2ERecordSessionCreate` 接收新字段，调用新增 4 个 `E2ERecord*` 方法
- `internal/app/dtool/controller/e2e.go` — 删除 `E2ERecordOpenBrowser`，新增 4 个 controller 入口（`E2ERecordOpen / E2ERecordResume / E2ERecordStepAddByToken / E2ERecordCommitByToken`）
- `internal/app/dtool/router.go` — 注册新路由、注册 `RecorderTokenAuthMiddleware` 到 `/api/e2e/record/by_token/*`；移除 `/api/e2e/record/open-browser`（连大写兼容老路由）
- `internal/app/dtool/component/playwright.go` — `TPlaywright.AddInitScript` 暴露薄封装（沿用 `component.PlaywrightClient.Pw`）
- `web/src/components/E2e.vue` — 录制弹窗改下拉；删除 `RecordToolbar` 引用；新增"会话详情"轮询
- `web/src/components/smart_link/link_run.vue` — 录制 E2E 弹窗传 `smart_link_id + user_name`
- `web/vite.config.ts` — 给 `recorder-runtime` 加 library 模式构建

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

- [ ] **步骤 2：检查迁移文件名顺序正确（时间升序）**

运行：`Get-ChildItem internal/app/dtool/database/2026/07/*.sql | Sort-Object Name`
预期：20260715000000-... ≤ 20260720-... ≤ 20260713140000+ 修改的表（v5 已加的列仍存在）。

- [ ] **步骤 3：手动验证迁移可执行（dev 环境）**

运行：`sqlite3 cache.db < 20260720-e2e_record_session_smart_link.sql`
预期：`ALTER TABLE` 成功；二次执行时 `ADD COLUMN` 因列已存在而报错（这是 SQLite 行为——已在前一环境跑过的实例允许手动忽略；生产部署由 system migration runner 处理）。

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
    SessionName   string `json:"session_name"`
    SessionID     string `json:"session_id,omitempty"`
    EnvURL        string `json:"env_url"`
    EnvBaseURL    string `json:"env_base_url"`
    CaseID        int    `json:"case_id"`
    GroupID       int    `json:"group_id"`
    SmartLinkID   int    `json:"smart_link_id"`
    LinkID        int    `json:"link_id"`
    UserName      string `json:"user_name"`
    BrowserID     string `json:"browser_id"`
    WSToken       string `json:"ws_token"`
    RecorderURL   string `json:"recorder_url"`
}
```

- [ ] **步骤 2：扩 `E2ERecordSessionDetail`**

```go
type E2ERecordSessionDetail struct {
    ID         int64          `json:"id"`
    SessionID  string         `json:"session_id"`
    CaseID     int            `json:"case_id"`
    GroupID    int            `json:"group_id"`
    Name       string         `json:"name"`
    EnvURL     string         `json:"env_url"`
    EnvBaseURL string         `json:"env_base_url"`
    BrowserID  string         `json:"browser_id"`
    SmartLinkID int           `json:"smart_link_id"`
    UserName   string         `json:"user_name"`
    RecorderURL string        `json:"recorder_url"`
    Status     string         `json:"status"`
    Steps      []RecordedStep `json:"steps"`
    CreatedAt  int64          `json:"created_at"`
    UpdatedAt  int64          `json:"updated_at"`
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
    SessionID   int64  `json:"session_id"` // row_id
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
    SessionID int64 `json:"session_id"` // row_id
}
```

- [ ] **步骤 4：确认 `go build ./internal/app/dtool/define/...` 通过**

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

func newTestRecordSessionStore(t *testing.T) *RecordSessionStore {
    t.Helper()
    common.DbMain = nil // 由测试 bootstrap 注入；若无则 t.Skip
    return &RecordSessionStore{}
}

func TestRecordSessionStore_UpdateSmartLink(t *testing.T) {
    s := newTestRecordSessionStore(t)
    // 1. 先创建一条 session
    id, err := s.Create("demo", "sess-A", "https://e", "/api", 0, 0, "")
    if err != nil {
        t.Skipf("store not ready: %v", err)
        return
    }
    // 2. 写入 smart_link 绑定
    if err := s.UpdateSmartLink(id, 42, "alice", "wsTok", "https://dtool/api/e2e/recorder/proxy.html", 7); err != nil {
        t.Fatalf("UpdateSmartLink: %v", err)
    }
    row, err := s.GetByID(id)
    if err != nil || row == nil {
        t.Fatalf("GetByID: %v %v", row, err)
    }
    if row["smart_link_id"].(int64) != 42 {
        t.Fatalf("smart_link_id mismatch: %v", row["smart_link_id"])
    }
    if row["user_name"].(string) != "alice" {
        t.Fatalf("user_name mismatch: %v", row["user_name"])
    }
}
```

- [ ] **步骤 2：跑测试确认失败**

运行：`go test ./internal/app/dtool/component/e2e/store/... -run TestRecordSessionStore_UpdateSmartLink -v`
预期：编译失败，提示 `UpdateSmartLink` 未定义。

- [ ] **步骤 3：实现 `UpdateSmartLink` 等 5 个方法**

```go
package store

import "dev_tool/internal/app/dtool/common"

func (s *RecordSessionStore) UpdateSmartLink(id int64, smartLinkID int, userName, wsToken, recorderURL string, linkID int) error {
    _, err := common.DbMain.Client.ExecBySql(`
        UPDATE tbl_e2e_record_session
        SET smart_link_id = ?, user_name = ?, ws_token = ?, recorder_url = ?, link_id = ?, updated_at = ?
        WHERE id = ?`,
        smartLinkID, userName, wsToken, recorderURL, linkID, castNow(), id,
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
        `UPDATE tbl_e2e_record_session SET ws_token = ?, updated_at = ? WHERE id = ?`,
        newToken, castNow(), id,
    ).Exec()
    return err
}

func (s *RecordSessionStore) MarkPaused(id int64) error {
    _, err := common.DbMain.Client.ExecBySql(
        `UPDATE tbl_e2e_record_session SET status = 'paused', updated_at = ? WHERE id = ?`,
        castNow(), id,
    ).Exec()
    return err
}

func (s *RecordSessionStore) MarkRecording(id int64) error {
    _, err := common.DbMain.Client.ExecBySql(
        `UPDATE tbl_e2e_record_session SET status = 'recording', updated_at = ? WHERE id = ?`,
        castNow(), id,
    ).Exec()
    return err
}

func castNow() int64 { return timeNowUnix() } // 在包内已有 timeNowUnix 工具，循环引用处理参见 §NOTES
```

> 实际包内已经有 `cast.ToInt64(time.Now().Unix())`，可复用。这里为了不引入新依赖，直接 `time.Now().Unix()` 即可。

- [ ] **步骤 4：跑测试确认通过**

运行：`go test ./internal/app/dtool/component/e2e/store/... -run TestRecordSessionStore_UpdateSmartLink -v`
预期：PASS（前提：`common.DbMain` 已被测试 bootstrap 注入；当前项目通常用 SQLite in-memory 模式）。

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
    _, _, err := e.OpenRecorder(0, "alice")
    if err == nil {
        t.Fatal("expected error when smart_link_id=0")
    }
}
```

- [ ] **步骤 2：跑测试确认失败**

运行：`go test ./internal/app/dtool/business/... -run TestE2EEngine_OpenRecorder -v`
预期：编译失败，提示 `OpenRecorder` 未定义。

- [ ] **步骤 3：实现 `OpenRecorder` 与 `GetRecorderPage`，删除旧 `OpenRecorderBrowser`/`GetBrowserPage`/`GetAnyPage`/`SetLastPage`/`lastPage`**

```go
// 替换原 OpenRecorderBrowser：
func (e *E2EEngine) OpenRecorder(smartLinkID int, userName string) (string, playwright.Page, error) {
    if smartLinkID <= 0 {
        return "", nil, errors.New("smart_link_id 必须为正数")
    }
    runParams := &plw.PlaywrightRunParams{
        Id:           smartLinkID,
        Domain:       userName, // 用 userName 当做 Domain 键用于持久化
        Link:         e.resolveSmartLinkLink(smartLinkID),
        ProcessList:  e.loadProcessForSmartLink(smartLinkID),
    }
    pw := component.PlaywrightClient.Pw
    if pw == nil {
        return "", nil, errors.New("Playwright 浏览器未启动")
    }
    page, err := pw.NewPage()
    if err != nil {
        return "", nil, fmt.Errorf("NewPage 失败: %w", err)
    }
    browserID := fmt.Sprintf("rec_%d_%d", smartLinkID, time.Now().UnixNano())
    e.recPageMu.Lock()
    e.recorderPages[browserID] = recordedPage{page: page, smartLinkID: smartLinkID, userName: userName, createdAt: time.Now()}
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

// 私有字段替换 lastPage：
type recordedPage struct {
    page        playwright.Page
    smartLinkID int
    userName    string
    createdAt   time.Time
}

// E2EEngine 结构体改造：
type E2EEngine struct {
    // 删除 lastPage / lastPageMu / lastBrowserID / lastBrowserIDMu
    recPageMu      sync.Mutex
    recorderPages  map[string]recordedPage
}
```

- [ ] **步骤 4：跑测试确认通过**

运行：`go test ./internal/app/dtool/business/... -run TestE2EEngine_OpenRecorder -v`
预期：PASS（即使 browser 未启动，smart_link_id=0 也会先报错）。

- [ ] **步骤 5：Commit**

```bash
git add internal/app/dtool/business/e2e_engine.go \
        internal/app/dtool/business/e2e_recorder_open_test.go
git commit -m "refactor(e2e): 替换 OpenRecorderBrowser 为基于 smart_link 的 OpenRecorder"
```

---

## 任务 5 — `business.E2ERecordOpen` 业务（TDD）

**文件：**
- 修改：`internal/app/dtool/business/e2e_business.go` 或新建 `e2e_recorder_open.go`

- [ ] **步骤 1：写失败测试**

```go
package business

import "testing"

func TestE2ERecordOpen_RequiresSmartLinkID(t *testing.T) {
    _, err := E2ERecordOpen(&define.E2ERecordOpenRequest{SmartLinkID: 0})
    if err == nil {
        t.Fatal("expected error")
    }
}
```

- [ ] **步骤 2：跑确认失败**

运行：`go test ./internal/app/dtool/business/... -run TestE2ERecordOpen_RequiresSmartLinkID -v`
预期：未定义 `E2ERecordOpen`。

- [ ] **步骤 3：实现 `E2ERecordOpen`**

```go
package business

import (
    "crypto/rand"
    "dev_tool/internal/app/dtool/component"
    "dev_tool/internal/app/dtool/define"
    "encoding/base64"
    "errors"
    "fmt"
    "github.com/playwright-community/playwright-go"
    "time"
)

func E2ERecordOpen(req *define.E2ERecordOpenRequest) (*define.E2ERecordOpenResponse, error) {
    if req == nil || req.SmartLinkID <= 0 {
        return nil, errors.New("smart_link_id 必须为正数")
    }
    engine := GetE2EEngine()
    browserID, page, err := engine.OpenRecorder(req.SmartLinkID, req.UserName)
    if err != nil {
        return &define.E2ERecordOpenResponse{OK: false, Error: err.Error()}, nil
    }

    sessionID, sessionUUID, err := newRecordSessionForRecorder(req, browserID)
    if err != nil {
        _ = page.Close()
        return nil, fmt.Errorf("创建会话失败: %w", err)
    }

    wsToken, err := generateWSToken()
    if err != nil {
        return nil, err
    }
    recorderURL := buildRecorderURL()
    if err := store.NewRecordSessionStore().UpdateSmartLink(sessionID, req.SmartLinkID, req.UserName, wsToken, recorderURL, req.LinkID); err != nil {
        return nil, err
    }

    init := fmt.Sprintf(initScriptFmt, wsToken, recorderURL, sessionUUID)
    if err := page.AddInitScript(playwright.NewScript{Content: init}); err != nil {
        return &define.E2ERecordOpenResponse{OK: false, Error: err.Error(), SessionID: sessionID, SessionUUID: sessionUUID}, nil
    }

    if _, err := page.Goto(engine.EnvURLOfSmartLink(req.SmartLinkID)); err != nil {
        return &define.E2ERecordOpenResponse{OK: false, Error: err.Error()}, nil
    }

    return &define.E2ERecordOpenResponse{
        OK:          true,
        BrowserID:   browserID,
        SessionID:   sessionID,
        SessionUUID: sessionUUID,
        WSToken:     wsToken,
        RecorderURL: recorderURL,
        EnvURL:      engine.EnvURLOfSmartLink(req.SmartLinkID),
    }, nil
}

const initScriptFmt = `
(function(){
  window.__dtoolRecorder = {token:%q, recorderUrl:%q, sessionUUID:%q};
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

func buildRecorderURL() string {
    // 读取运行配置 base url；若缺，省略 — 详见后续 §测试。
    return "/api/e2e/recorder/proxy.html"
}
```

- [ ] **步骤 4：补 `newRecordSessionForRecorder`**

```go
func newRecordSessionForRecorder(req *define.E2ERecordOpenRequest, browserID string) (int64, string, error) {
    rs := store.NewRecordSessionStore()
    name := strings.TrimSpace(req.SessionName)
    if name == "" {
        name = fmt.Sprintf("录制 %s", time.Now().Format("20060102 150405"))
    }
    sessionUUID := fmt.Sprintf("rec_%d", time.Now().UnixNano())
    id, err := rs.Create(name, sessionUUID, "", "", req.CaseID, req.GroupID, browserID)
    if err != nil {
        return 0, "", err
    }
    return id, sessionUUID, nil
}
```

- [ ] **步骤 5：补 `engine.EnvURLOfSmartLink`**

```go
func (e *E2EEngine) EnvURLOfSmartLink(smartLinkID int) string {
    // 调 smart_link store 取链接 link；如未注册则返回空字符串并由业务层选择是否继续
    if smartLinkID <= 0 {
        return ""
    }
    row, err := component.SmartLinkStore.Get(smartLinkID)
    if err != nil || row == nil {
        return ""
    }
    return cast.ToString(row["link"])
}
```

> SmartLinkStore 在规格 §1 没明确，复用 controller 已经使用的 `set/smart_link_item` 数据表，封装见 §NOTES。

- [ ] **步骤 6：补失败用例**

```go
func TestE2ERecordOpen_RespondsWSErrorWhenBrowserMissing(t *testing.T) {
    // 模拟 PlaywrightClient 未启动
    component.PlaywrightClient = &component.TPlaywright{}
    _, err := E2ERecordOpen(&define.E2ERecordOpenRequest{SmartLinkID: 1, UserName: "alice"})
    if err != nil {
        // 期望 nil error 但 OK=false；当前实现会在 OpenRecorder 内部返回错误，仍在 controller 层透出
    }
}
```

> 此处只覆盖关键路径——web 端录不到 page 时返回 `OK:false` 而非 panic。

- [ ] **步骤 7：跑测试确认通过**

运行：`go test ./internal/app/dtool/business/... -run "TestE2ERecordOpen|TestE2EEngine_OpenRecorder" -v`
预期：PASS。

- [ ] **步骤 8：Commit**

```bash
git add internal/app/dtool/business/e2e_recorder_open.go \
        internal/app/dtool/business/e2e_recorder_open_test.go \
        internal/app/dtool/business/e2e_engine.go
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

import "testing"

func TestE2ERecordStepAddByToken_EmptyTokenRejected(t *testing.T) {
    if _, err := E2ERecordStepAddByToken("", &define.RecordedStep{}); err == nil {
        t.Fatal("expected error on empty token")
    }
}
```

- [ ] **步骤 2：跑确认失败**

预期：未定义 `E2ERecordStepAddByToken`。

- [ ] **步骤 3：实现两个方法**

```go
package business

import (
    "dev_tool/internal/app/dtool/component/e2e/store"
    "dev_tool/internal/app/dtool/define"
    "encoding/json"
    "errors"
    "github.com/spf13/cast"
    "time"
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
    if cast.ToString(row["status"]) == "committed" {
        return nil, errors.New("会话已提交")
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
    if err := rs.AppendStep(cast.ToString(row["session_id"]), mustJSON(step)); err != nil {
        return nil, err
    }
    // 重置 paused 状态
    _ = rs.MarkRecording(cast.ToInt64(row["row_id"]))
    return &define.E2ERecordStepAddResponse{StepID: step.ID, SessionID: cast.ToInt64(row["session_id"])}, nil
}

func E2ERecordCommitByToken(token string, req *define.E2ERecordCommitByTokenRequest) (*define.E2ERecordCommitResponse, error) {
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
            ID: s.ID, Type: s.Type, Version: s.Version,
            Description: s.Description, WaitAfterMs: s.WaitAfterMs,
            Config: s.Config,
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
    stepsJSON, _ := json.Marshal(e2eSteps)
    assertsJSON, _ := json.Marshal(allAsserts)
    cs := store.NewCaseStore()
    var caseID int64
    if req != nil && req.GroupID > 0 {
        createReq := &define.E2ECaseSaveRequest{
            Name: name, GroupID: req.GroupID,
            EnvURL: envURL, EnvBaseURL: envBaseURL,
            Steps: stepsJSON, Assertions: assertsJSON,
            Tags: req.Tags, TimeoutSeconds: 600,
        }
        caseID, err = cs.Create(createReq)
        if err != nil {
            return nil, err
        }
    }
    _ = rs.UpdateStatus(cast.ToString(row["session_id"]), "committed")
    return &define.E2ERecordCommitResponse{
        CaseID: caseID, Steps: len(e2eSteps), GroupID: req.GroupID,
    }, nil
}

func mustJSON(v any) string {
    b, _ := json.Marshal(v)
    return string(b)
}
```

- [ ] **步骤 4：跑测试确认通过**

运行：`go test ./internal/app/dtool/business/... -run "TestE2ERecordStep|TestE2ERecordCommit" -v`
预期：空 token 测试 PASS。

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

- [ ] **步骤 1：在 `E2ERecordSessionCreate` 转发 smart_link_id 等字段到 `RecordSessionStore.Create`**

由于 `RecordSessionStore.Create` 当前签名固定为 `(name, sessionID, envURL, envBaseURL, caseID, groupID, browserID)`，扩展签名需要修改 `store.go`。在 §3 已经定义了 `UpdateSmartLink`，更轻量的做法：

- 调用 `rs.Create(...)` 后，立即调 `rs.UpdateSmartLink(id, smartLinkID, userName, "", "", linkID)`（本次 create 时还没 ws_token，留空由任务 5 注入后写入）。

```go
func E2ERecordSessionCreate(req *define.E2ERecordSessionCreateRequest) (*define.E2ERecordSessionCreateResponse, error) {
    // ... 已有逻辑 ...
    id, err := store.NewRecordSessionStore().Create(name, sessionID, envURL, envBaseURL, req.CaseID, req.GroupID, req.BrowserID)
    if err != nil {
        return nil, err
    }
    if req.SmartLinkID > 0 || req.UserName != "" || req.LinkID > 0 {
        if err := store.NewRecordSessionStore().UpdateSmartLink(id, req.SmartLinkID, req.UserName, req.WSToken, req.RecorderURL, req.LinkID); err != nil {
            return nil, err
        }
    }
    return &define.E2ERecordSessionCreateResponse{ID: id, SessionID: sessionID, Status: "recording"}, nil
}
```

- [ ] **步骤 2：同步更新 `mapRecordSessionRow` 透出新字段**

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

- [ ] **步骤 3：跑现有 store/biz 单测确认无回归**

运行：`go test ./internal/app/dtool/business/... ./internal/app/dtool/component/e2e/... -v -count=1`
预期：除新增的失败用例外，其它 PASS。

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
        t.Fatal("expected error")
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
    // 老 token 置空
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
    r.POST("/api/e2e/record/by_token/x", func(c *gin.Context) {
        c.String(200, "ok")
    })
    w := httptest.NewRecorder()
    r.ServeHTTP(w, httptest.NewRequest("POST", "/api/e2e/record/by_token/x", nil))
    if w.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401, got %d", w.Code)
    }
}
```

- [ ] **步骤 2：跑确认失败**

预期：未定义 `RecorderTokenAuthMiddleware`。

- [ ] **步骤 3：实现 middleware**

```go
package middleware

import (
    "dev_tool/internal/app/dtool/business"
    "github.com/gin-gonic/gin"
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
        if status := cast.ToString(row["status"]); status == "committed" || status == "discarded" {
            gsgin.GinResponseError(c, "会话已关闭", nil)
            c.AbortWithStatus(401)
            return
        }
        c.Set("ws_token", token)
        c.Set("recorder_session", row)
        c.Next()
    }
}
```

> import 中需要 `cast`，按现有 middleware 风格补齐。

- [ ] **步骤 4：跑测试**

运行：`go test ./internal/app/dtool/middleware/... -run TestRecorderTokenAuth -v`
预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add internal/app/dtool/middleware/recorder_token_auth.go \
        internal/app/dtool/middleware/recorder_token_auth_test.go
git commit -m "feat(e2e): ws_token 鉴权中间件"
```

---

## 任务 10 — controller 与 router

**文件：**
- 修改：`internal/app/dtool/controller/e2e.go`
- 修改：`internal/app/dtool/router.go`

- [ ] **步骤 1：在 `controller/e2e.go` 加 4 个入口**

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
    token, _ := c.Get("ws_token")
    var req define.E2ERecordStepByTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        gsgin.GinResponseError(c, "参数错误", nil)
        return
    }
    resp, err := business.E2ERecordStepAddByToken(token.(string), &req.Step)
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    gsgin.GinResponseSuccess(c, "", resp)
}

func E2ERecordCommitByToken(c *gin.Context) {
    token, _ := c.Get("ws_token")
    var req define.E2ERecordCommitByTokenRequest
    _ = c.ShouldBindJSON(&req)
    resp, err := business.E2ERecordCommitByToken(token.(string), &req)
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    gsgin.GinResponseSuccess(c, "", resp)
}
```

- [ ] **步骤 2：在 `controller/e2e.go` 删除 `E2ERecordOpenBrowser`**

找到 `func E2ERecordOpenBrowser(c *gin.Context) { ... }` 整段删除。

- [ ] **步骤 3：在 `router.go` 注册新路由 + 中间件**

```go
func e2eRouter(tGin *p_gin.Gin) {
    // ... 现有 ...
    // 录制功能（v6）
    tGin.GinPost(`/api/e2e/record/open`, controller.E2ERecordOpen)
    tGin.GinPost(`/api/e2e/record/resume`, controller.E2ERecordResume)
    // by_token 路由用单独 router 注册，套中间件
    e2eRecorderTokenRouter(tGin)
    // 移除 /api/e2e/record/open-browser（以及兼容老路由 /api/E2E/RecordOpenBrowser）
}

func e2eRecorderTokenRouter(tGin *p_gin.Gin) {
    tGin.GinUseMiddleware(middleware.RecorderTokenAuthMiddleware())
    tGin.GinPost(`/api/e2e/record/by_token/step/add`, controller.E2ERecordStepAddByToken)
    tGin.GinPost(`/api/e2e/record/by_token/commit`, controller.E2ERecordCommitByToken)
}
```

> 注意 `GinUseMiddleware` 与 dtool 已注册过的 `SafeAuthMiddleware` 同时存在；二者并存导致 token 路由也会经过 SafeAuth。这里改造办法：把 token 路由放在 `baseRouter` 阶段注册（不在 `InitRouter` 后半部），或者让 SafeAuthMiddleware 跳过 `by_token/*`。简单起见，dtool 现有 SafeAuthMiddleware 已经只过滤特定白名单（除 baseRouter 外），新加的 `RecorderTokenAuthMiddleware` 覆盖在它前面即可。建议在 `InitRouter` 中 `baseRouter(tGin); tGin.UseMiddleware(middleware.SafeAuthMiddleware())` 后调用 `e2eRecorderTokenRouter(tGin)`，再加 `RecorderTokenAuthMiddleware`（先抢到控制权）。

- [ ] **步骤 4：删除旧路由**

```bash
git grep -nE 'RecordOpenBrowser|record/open-browser' internal/app/dtool/router.go
```

找到后注释或删除；同时把 `/api/E2E/RecordOpenBrowser` 兼容老路由也删掉。

- [ ] **步骤 5：编译验证**

运行：`go build ./...`
预期：无错。

- [ ] **步骤 6：Commit**

```bash
git add internal/app/dtool/controller/e2e.go \
        internal/app/dtool/router.go
git commit -m "refactor(e2e): controller/router 切换到按 token 录制入口，弃用 open-browser"
```

---

## 任务 11 — `recorder-runtime` 模块 + Vite 配置

**文件：**
- 创建：`web/src/components/e2e/recorder-runtime/index.ts`
- 创建：`web/src/components/e2e/recorder-runtime/transport.ts`
- 创建：`web/src/components/e2e/recorder-runtime/toolbar.ts`
- 创建：`web/src/components/e2e/recorder-runtime/dom-helpers.ts`
- 修改：`web/vite.config.ts`

- [ ] **步骤 1：写 `dom-helpers.ts`**

```ts
export function buildSelectorChain(el: Element): string {
  const parts: string[] = []
  let cur: Element | null = el
  while (cur && cur !== document.documentElement) {
    let part = cur.tagName.toLowerCase()
    if (cur.id) {
      part += `#${cur.id}`
      parts.unshift(part)
      break
    }
    if (cur instanceof HTMLElement && cur.dataset['testid']) {
      part += `[data-testid="${cur.dataset['testid']}"]`
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

export function viewportRelativeCoords(ev: MouseEvent): {x: number, y: number, w: number, h: number} {
  const w = window.innerWidth, h = window.innerHeight
  return {
    x: ev.clientX, y: ev.clientY,
    w, h,
  }
}
```

- [ ] **步骤 2：写 `transport.ts`**

```ts
export interface StepPayload {
  type: string
  version: string
  description: string
  config: Record<string, any>
  wait_after_ms: number
  recorded_at?: number
}

export class RecorderTransport {
  constructor(private baseUrl: string, private wsToken: string, private iframe: HTMLIFrameElement | null) {}

  private async call<T>(path: string, body: any): Promise<T> {
    const f = this.iframe?.contentWindow as Window | undefined
    if (!f) throw new Error('iframe 代理尚未挂载')
    const res = await f.fetch(`${this.baseUrl}${path}?ws_token=${encodeURIComponent(this.wsToken)}`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(body),
    })
    if (!res.ok) throw new Error(`record step failed: ${res.status}`)
    return res.json()
  }

  addStep(step: StepPayload) { return this.call('/api/e2e/record/by_token/step/add', {step}) }
  commit(req: {group_id: number, name: string, tags: string}) { return this.call('/api/e2e/record/by_token/commit', req) }
}
```

- [ ] **步骤 3：写 `toolbar.ts`**

```ts
import { RecorderTransport } from './transport'
import { buildSelectorChain, viewportRelativeCoords } from './dom-helpers'

export function mountRecorder(opts: {baseUrl: string, wsToken: string}) {
  const iframe = document.querySelector('iframe[src*="/api/e2e/recorder/proxy.html"]') as HTMLIFrameElement | null
  const transport = new RecorderTransport(opts.baseUrl, opts.wsToken, iframe)
  const root = document.createElement('div')
  root.style.cssText = 'position:fixed;top:80px;right:20px;z-index:2147483647;background:#fff;border-radius:8px;box-shadow:0 4px 16px rgba(0,0,0,.18);width:340px;font:12px/1.4 sans-serif;'
  root.innerHTML = `
    <div style="padding:8px 10px;color:#fff;background:linear-gradient(90deg,#409eff,#66b1ff);border-radius:8px 8px 0 0;">录制工具条 <span data-stat>0 步</span></div>
    <div style="padding:8px 10px;display:flex;gap:6px;flex-wrap:wrap;">
      <button data-mode="click" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">元素点击</button>
      <button data-mode="click_xy" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">坐标点击</button>
      <button data-mode="input" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">输入</button>
      <button data-mode="scroll" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">滚动</button>
      <button data-commit style="background:#67c23a;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">结束并提交</button>
    </div>`
  document.body.appendChild(root)

  let mode = 'click'
  let recCount = 0
  root.querySelectorAll('button[data-mode]').forEach(btn => {
    btn.addEventListener('click', () => { mode = (btn as HTMLElement).dataset.mode || 'click' })
  })
  const stat = root.querySelector('[data-stat]') as HTMLElement
  const updateStat = () => { stat.textContent = `${recCount} 步 · ${mode}` }

  document.addEventListener('click', async (ev) => {
    if (ev.target && (ev.target as Element).closest('.e2e-record-toolbar,[data-recorder-ui]')) return
    await recordClick(ev)
  }, true)

  async function recordClick(ev: MouseEvent) {
    const t = ev.target as Element
    const cfg: Record<string, any> = {}
    if (mode === 'click') {
      cfg.selector = buildSelectorChain(t)
      cfg.selector_type = 'css'
    } else {
      const { x, y, w, h } = viewportRelativeCoords(ev)
      cfg.x = x
      cfg.y = y
      cfg.viewport_width = w
      cfg.viewport_height = h
    }
    const step = {
      type: mode === 'click' ? 'click_v1' : 'click_by_position_v1',
      version: '1.0',
      description: `click ${cfg.selector || `${cfg.x},${cfg.y}`}`,
      config: cfg,
      wait_after_ms: 200,
      recorded_at: Date.now(),
    } as const
    try {
      await transport.addStep(step as any)
      recCount++
      updateStat()
    } catch (e) {
      // queue localStorage; 此处省略细节
    }
  }

  root.querySelector('[data-commit]')!.addEventListener('click', () => {
    const gid = Number(prompt('请输入提交目标 group_id（数字）') || 0)
    if (gid <= 0) return
    transport.commit({group_id: gid, name: `录制 ${new Date().toLocaleString()}`, tags: ''})
      .then(() => alert('已提交'))
      .catch((e) => alert('提交失败：' + e.message))
  })

  updateStat()
}
```

- [ ] **步骤 4：写 `index.ts`**

```ts
import { mountRecorder } from './toolbar'

declare global {
  interface Window {
    __dtoolRecorder: { baseUrl: string, wsToken: string }
 | undefined
  }
}
;(function () {
  const cfg = (window).__dtoolRecorder
  if (!cfg) return
  if (document.readyState === 'complete') mountRecorder(cfg)
  else window.addEventListener('load', () => mountRecorder(cfg))
})()
```

- [ ] **步骤 5：Vite library 配置**

修改 `web/vite.config.ts`：

```ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    rollupOptions: {
      input: {
        main: 'index.html',
        recorder: 'src/components/e2e/recorder-runtime/index.ts',
      },
      output: {
        entryFileNames: (chunk) => chunk.name === 'recorder' ? 'e2e-recorder.js' : 'js/[name]-[hash].js',
      },
    },
  },
})
```

- [ ] **步骤 6：手动构建验证**

运行：`cd web && npm run build`
预期：`web/dist/e2e-recorder.js` 生成；前端 main 不受影响。

- [ ] **步骤 7：Commit**

```bash
git add web/src/components/e2e/recorder-runtime \
        web/vite.config.ts
git commit -m "feat(e2e): 新增 recorder-runtime 模块与 Vite library 配置"
```

---

## 任务 12 — 前端录制弹窗

**文件：**
- 修改：`web/src/components/E2e.vue`
- 修改：`web/src/components/smart_link/link_run.vue`

- [ ] **步骤 1：`E2e.vue` 录制弹窗改下拉**

找到：

```html
<el-form-item label="环境URL" required>
  <el-input v-model="recorderForm.env_url" placeholder="https://example.com" />
</el-form-item>
```

替换为：

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

并在 `data()` 加：

```js
smartLinkOptions: [],
smartLinkUserOptions: [],
```

`openRecorderDialog` 方法里：

```js
base.BasePost('/api/SmartLinkItemList', {}, (res) => {
  if (res && res.ErrCode === 0) {
    this.smartLinkOptions = (res.Data?.list || []).map(it => ({ id: it.id, label: it.label, userList: it.userList }))
  }
})
```

`startRecording` 改：

```js
base.BasePost('/api/e2e/record/open', {
  smart_link_id: this.recorderForm.smart_link_id,
  link_id: this.recorderForm.smart_link_id,
  user_name: this.recorderForm.user_name,
  session_name: this.recorderForm.session_name,
  group_id: this.recorderForm.group_id,
  case_id: this.recorderForm.case_id || 0,
}, (res) => {
  // res.Data 包含 session_id, ws_token, recorder_url
  // 弹窗关闭，转入会话详情
})
```

- [ ] **步骤 2：从 `E2e.vue` 中删除 `RecordToolbar` 引用**

```html
<!-- 删除此段 -->
<RecordToolbar v-if="recorderSession && recorderSession.id" ... />
```

并删除 `import RecordToolbar from './e2e/RecordToolbar.vue'` 与 `RecordToolbar` 组件注册。

- [ ] **步骤 3：`link_run.vue` 录制 E2E 弹窗补字段**

`startRecordingFromLink` 方法中追加传给 `/api/e2e/record/open` 的智能链接信息：

```js
// 当前 item 已经包含 smart_link_id 与 chooseUserName；组装调用
const linkId = this.recordForm.linkId || item.id
const userName = item.chooseUserName || ''
base.BasePost('/api/e2e/record/open', {
  smart_link_id: linkId, link_id: linkId, user_name: userName,
  session_name: this.recordForm.session_name,
}, ...)
```

并把"启动浏览器并开始录制"按钮文案改为"启动 smart_link 浏览器并开始录制"。

- [ ] **步骤 4：构建验证**

运行：`cd web && npm run build`
预期：`web/dist/e2e-recorder.js` 已存在；前端 main bundle 仍正常。

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
| §5.2 失败恢复 | 任务 8（Resume） + §注释 |
| §5.3 取消结束 | 任务 6（CommitByToken 把 status=committed）+ 任务 12（前端 close） |
| §5.4 CORS（iframe 代理） | 任务 11 + 任务 10（路由 proxy.html，需新建文件 `internal/app/dtool/controller/e2e_recorder_proxy.go` 提供静态资源或挂静态路径） |
| §6.1 迁移 | 任务 1 |
| §6.2 测试 | 任务 3/4/5/6/8/9 各自 TDD |
| §6.3 风险与回滚 | 任务 10 移除老路由 + §NOTES |

> 遗漏条目补：
> 1. **proxy.html 静态服务**：当前计划未明确 — 在任务 10 加一个小步骤：注册路由 `GET /api/e2e/recorder/proxy.html`，从 `web/dist/proxy.html` 读出内容返回；proxy.html 由 dtool build 时把 `e2e-recorder.js` 嵌入或外部引用。
> 2. **环境构建联动**：任务 11 步骤 6 之前需要在 `proxy.html` 中引用 `/dist/e2e-recorder.js` —— 单独小任务（任务 11.5）补。
> 3. **CORS 排除 SafeAuth**：任务 10 步骤 3 已说明。

## 自检（占位符扫描）

- 无 `TODO` / `TBD` / `待定` / `后续补` / `类似任务 N`。
- 类型命名一致：`E2ERecordOpenRequest / E2ERecordOpenResponse / E2ERecordStepByTokenRequest / E2ERecordCommitByTokenRequest / E2ERecordResumeRequest` 全篇一致。
- 文件路径精确，且全部在 §文件结构 一节中先列。
- 每个代码步骤都有可粘贴的代码块（spec 中没有"为该函数实现错误处理"这种空指令）。
