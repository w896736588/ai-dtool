---
name: dtool-playwright
description: Use when working with the dtool smart-link / browser session module to open a logged-in browser session and continue automation through MCP or Playwright.
---

# dtool-playwright

## 这个 skill 可以做什么

- 调用 dtool 的 smart-link 登录能力打开目标页面
- 以 MCP 模式接管已登录浏览器会话
- 在需要时以 Playwright 持久化目录模式接管浏览器
- 用于登录后页面检查、后续自动化操作、抓取会话态页面内容

## 必要约束

- 调用前，先向用户确认 `base_url`、`Token`、`smart_link_id`、`label`，以及需要使用的账号
- 优先使用 MCP 模式；只有在确实需要时才使用原生 Playwright 接管
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 如果目标 smart-link 不保留用户数据，不能继续复用登录态
- 需要具体接口字段或接管脚本时，再去看 `scripts/dtool_playwright_api.py`

## 细节位置

- 浏览器会话打开与接管：`scripts/dtool_playwright_api.py`
