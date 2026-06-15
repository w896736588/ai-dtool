---
name: dtool-workflow
description: Use when the task involves updating workflow node status for a workflow_id through the dtool workflow API.
---

# dtool-workflow

## 这个 skill 可以做什么

- 更新工作流节点状态
- 支持将节点标记为 `pending`
- 支持将节点标记为 `running`
- 支持将节点标记为 `completed`
- 适用于需求分析、开发执行、接口生成、测试修复、代码检查、浏览器测试等步骤

## 必要约束

- 调用前，先向用户确认 `workflow_id`，以及 `base_url`、`Token`
- 必须使用合法的状态值（pending/running/completed），step key 支持固定步骤和自定义步骤（custom_{id}）
- 只更新当前明确执行的步骤，不猜测用户未要求变更的节点
- 需要具体 step key 或脚本入口时，再去看 `scripts/update_workflow_status.py`

## 细节位置

- 工作流节点状态更新脚本：`scripts/update_workflow_status.py`
