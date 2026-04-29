CREATE TABLE IF NOT EXISTS "tbl_home_task_config"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "key"         text    NOT NULL DEFAULT '',
    "value"       text    NOT NULL DEFAULT '',
    "name"        text    NOT NULL DEFAULT '',
    "desc"        text    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);

-- 从 tbl_global 迁移已有首页任务配置
INSERT INTO tbl_home_task_config (key, value, name, desc, create_time, update_time)
SELECT 'home_task_daily_report_prompt', COALESCE(value, ''), '工作日报提示词', '首页任务工作日报 AI 提示词',
       strftime('%s', 'now'), strftime('%s', 'now')
FROM tbl_global WHERE key = 'home_task_daily_report_prompt'
WHERE NOT EXISTS (SELECT 1 FROM tbl_home_task_config WHERE key = 'home_task_daily_report_prompt');

INSERT INTO tbl_home_task_config (key, value, name, desc, create_time, update_time)
SELECT 'home_task_daily_report_model_id', COALESCE(value, ''), '工作日报模型', '首页任务工作日报所用模型 id',
       strftime('%s', 'now'), strftime('%s', 'now')
FROM tbl_global WHERE key = 'home_task_daily_report_model_id'
WHERE NOT EXISTS (SELECT 1 FROM tbl_home_task_config WHERE key = 'home_task_daily_report_model_id');

INSERT INTO tbl_home_task_config (key, value, name, desc, create_time, update_time)
SELECT 'home_task_fragment_prompt', COALESCE(value, ''), '任务知识片段提示词', '新建任务时自动创建知识片段的提示词模板',
       strftime('%s', 'now'), strftime('%s', 'now')
FROM tbl_global WHERE key = 'home_task_fragment_prompt'
WHERE NOT EXISTS (SELECT 1 FROM tbl_home_task_config WHERE key = 'home_task_fragment_prompt');

-- 清理 tbl_global 中已迁移的首页任务配置
DELETE FROM tbl_global WHERE key = 'home_task_daily_report_prompt';
DELETE FROM tbl_global WHERE key = 'home_task_daily_report_model_id';
DELETE FROM tbl_global WHERE key = 'home_task_fragment_prompt';
