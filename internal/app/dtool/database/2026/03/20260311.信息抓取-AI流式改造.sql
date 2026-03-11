DROP INDEX IF EXISTS "idx_info_crawl_run_page_run_task_page";
DROP INDEX IF EXISTS "idx_info_crawl_run_task_time";
DROP INDEX IF EXISTS "idx_info_crawl_task_page_task_status_sort";
DROP INDEX IF EXISTS "idx_info_crawl_task_status_update";

DROP TABLE IF EXISTS "tbl_info_crawl_run_page";
DROP TABLE IF EXISTS "tbl_info_crawl_task_page";
DROP TABLE IF EXISTS "tbl_info_crawl_run";
DROP TABLE IF EXISTS "tbl_info_crawl_task";

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_task"
(
    "id"           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"         TEXT    NOT NULL DEFAULT '',
    "prompt"       TEXT    NOT NULL DEFAULT '',
    "ai_model_id"  INTEGER NOT NULL DEFAULT 0,
    "status"       INTEGER NOT NULL DEFAULT 1,
    "create_time"  INTEGER NOT NULL DEFAULT 0,
    "update_time"  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_run"
(
    "id"                 INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "task_id"            INTEGER NOT NULL DEFAULT 0,
    "status"             TEXT    NOT NULL DEFAULT 'running',
    "run_message"        TEXT    NOT NULL DEFAULT '',
    "prompt_snapshot"    TEXT    NOT NULL DEFAULT '',
    "ai_model_snapshot"  TEXT    NOT NULL DEFAULT '',
    "output_content"     TEXT    NOT NULL DEFAULT '',
    "error_message"      TEXT    NOT NULL DEFAULT '',
    "create_time"        INTEGER NOT NULL DEFAULT 0,
    "update_time"        INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS "idx_info_crawl_task_status_update"
    ON "tbl_info_crawl_task" ("status", "update_time");

CREATE INDEX IF NOT EXISTS "idx_info_crawl_run_task_time"
    ON "tbl_info_crawl_run" ("task_id", "create_time");
