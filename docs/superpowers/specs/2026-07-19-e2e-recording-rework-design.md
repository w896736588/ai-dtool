# E2E 录制与 smart_link 打通 — 设计规格

- 日期：2026-07-19
- 作者：Cursor（brainstorming 流程产出）
- 状态：待人工审查

## 1. 场景与目标

### 1.1 现状痛点

1. 录制入口只能"自己启一个 Playwright 裸 page"，独立维护登录态，绕开了 smart_link
   里"链接 + 账号 + 自动化登录"能力；
2. 工具条 `RecordToolbar.vue` 是 dtool 主窗 viewport 的浮窗，但它要监听的是另一个
   Playwright 浏览器页面，跨窗口根本无法精准捕获——所以现有录制"理论上能跑、
   实际上录不上"。

### 1.2 改造目标

- 录制入口与工具条共享同一个被录制 page；
- 被录制 page 通过 smart_link 的能力栈（链接配置 + Process 登录 +
  ContextPageList 持久化）打开；
- dtool 主窗变成"录制会话的观察/控制台"，不再渲染工具条。

### 1.3 非目标

- 不动 smart_link 现有 Process 执行逻辑；
- 不重写 e2e 引擎、对用例运行无影响；
- 不引入新 SSE/WebSocket，事件传输走普通 HTTP；
- 老 `/api/e2e/record/open-browser` 旧接口直接弃用，前端硬迁移，不保留兼容层。

## 2. 现状事实（基于实际代码）

| 位置 | 行为 | 改动必要性 |
|---|---|---|
| `business/e2e_engine.go` `OpenRecorderBrowser` | 直接拿 `component.PlaywrightClient.BrowserWebkitChrome.NewPage()` 不走 ContextPageList，登录态完全丢失 | 必须改走 smart_link 链路 |
| `business/e2e_engine.go` `lastPage` 字段 | 维护"最近一次 page"，录制回放靠它捞 | 需要改成"按 smart_link_id + user 索引 page" |
| `component/e2e/step_executor/...` `ExecuteStepForTest / ApplyPostStepWaitForTest` | 暴露 `executeStep` 给单步回放 | 保留，但内部 page 由 smart_link 提供 |
| `define/e2e_types.go` 中已定义 `RecordedStep / E2ERecordOpenBrowserRequest` 等录制结构体 | 已存在但未带 smart_link 绑定字段 | 加 `smart_link_id` / `user_name` 字段 |
| `database/2026/07/20260715000000-e2e_record_session_v5.sql`（untracked） | 已经为 record_session 加了 `row_id/session_id/group_id/browser_id/status` | 还需再加 5 列，本规格就是其内容 |
| `controller/e2e.go` | 已有 `E2ERecordSessionCreate/.../OpenBrowser` | 新增 `E2ERecordOpen`、替换 `OpenBrowser` |
| `web/components/E2e.vue` 录制弹窗 | 让用户手填 `env_url` | 改成"下拉选 smart_link 链接 + 选账号" |
| `web/components/smart_link/link_run.vue` `showRecordDialog` | 内部仍调 `/api/e2e/record/open-browser`，没真正接管登录态 | 改为调用新的 `/api/e2e/record/open`，传 `smart_link_id + user_name`，同时复用 `RedirectLink/RunItem` 中已选的账号 |
| `web/components/e2e/RecordToolbar.vue` | 渲染在 dtool 主 viewport | 重写为"独立 JS 模块（dist 打包成单一 bundle）"，由后端注入到被录 page |
| `web/components/e2e/StepConfirmDialog.vue` | 由 dtool 主窗弹出，编辑刚才一步 | 改写为"在注入页内自渲染"，保留弹窗 UI/UX |
| `plw.Playwright.GetPage` | 登录态 + Process 完整流程 | 不动，作为被录 page 来源 |

## 3. 端到端流程

```
[dtool 主窗 E2e.vue / link_run.vue]
  选择 smart_link_id + user_name
        │
        ▼
POST /api/e2e/record/open {smart_link_id, user_name, session_name, group_id?, case_id?}
        │
        │ 业务层
        ▼
1) 创建 e2e_record_session（status=recording）
2) 拼装 PlaywrightRunParams，复用 plw.Playwright.GetPage(...) ─ ─ ── 复用 Process 登录 + ContextPageList
3) Page 对象调用 PlaywrightBrowser.AddInitScript(...)，把 dist 后的 recorder.js 注入
   ▸ recorder.js 内 fetch('BASE_URL/api/e2e/record/by_session?ws_token=...')
   ▸ AddInitScript 中传入 smart_link_id + ws_token + dtool baseUrl + recorder UI 元素
4) 返回 {ok, browser_id, session_id, ws_token, recorder_url}
   ▸ recorder_url：dtool 同源静态资源 `/api/e2e/recorder/proxy.html`（见 §5.4），
     recorder.js 不直接走它，所有 fetch 经该 iframe 转发到 dtool 后端。

[被录 page 内 recorder.js（in-page 浮窗)]
  用户点击 / 输入 / 滚动 → fetch POST /api/e2e/record/by_token/step/add
                                                  │
                                                  ▼
                                       AppendStep（数据库） │
                                                  │
用户"结束录制"（recorder UI 内按钮 → POST /api/e2e/record/by_token/commit）
                                                  │
                                                  ▼
                              E2ERecordCommit（落库成 e2e_case，套现有逻辑）
                                                  │
[dtool E2e.vue "会话详情"弹窗]
  拉取 session.get → 步骤列表 + 注入页仍在持续往里 append → 实时同步
  用户可点击"回放整段"→ 后端用 e2e_engine 通过 lastPage 找到 current page 执行 steps
```

要点：

- **唯一权威态**：会话期间 page 只有 1 个，在 smart_link 的 `ContextPageList` 内，
  被 `smart_link_id + user_name` 索引；e2e_engine 新加方法 `GetRecorderPage(smartLinkID, userName)`。
- **工具条不再是 dtool 浮窗**：它是被注入 page 上的 DOM 浮窗。
- **事件传输**：浏览器内 JS 直接 fetch `e2e/record/by_token/step/add`，鉴权靠 `ws_token`
  （一次性、不可被外部 JS 读，仅在 AddInitScript 时通过闭包变量持有）。

## 4. 组件切分（小单元边界）

### 4.1 Go 后端

- `business.GetE2EEngine().OpenRecorder(smartLinkID, userName) (browserID, page, err)` —
  不再从裸 BrowserWebkitChrome 拿 page；走 `plw.Playwright.GetPage` 同一路径。
- `business.GetE2EEngine().GetRecorderPage(browserID) (playwright.Page, error)` —
  取代 `lastPage`，按 `browser_id` 检索。
- `business.E2ERecordSessionCreate` 新入参：`SmartLinkID int`、`UserName string`、
  `WSToken string`。
- `business.E2ERecordOpen` 替换 `E2ERecordOpenBrowser`；接受 smart_link_id + user_name，
  返回 ws_token + browser_id。
- `business.E2ERecordStepAddByToken(wsToken, step)` 新接口（避免 session_id 暴露），
  查 token→session 映射。
- 删除：`OpenRecorderBrowser`、`GetBrowserPage`、`GetAnyPage`、`SetLastPage`、
  `ExecuteStepForTest` / `ApplyPostStepWaitForTest`（保留用作单步回放内部方法，不导出）。

### 4.2 前端

- `web/src/components/e2e/recorder-runtime/` ← 新目录，产出独立 bundle
  `dist/e2e-recorder.js`（Vite library 模式）；
  - `recorder-runtime/index.ts` — 暴露 `mount({baseUrl, wsToken})`；
  - `recorder-runtime/toolbar.ts` — 移植现有 `RecordToolbar.vue` 的视图与逻辑到原生 DOM
    （**不依赖 vue 运行时**，避免重复打包 vue）；
  - `recorder-runtime/transport.ts` — fetch 封装；
  - `recorder-runtime/dom-helpers.ts` — selector 抓取 + 坐标 → viewport% 计算。
- `E2e.vue` 录制弹窗改：
  - 把 `env_url` 输入框替换为 `<el-select v-model="recorderForm.smart_link_id">`
    （来自 `smart_link` 接口），后跟账号 `<el-select>`。
  - "开始录制"按钮调 `/api/e2e/record/open`，拿到 `ws_token` 后弹窗关闭，回到主窗
    继续显示"会话详情"，并通过 polling 同步步骤。
- `link_run.vue` "录制 E2E" 弹窗改：把现成"链接 + 选中的账号"作为参数传入
  `/api/e2e/record/open`，不再让用户重复填。
- `E2e.vue` 主体的 `RecordToolbar.vue` 浮窗：删除，不再渲染在主窗。
- `web/vite.config.ts` —— 给 `recorder-runtime` 加一个 `library` 模式构建，目标 `umd`。

### 4.3 资源 / 打包

- `recorder-runtime` 的产出（`web/dist/e2e-recorder.js`）放在 `/web/dist/`（静态）；
  dtool 后端 `staticServe` 必须将其对外可访问（已存的 `/share/...` 类似机制）。
- AddInitScript 内注入的字符串：一个 `Function` 包的闭包 + 资源 URL，page 加载完后
  通过 `<script src=...>` 拉对应资源。避免每次都把 200kb JS 字符串硬塞 init script。

### 4.4 数据库迁移（增列，不破坏现有行）

`tbl_e2e_record_session` 新增列：

- `smart_link_id INTEGER NOT NULL DEFAULT 0` — 关联 `smart_link.id`
- `user_name TEXT NOT NULL DEFAULT ''`
- `ws_token TEXT NOT NULL DEFAULT ''`（唯一索引）
- `link_id INTEGER NOT NULL DEFAULT 0`（冗余，方便按链接过滤）
- `recorder_url TEXT NOT NULL DEFAULT ''`（注入脚本落地 URL，dtool 同源 proxy.html；详见 §5.4）

## 5. 鉴权 / 失败恢复 / 取消

### 5.1 鉴权

被录 page 的 JS 跑在第三方业务站点域名下（例如 `biz.example.com`），同源 / cookie 都不
复用 dtool。所以：

- **不走 cookie session**，session 中间件会拒。
- 用一次性 `ws_token`（32 字节 base64），生成后只在 AddInitScript 的闭包里持有；fetch
  时带在 query string 上：`POST /api/e2e/record/by_token/step/add?ws_token=xxx`。
- ws_token 仅本次会话期间有效，进入 `committed` / `discarded` 后立即失效
  （中间件拒绝）。
- 后端新增中间件 `RecorderTokenAuthMiddleware`：只放行 `/api/e2e/record/by_token/*`
  一组 URL；校验 token 查表，命中即允许。**不放行**其它 `/api/e2e/*`
  （那些仍走原 cookie/safe_auth 中间件）。
- `/api/e2e/record/commit` 仍走原鉴权（dtool 主窗调用，不需要 token）。

### 5.2 失败恢复

| 场景 | 行为 |
|---|---|
| dtool 进程被强杀但浏览器还在 | 浏览器 page 仍在跑、recorder.js 仍在 append；用户重开 dtool 在 E2e 页面调用 `/api/e2e/record/session/list?status=recording` 看到未结束会话，**支持续录**：前端弹"接管会话"按钮，调 `/api/e2e/record/resume?id=...`，后端复用原 smart_link_id + user_name + browser_id，重新 AddInitScript 时更新 token（旧的失效） |
| 浏览器被关闭 | recorder.js 报错；后端 `step/add` 在一定时间内（默认 30s）无新请求时把 session 标记为 `paused`，等用户点击"接管"或"丢弃" |
| 网络抖动 fetch 失败 | recorder.js 端用 fetch retry（指数退避，3 次）+ `localStorage` 暂存失败 step（key=`recorder_<token>_queue`），恢复后 flush |
| smart_link 链接被删 | `/api/e2e/record/open` 返回 4xx，前端弹"链接已删除，是否继续？"；用户选继续则回退为"无登录态裸开"（仅本次允许，并标注 `recorder_url=null` 不注入）—— **这次回退要明确写在边界** |
| AddInitScript 失败（page 已被 navigate 走） | 后端捕获 page.AddInitScript 错误，标记 `recorder_url=''`，session 状态保持 `recording` 但前端"接管"按钮会出现 |
| 跨域 CORS | 见 §5.4 |

### 5.3 取消 / 结束

- recorder.js 内"结束录制"按钮 → POST `/api/e2e/record/by_token/commit?ws_token=...&group_id=...`，
  立即停接收事件；
- 主窗 dtool "关闭会话"按钮 → POST `/api/e2e/record/session/delete?ws_token=...`，
  后端先把 session 状态置 `discarded`、token 失效，再异步关 page（避免阻塞 UI）。
- 浏览器被用户直接 X 掉：recorder.js 的 `window.unload` 发 `navigator.sendBeacon`
  给后端"end-of-recording"端点。

### 5.4 CORS 解决方案（重点）

被录 page 跑在 `biz.example.com`，dtool 在 `localhost:port`。浏览器 fetch 跨域默认拦截。

**采用"同源 iframe 代理"：**

1. AddInitScript 注入时附加一段代码，先把
   `window.__dtoolRecorder = { token, baseUrl, sessionId, iframe: null }`，
   然后在 DOM ready 后创建一个隐藏 iframe：`src="<baseUrl>/api/e2e/recorder/proxy.html"`。
2. `/api/e2e/recorder/proxy.html` 是 dtool 同源 HTML（新增极小静态文件，
   作为 recorder.js 的"宿主"），它里面加载 recorder.js 主逻辑；recorder.js 所有 fetch
   都改成 `iframe.contentWindow.fetch(...)`。
3. 这样所有 fetch 实际从 dtool 同源发，绕过 CORS。代价是页面 DOM 多一个隐藏 iframe
   （`<iframe style="position:fixed; width:1px; height:1px; opacity:0; pointer-events:none; border:0;">`）。

## 6. 迁移 / 测试 / 风险与回滚

### 6.1 数据库迁移

新增文件 `internal/app/dtool/database/2026/07/20260720-e2e_record_session_smart_link.sql`：

```sql
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

- 旧的 v5 迁移文件（`20260715000000-e2e_record_session_v5.sql`，untracked）保持不变——
  它已加的列仍然需要（已被本仓库当前 untracked 状态引用），后续在 PR 时一起追踪。
- 既有数据：列默认 0/空，老会话不再可被 `record/by_token/*` 访问（无 token），
  老 record_session 不再参与新增业务流。

### 6.2 测试计划

| 层 | 覆盖 |
|---|---|
| Go 单测 | `TestRecorderOpenRequiresSmartLinkID`、`TestRecorderTokenMiddleware`、`TestRecorderCommitFlow`、`TestRecorderResumeReplacesToken` |
| Go 集成 | 用 docker 内 Playwright 起一个无登录 smart_link（mock Process），断言 AddInitScript 后 page 内 `window.__dtoolRecorder` 存在、fetch 跨域 iframe proxy 成功 |
| 前端单测 | `recorder-runtime` 内 `transport.ts` mock fetch；`toolbar.ts` 渲染 DOM 后断言 click → step JSON shape |
| 前端 E2E | （用户决定是否做）用真实 smart_link + 一条含登录 Process 的链接，从头到尾录 5 步，commit 后比对 step 配置 |
| 手工验证 | smart_link 既有 2 个真实链接（登录型 + 免登录型）分别录一遍 |

### 6.3 风险 / 回滚

| 风险 | 缓解 | 回滚 |
|---|---|---|
| iframe 代理被目标站点 X-Frame-Options 拦 | AddInitScript 检测到 iframe 加载失败时 fallback 为 `navigator.sendBeacon`（仅支持 step.add，无法 commit） | 关闭新增入口，恢复 `E2ERecordOpenBrowser` 路由（保留一个开关通过 `runtime.ini` 控制） |
| recorder.js 体积大 | library 模式 + tree-shake，目标 ≤ 80kb gz | 关闭新增入口 |
| AddInitScript 闭包泄漏 ws_token 到全局 | init script 把 token 写入 IIFE 局部变量，只通过单一 fetch helper 暴露；release 时再 review | 同上 |
| smart_link Process 登录失败 | `OpenRecorder` 返回错误码 `OPEN_FAILED_LOGIN`，前端弹"登录流程未完成，是否手动登录后继续？"；用户在 page 内手动登录后点"已登录，继续录制"→ 后端 resume | 同上 |
| 用户点击录制入口但 dtool 同源端口变了（重新部署） | recorder_url 始终由后端在 open 时生成并随 token 一起返回，不写死 | 无 |

### 6.4 不在本期范围

- 录制多 page 切换（同一会话内多个 page）；
- 录制过程直接转"边录边跑"（mix record + replay）；
- 把 recorder.js 打包成 Chrome 扩展。
