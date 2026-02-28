# Dashboard 命令系统 AI 执行指南（Git 版）

## 1. 目标

`Dashboard.vue` 的输入框承担统一入口：通过文本 + Tab 补全执行 Git/Docker/自定义网页/自定义脚本/终端输出等功能。

本次已先完成 Git：
- 支持 `/` 前缀和无前缀（例：`g`）
- 支持别名（例：`ch` 代表 `checkout`）
- 支持分步引导：动作 -> 项目/环境 -> 分支参数（仅需要时）
- 支持 Git 页面“按钮 + 更多操作”的命令化

## 2. 关键文件

- `web/src/config/commandConfig.js`
- `web/src/components/Dashboard.vue`
- `web/src/utils/base/git.js`

## 3. 当前 Git 命令能力

### 3.1 一级命令
- `git`（别名：`g`）

### 3.2 子命令
- `pull`（`pl`）
- `status`（`st`）
- `branch`（`br`/`current`）
- `log`（`lg`）
- `checkout`（`ch`/`co`）：需要“项目 + 分支名”
- `checkout-remote`（`chr`/`cor`）：需要“项目 + 远程分支名”
- `save-credentials`（`save`/`cred`）：需要“项目”
- `set-safe`（`safe`）：需要“项目”
- `view-config`（`cfg`）：跳转 `/Git` 查看文档
- `help`（`h`）：跳转 `/Git` 帮助

### 3.3 交互示例
- `g` + Tab -> `git`
- `git ch` + Tab -> `git checkout`
- `git checkout ` + Tab 选择项目
- `git checkout <项目> <分支>` + Enter 执行
- `git pull <项目>` + Enter 执行（无需分支）

## 4. 扩展命令时必须遵守的规则

### 4.1 先改配置，不要先改执行逻辑
在 `commandConfig.js` 里定义：
- `command`、`name`、`desc`
- `aliases`（短命令）
- `action`
- `needTarget` + `dynamicChildren`（是否要选目标）
- `needInput` + `inputPlaceholder`（是否要额外输入参数）

### 4.2 再改 Dashboard 执行映射
在 `Dashboard.vue`：
1. `executeAction` 增加 action 分发
2. 若为 Git API：在 `executeGitAction` 增加 case
3. 若需动态目标列表：补齐 `loadDynamicChildren` 及对应 `loadXxxList`

### 4.3 参数链路要求
- 需要目标时，必须先确认 `targetCmd.data` 存在
- 需要输入时，必须在执行前校验 `currentInputValue`
- 缺参数时，输出明确提示，不允许静默失败

## 5. 后续扩展到 Docker/脚本/终端输出的建议

- 先在 `commandConfig.js` 为每个动作补 `aliases`
- 统一模式：`动作 -> 目标 -> 可选参数`
- 统一错误文案：
  - 缺目标：`请先选择项目/环境`
  - 缺参数：`请输入...`
- 每新增动作，都要在 Dashboard 输出窗口给出开始语句和完成标记

## 6. PowerShell + UTF-8 强制规范

在 Windows 下让 AI 执行命令前，必须先设置 UTF-8，否则中文会乱码：

```powershell
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8
```

读取/写入文件必须显式使用 UTF-8（示例）：

```powershell
Get-Content -Path 'web/src/components/Dashboard.vue' -Encoding utf8
Set-Content -Path 'web/src/components/Dashboard.vue' -Value $text -Encoding utf8
```

## 7. 交给 AI 的标准任务模板

```text
请在 web/src/config/commandConfig.js 新增 xxx 命令（含 aliases），
并在 web/src/components/Dashboard.vue 完成动作映射与参数校验。
要求支持 Tab 补全、空格推进、Enter 执行，
并保持 PowerShell UTF-8 操作规范。
完成后汇报改动文件、执行链路、验证结果。
```
