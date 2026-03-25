# GitLogs 今日任务漏抓优化方案

## 背景

当前 `GitLogs` 的 GitLab 聚合逻辑位于 `internal/pkg/p_gitlab/gitlab.go`。

本次仅针对以下两个已确认的漏抓点做最小影响优化：

1. 第 3 点：MR 必须满足 `created_at / updated_at / merged_at` 之一在今日，才会进入分支提交检查，导致“今天在老分支继续开发，但 MR 时间字段没更新”时漏抓。
2. 第 7 点：`开发中 / 对接中` 依赖 commit 信息中是否包含固定文案 `Merge branch`，对 squash、rebase、cherry-pick、机器人提交等场景兼容性不足。

## 目标

1. 提高“今日参与任务”的召回率，覆盖“老 MR 今日仍有提交”的场景。
2. 降低对固定 merge commit 文案的依赖，提高不同 Git 工作流下的稳定性。
3. 保持现有接口、现有返回结构、现有状态文案不变，避免影响前端展示和调用方式。

## 现状问题

### 问题 1：时间门槛过早

当前 `checkMerges()` 先用 MR 时间字段做一次硬过滤：

- `createdToday`
- `updatedToday`
- `mergedToday`

只有三者之一命中，才会继续调用 `checkMergeUserOp()`。

这会漏掉：

- MR 是前几天创建的
- 今天开发者继续向源分支提交代码
- 但 GitLab 的 MR `updated_at` 未按预期反映本次提交

### 问题 2：merge commit 识别规则过窄

当前 `isMergeBranch(message string)` 仅判断：

- 是否包含 `Merge branch`

这会导致：

- GitLab 的其他 merge message 形式无法识别
- squash / rebase 后没有标准 merge commit 文案时误判
- 自动化账号或特殊提交策略下，把非开发提交算成“开发中 / 对接中”

## 优化方案

## 方案 A：放宽 MR 入围条件，增加“分支今日活跃”兜底

### 思路

保留现有基于 MR 时间字段的快速过滤，但新增“分支今日活跃”判定：

- 若 MR 自身时间命中今日，沿用现有逻辑
- 若 MR 时间未命中今日，则继续检查该 MR 源分支今日是否存在有效提交
- 只要源分支今日存在作者提交或其他人提交，即认为该 MR 应进入状态计算

### 实现方式

建议将 `checkMergeUserOp()` 的返回值从：

- `authorJoin`
- `authorCommit`
- `otherCommitToday`

扩展为一个轻量结果结构，例如：

- `authorJoin`
- `authorCommitToday`
- `otherCommitToday`
- `branchActiveToday`

其中：

- `branchActiveToday = authorCommitToday || otherCommitToday || authorMergeLikeCommitToday`

然后在 `checkMerges()` 中把 `relevantByTime` 由“只看 MR 时间”调整为：

- `mrTimeMatched || branchActiveToday`

### 优点

1. 对现有主流程侵入小，只是在现有 `checkMergeUserOp()` 结果上追加一个兜底条件。
2. 不改接口入参和返回结构，前端无感知。
3. 能补上“老 MR 今日继续开发”的主要漏抓场景。

### 风险

1. 会比现在多读取一部分 MR 的源分支提交，API 请求量略增。
2. 如果某些老 MR 长期挂着，今天分支有机器人提交，也可能被纳入候选；需要配合方案 B 降低误判。

## 方案 B：把“是否为合并类提交”从单一文案判断升级为多特征判断

### 思路

不再只依赖 `Merge branch`，改成“合并类提交识别器”：

优先根据 commit message 识别以下模式：

- `Merge branch`
- `Merge remote-tracking branch`
- `Merge pull request`
- `Merged in`
- `See merge request`

同时增加兜底规则：

- 若 message 首行以 `merge ` 开头，视为合并类提交
- 若 message 明显是自动合并/同步主干文案，也视为合并类提交

### 实现方式

将 `isMergeBranch()` 重构为更通用的 `isMergeLikeCommit(message string)`，内部做：

1. `TrimSpace`
2. 统一转小写
3. 用关键词数组匹配
4. 对前缀 `merge ` 做补充判断

然后在 `checkMergeUserOp()` 中继续复用该函数，替代原有单点判断。

### 优点

1. 能覆盖更多 GitLab / GitHub / 手工合并文案。
2. 仍然保持“开发提交”和“合并类提交”的现有业务语义，不需要重做状态体系。
3. 改动集中，测试点清晰。

### 风险

1. 关键词过宽时，可能把少量普通提交误判为合并类提交。
2. 需要控制关键词集合，优先选典型模式，避免过度泛化。

## 推荐落地方案

推荐采用：

- 方案 A + 方案 B 组合落地

原因：

1. 方案 A 解决“为什么今日明明参与了却没被纳入候选”的问题。
2. 方案 B 解决“纳入候选后状态判断不稳定”的问题。
3. 两者都能基于现有函数微调完成，不需要新增模块，也不需要改前端。

## 拟修改点

1. `internal/pkg/p_gitlab/gitlab.go`
2. 主要调整函数：
   - `checkMerges`
   - `checkMergeUserOp`
   - `isMergeBranch`（建议升级命名与实现）

## 兼容性要求

1. 保持 `Combine{Message, Status}` 返回结构不变。
2. 保持状态值仍为：
   - `已上线`
   - `对接中`
   - `开发中`
3. 保持 `GitLogs` 控制器和 `/api/GitLab` 接口协议不变。

## 验证方案

### 用例 1：老 MR 今日继续开发

- MR 创建于昨天
- 今天作者向源分支新增普通提交
- 预期：应被统计为 `开发中`

### 用例 2：老 MR 今日被他人接手

- MR 创建于昨天
- 今天其他成员向源分支新增普通提交
- 作者本人历史上参与过该分支
- 预期：应被统计为 `对接中`

### 用例 3：今日只有 merge 类提交

- 今天只有同步主干或合并类提交
- 没有实际开发提交
- 预期：不应误判为 `开发中`

### 用例 4：squash / rebase 场景

- 提交信息不包含 `Merge branch`
- 但属于明确的合并类文案
- 预期：应按合并类提交排除，不算开发提交

## 回滚点

若优化后召回过宽，可优先回滚以下两处：

1. 取消 `branchActiveToday` 兜底，仅保留 MR 时间过滤
2. 缩减 merge-like 关键词集合，仅保留最稳定的 2 到 3 条规则

## GitNexus 使用结论

本次使用了 GitNexus 的：

1. `impact`
2. `context`

关键结论：

1. `GetTodayLogs` 上游影响范围很小，仅通过 `GitLogs -> apiUse` 被使用。
2. 风险等级为 `LOW`，适合做最小影响优化。

但从“方案设计收益”角度看，这次 GitNexus 的帮助主要在确认影响范围；对第 3、7 点这类单文件过滤逻辑优化，源码直读的信息密度更高，GitNexus 有辅助价值，但不是必需，且会带来额外 token 消耗。
