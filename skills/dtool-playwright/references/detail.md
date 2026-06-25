# dtool-playwright 详细说明

## 必要约束

- 调用前，先向用户确认 `base_url`、`Token`、`smart_link_id`，以及需要使用的账号
- 优先使用 MCP 模式；只有在确实需要时才使用原生 Playwright 接管
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 如果目标 smart-link 不保留用户数据，不能继续复用登录态
- 需要具体接口字段或接管脚本时，再去看 `scripts/` 下文件

## 文件索引

- 浏览器会话打开与接管：`scripts/dtool_playwright_api.py`
- 请求头抓取：`scripts/browser_api.py`
- 网页截图：`scripts/screenshot_api.py`
