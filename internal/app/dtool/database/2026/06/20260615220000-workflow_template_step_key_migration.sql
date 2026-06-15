-- 工作流模板步骤 step_key 统一为 custom_{id} 格式
-- 旧版默认模板使用了旧式 key（requirement、design、api-dev 等），统一改为 custom_{id}

-- 1. 更新默认模板中非固定步骤的 step_key 为新格式
UPDATE "tbl_workflow_template_step"
SET "step_key" = 'custom_' || CAST("id" AS TEXT),
    "update_time" = strftime('%s','now')
WHERE "template_id" = 1
  AND "is_fixed" = 0
  AND "step_key" NOT LIKE 'custom_%';

-- 注意：tbl_task_workflow 中的 step_prompts 和 node_statuses JSON 的 key 需要在
-- Go 代码中迁移（SQLite 不支持动态 JSON key 重命名）
