# 首页 script 自定义脚本命令 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在首页新增 `script` 顶级命令并移除首页 `variable` 顶级命令，让自定义脚本执行完全由命令框输入和候选项驱动。

**Architecture:** 通过调整首页命令配置和 `Dashboard.vue` 的命令框状态流，引入独立 `scriptSession` 会话状态机。普通命令继续沿用原有逻辑，只有 `script` 链路进入脚本会话模式，按当前阶段决定候选展示与回车行为。

**Tech Stack:** Vue 3、Element Plus、现有首页命令配置与前端测试用例

---

### Task 1: 调整首页命令配置

**Files:**
- Modify: `web/src/config/commandConfig.js`
- Test: `web/test/` 下与首页命令候选相关的回归测试

**Step 1: Write the failing test**

补一个首页命令候选测试，断言：

- `script` 出现在顶级命令中
- `variable` 不再出现在顶级命令中

**Step 2: Run test to verify it fails**

Run: `npm test -- web/test/<new-or-updated-test>.cjs`

Expected: FAIL，提示仍然存在 `variable` 或不存在 `script`

**Step 3: Write minimal implementation**

在 `web/src/config/commandConfig.js` 中：

- 删除 `variable` 顶级命令配置
- 新增 `script` 顶级命令配置
- `script` 仅保留 `run` 子命令
- 为 `run` 指定独立动态候选键 `scriptList`

**Step 4: Run test to verify it passes**

Run: `npm test -- web/test/<new-or-updated-test>.cjs`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/config/commandConfig.js web/test/<new-or-updated-test>.cjs
git commit -m "feat: replace dashboard variable command with script"
```

### Task 2: 为 script 会话补失败测试

**Files:**
- Modify: `web/test/` 下首页命令相关测试
- Reference: `web/src/components/Dashboard.vue`

**Step 1: Write the failing test**

补测试覆盖以下行为：

- 输入 `script run` 后，下拉显示脚本列表
- 启动脚本后若返回输入步骤，placeholder 切换为步骤输入提示
- 启动脚本后若返回选项步骤，下拉显示步骤选项

**Step 2: Run test to verify it fails**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: FAIL，提示当前 `Dashboard.vue` 不支持 `scriptSession` 交互流

**Step 3: Write minimal implementation**

先只在测试中把需要 mock 的脚本列表和脚本步骤响应整理好，不改生产代码。

**Step 4: Run test to verify it still fails for the right reason**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: FAIL，且失败原因是页面逻辑缺失，不是测试写错

**Step 5: Commit**

```bash
git add web/test/<script-session-test>.cjs
git commit -m "test: cover dashboard script command session flow"
```

### Task 3: 在 Dashboard 中引入独立 scriptSession

**Files:**
- Modify: `web/src/components/Dashboard.vue`
- Test: `web/test/<script-session-test>.cjs`

**Step 1: Write the failing test**

补一个更细的状态流测试，断言：

- `script run` 进入 `selecting_script`
- 选择脚本后，能根据响应进入 `waiting_input` 或 `waiting_option`

**Step 2: Run test to verify it fails**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: FAIL，提示缺少 `scriptSession` 或状态切换不正确

**Step 3: Write minimal implementation**

在 `Dashboard.vue` 中：

- 新增 `scriptSession`
- 增加脚本会话的阶段常量或阶段判断
- 增加进入脚本模式与重置脚本会话的基础方法
- 让 `script` 命令进入脚本选择态

**Step 4: Run test to verify it passes**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Dashboard.vue web/test/<script-session-test>.cjs
git commit -m "feat: add dashboard script session state"
```

### Task 4: 接入 script 脚本列表候选

**Files:**
- Modify: `web/src/components/Dashboard.vue`
- Modify: 如有需要，新增或修改脚本 API 封装文件
- Test: `web/test/<script-session-test>.cjs`

**Step 1: Write the failing test**

断言：

- `script run` 后命令框下拉只显示脚本列表
- 支持按脚本名过滤候选
- 选择脚本后触发启动动作

**Step 2: Run test to verify it fails**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: FAIL，提示候选来源或选择行为不正确

**Step 3: Write minimal implementation**

在 `Dashboard.vue` 中：

- 新增 `scriptList` 加载逻辑
- 在 `selecting_script` 阶段将 `currentChildren` 切换到脚本候选
- 选择脚本后调用启动逻辑

**Step 4: Run test to verify it passes**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Dashboard.vue web/test/<script-session-test>.cjs
git commit -m "feat: support script list selection on dashboard"
```

### Task 5: 支持输入步骤与选项步骤

**Files:**
- Modify: `web/src/components/Dashboard.vue`
- Test: `web/test/<script-session-test>.cjs`

**Step 1: Write the failing test**

断言：

- 输入步骤时，命令框直接接收用户输入并提交
- 选项步骤时，命令框下拉显示步骤选项并支持确认

**Step 2: Run test to verify it fails**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: FAIL，提示回车行为或候选来源错误

**Step 3: Write minimal implementation**

在 `Dashboard.vue` 中：

- 按 `scriptSession.stage` 区分回车行为
- `waiting_input` 时将输入框内容直接视作步骤值
- `waiting_option` 时将候选切换为当前步骤选项
- 提交后按响应推进脚本状态

**Step 4: Run test to verify it passes**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Dashboard.vue web/test/<script-session-test>.cjs
git commit -m "feat: support script input and option steps"
```

### Task 6: 支持脚本就绪执行与会话清理

**Files:**
- Modify: `web/src/components/Dashboard.vue`
- Test: `web/test/<script-session-test>.cjs`

**Step 1: Write the failing test**

断言：

- 脚本进入可执行状态后，回车直接执行
- 执行完成后恢复首页默认态
- 执行中禁止重复提交

**Step 2: Run test to verify it fails**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: FAIL，提示 `ready_execute` 或 `finished` 流程缺失

**Step 3: Write minimal implementation**

在 `Dashboard.vue` 中：

- 增加 `ready_execute / executing / finished` 状态处理
- 执行完成后清理 `scriptSession`
- 恢复默认 placeholder、候选和命令框行为

**Step 4: Run test to verify it passes**

Run: `npm test -- web/test/<script-session-test>.cjs`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Dashboard.vue web/test/<script-session-test>.cjs
git commit -m "feat: execute script from dashboard session"
```

### Task 7: 统一首页文案并做回归验证

**Files:**
- Modify: `web/src/components/Dashboard.vue`
- Test: `web/test/` 下首页命令与结果文案相关测试

**Step 1: Write the failing test**

断言：

- 首页脚本链路的 placeholder、结果区提示、候选提示不再包含 `variable`
- 其他命令如 `git`、`docker` 仍按原逻辑工作

**Step 2: Run test to verify it fails**

Run: `npm test -- web/test/<script-session-test>.cjs web/test/<existing-dashboard-tests>.cjs`

Expected: FAIL，提示旧文案残留或普通命令行为被破坏

**Step 3: Write minimal implementation**

在 `Dashboard.vue` 中：

- 统一脚本链路文案为 `script/脚本`
- 修正脚本会话与普通命令模式的切换条件

**Step 4: Run test to verify it passes**

Run: `npm test -- web/test/<script-session-test>.cjs web/test/<existing-dashboard-tests>.cjs`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Dashboard.vue web/test/<script-session-test>.cjs web/test/<existing-dashboard-tests>.cjs
git commit -m "refactor: align dashboard script command copy"
```
