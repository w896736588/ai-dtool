CREATE TABLE IF NOT EXISTS "tbl_cron_task"
(
    "id"                integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"              text    NOT NULL DEFAULT '',
    "type"              text    NOT NULL DEFAULT '',
    "enabled"           integer NOT NULL DEFAULT 0,
    "trigger_time"      text    NOT NULL DEFAULT '',
    "last_trigger_time" integer NOT NULL DEFAULT 0,
    "create_time"       integer NOT NULL DEFAULT 0,
    "update_time"       integer NOT NULL DEFAULT 0
);

-- 从 tbl_global 迁移已有定时任务配置
INSERT INTO tbl_cron_task (name, type, enabled, trigger_time, create_time, update_time)
SELECT 'AI 生成工作日报', 'daily_report',
       CAST(COALESCE((SELECT value FROM tbl_global WHERE key = 'cron_daily_report_enabled'), '0') AS INTEGER),
       COALESCE((SELECT value FROM tbl_global WHERE key = 'cron_daily_report_time'), ''),
       strftime('%s', 'now'), strftime('%s', 'now')
WHERE NOT EXISTS (SELECT 1 FROM tbl_cron_task WHERE type = 'daily_report');

-- 清理 tbl_global 中已迁移的定时任务配置
DELETE FROM tbl_global WHERE key = 'cron_daily_report_enabled';
DELETE FROM tbl_global WHERE key = 'cron_daily_report_time';
