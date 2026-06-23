-- 清理 tbl_task_workflow 表中已被 tbl_task_workflow_document 和 step_prompts 替代的旧字段
-- 保留字段：id, home_task_id, status, current_stage, requirement_fragment_id, fragment_folder_name,
--          requirement_fetch_status, requirement_fetch_started_at, requirement_fetch_finished_at,
--          requirement_fetch_error, requirement_source_url, base_branch, feature_branch, last_error,
--          node_statuses, step_prompts, create_time, update_time

-- 删除已迁移到 step_prompts JSON 的旧 prompt 字段
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_requirement";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_api_dev";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_api_test";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_design";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_plain_text_requirement";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_design_plan_requirement";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_browser_test";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "prompt_code_review";

-- 删除已迁移到 tbl_task_workflow_document 的旧 fragment_id 字段（保留 requirement_fragment_id）
ALTER TABLE "tbl_task_workflow" DROP COLUMN "dev_plan_fragment_id";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "design_fragment_id";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "api_doc_fragment_id";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "plain_text_requirement_fragment_id";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "design_plan_requirement_fragment_id";

-- 删除冗余字段
ALTER TABLE "tbl_task_workflow" DROP COLUMN "latest_plan_run_id";
ALTER TABLE "tbl_task_workflow" DROP COLUMN "latest_test_run_id";

-- 删除已迁移到 tbl_task_workflow_document 的 step_fragment_refs JSON
ALTER TABLE "tbl_task_workflow" DROP COLUMN "step_fragment_refs";
