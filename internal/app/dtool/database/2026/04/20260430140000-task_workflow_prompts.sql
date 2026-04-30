ALTER TABLE tbl_task_workflow ADD COLUMN prompt_requirement TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_task_workflow ADD COLUMN prompt_api_dev TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_task_workflow ADD COLUMN prompt_api_test TEXT NOT NULL DEFAULT '';
