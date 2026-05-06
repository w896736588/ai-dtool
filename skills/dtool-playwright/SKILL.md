---
name: dtool-playwright
description: Use when operating the dtool 自定义网页 / Playwright 模块 and the task involves running a smart-link login flow, receiving a Chromium userDataDir, and then using native Playwright launchPersistentContext(userDataDir) to take over the logged-in profile locally.
---

# dtool Playwright 技能

- 提供 dtool 自定义网页登录准备接口的使用说明。
- 支持两种模式：
  - **MCP 模式（推荐）**：服务端保持浏览器存活，AI 通过 MCP 工具直接操作，无需每步开关浏览器和截图。
  - **Playwright 模式（旧）**：服务端返回 userDataDir，AI 用原生 Playwright 接管。

## 强制约束

1. 调用接口前，必须向用户确认以下信息：
   - **请求地址**（`base_url`）：dtool 服务完整地址，例如 `http://127.0.0.1:17170`
   - **Token**：认证令牌，放在请求头 `Token` 中
   - **smart_link_id**：自定义网页配置 ID
   - **label**：自定义网页里目标链接的 label
   - **账号信息**：如果该链接依赖账号，确认使用哪个账号名
2. 所有请求统一使用 `POST`，`Content-Type: application/json; charset=utf-8`。
3. 统一使用 Python 脚本发送请求，避免 bash/PowerShell 编码问题。
4. 如果目标 smart-link 配置为"不保存用户数据"，该接口不能返回可复用目录，AI 不能继续原生接管。

## 模式一：MCP Server（推荐）

设置 `enable_mcp: true`，服务端登录后保持浏览器存活，创建 MCP SSE Server。AI 通过 MCP 工具直接操作浏览器。

### 请求体（新增 `enable_mcp` 字段）

```json
{
  "smart_link_id": 12,
  "label": "登录后首页",
  "account": "tester",
  "open_type": 0,
  "reuse_if_open": true,
  "enable_mcp": true
}
```

### MCP 模式响应示例

```json
{
  "browser_type": "chromium",
  "source_browser_closed": false,
  "user_data_dir": "C:/path/to/profile/4",
  "user_data_index": 4,
  "smart_link": { "id": 12, "label": "登录后首页" },
  "site": { "domain": "example.com", "url": "https://example.com/dashboard" },
  "current_page": { "url": "https://example.com/dashboard", "title": "控制台" },
  "mcp": {
    "enabled": true,
    "session_id": "mcp-br-12345",
    "sse_endpoint": "http://127.0.0.1:17170/mcp/ai-browser/mcp-br-12345/sse",
    "msg_endpoint": "http://127.0.0.1:17170/mcp/ai-browser/mcp-br-12345/message"
  },
  "usage_hint": "MCP模式：浏览器保持存活，AI通过MCP SSE端点直接调用browser_snapshot/browser_click等工具操作浏览器，无需重新打开浏览器"
}
```

### MCP 工作流

1. 确认 `base_url`、`Token`、`smart_link_id`、`label`、`account`
2. 调用 `/api/ai/browser/session/open`，设置 `enable_mcp: true`
3. 从响应中获取 `mcp.sse_endpoint` 和 `mcp.msg_endpoint`
4. 将 MCP SSE 端点配置到 AI 客户端（如 Claude Code、Cursor）
5. AI 通过 MCP 工具操作浏览器：

| 工具 | 说明 |
|---|---|
| `browser_snapshot` | 获取页面 Accessibility Tree（毫秒级，无需截图） |
| `browser_click` | 点击元素（用 ref 或 role+name 定位） |
| `browser_type` | 在输入框中输入文本 |
| `browser_fill` | 清空输入框并填入新文本 |
| `browser_navigate` | 导航到 URL |
| `browser_select_option` | 选择下拉选项 |
| `browser_screenshot` | 截图（仅最终验证用） |
| `browser_close` | 关闭浏览器会话 |

6. 典型操作流程：`browser_snapshot` → 分析结构 → `browser_click(ref="e3")` → `browser_snapshot` → 继续
7. 完成后调用 `browser_close` 关闭会话（或 30 分钟无操作自动关闭）

### MCP 工具调用示例

```
AI 调用 browser_snapshot
→ 返回:
- page
  - heading "用户登录" [level=1]
  - textbox "用户名" [ref=e1, value=""]
  - textbox "密码" [ref=e2, value=""]
  - button "登录" [ref=e3]
  - link "忘记密码" [ref=e4]

AI 分析后调用 browser_fill: {ref: "e1", text: "admin"}
AI 调用 browser_fill: {ref: "e2", text: "password123"}
AI 调用 browser_click: {ref: "e3"}
```

## 模式二：原生 Playwright（旧模式）

不设置 `enable_mcp`，服务端登录后关闭浏览器，返回 `userDataDir`，AI 用原生 Playwright 接管。

### 会话持久化约束

1. **禁止每个动作都重新打开浏览器**。AI 必须在单个 `with sync_playwright() as p:` 上下文中完成所有操作。
2. **禁止用截图来分析页面结构**。使用 `page.accessibility.snapshot()` 或 `page.evaluate()` 分析页面。

### 原生 Playwright 接管示例

```python
from playwright.sync_api import sync_playwright

user_data_dir = r"C:\path\to\profile\4"

with sync_playwright() as p:
    context = p.chromium.launch_persistent_context(
        user_data_dir=user_data_dir,
        headless=False,
        no_viewport=True,
    )
    page = context.pages[0] if context.pages else context.new_page()
    page.goto("https://example.com/dashboard")

    # 用无障碍树分析页面结构（毫秒级）
    snapshot = page.accessibility.snapshot()
    print(snapshot)

    # 用 locator 直接操作元素
    page.get_by_role("button", name="提交").click()
    page.get_by_label("用户名").fill("test")
```

## 接口详情

### `POST /api/ai/browser/session/open`

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `smart_link_id` | int | 是 | 自定义网页配置 ID |
| `label` | string | 是 | 自定义网页中的链接 label |
| `account` | string | 否 | 账号名 |
| `open_type` | int | 否 | 打开方式，`0` 沿用配置值 |
| `reuse_if_open` | bool | 否 | 兼容保留字段，默认 `true` |
| `enable_mcp` | bool | 否 | 启用 MCP 模式，默认 `false` |

## Python 调用脚本

详细脚本见 `scripts/dtool_playwright_api.py`。
