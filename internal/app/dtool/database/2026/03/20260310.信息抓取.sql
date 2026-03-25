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

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_task_page"
(
    "id"                   INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "task_id"              INTEGER NOT NULL DEFAULT 0,
    "name"                 TEXT    NOT NULL DEFAULT '',
    "url"                  TEXT    NOT NULL DEFAULT '',
    "note"                 TEXT    NOT NULL DEFAULT '',
    "login_check_selector" TEXT    NOT NULL DEFAULT '',
    "login_status"         INTEGER NOT NULL DEFAULT 0,
    "user_data_dir"        TEXT    NOT NULL DEFAULT '',
    "sort"                 INTEGER NOT NULL DEFAULT 0,
    "status"               INTEGER NOT NULL DEFAULT 1,
    "create_time"          INTEGER NOT NULL DEFAULT 0,
    "update_time"          INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_run"
(
    "id"                 INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "task_id"            INTEGER NOT NULL DEFAULT 0,
    "status"             TEXT    NOT NULL DEFAULT 'running',
    "run_message"        TEXT    NOT NULL DEFAULT '',
    "prompt_snapshot"    TEXT    NOT NULL DEFAULT '',
    "ai_model_snapshot"  TEXT    NOT NULL DEFAULT '',
    "planner_content"    TEXT    NOT NULL DEFAULT '',
    "summary_content"    TEXT    NOT NULL DEFAULT '',
    "page_total"         INTEGER NOT NULL DEFAULT 0,
    "page_success_total" INTEGER NOT NULL DEFAULT 0,
    "page_failed_total"  INTEGER NOT NULL DEFAULT 0,
    "create_time"        INTEGER NOT NULL DEFAULT 0,
    "update_time"        INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_run_page"
(
    "id"              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "run_id"          INTEGER NOT NULL DEFAULT 0,
    "task_page_id"    INTEGER NOT NULL DEFAULT 0,
    "page_name"       TEXT    NOT NULL DEFAULT '',
    "url"             TEXT    NOT NULL DEFAULT '',
    "status"          TEXT    NOT NULL DEFAULT '',
    "error_message"   TEXT    NOT NULL DEFAULT '',
    "planner_action"  TEXT    NOT NULL DEFAULT '',
    "execute_log"     TEXT    NOT NULL DEFAULT '',
    "raw_text"        TEXT    NOT NULL DEFAULT '',
    "raw_html"        TEXT    NOT NULL DEFAULT '',
    "screenshot_path" TEXT    NOT NULL DEFAULT '',
    "create_time"     INTEGER NOT NULL DEFAULT 0,
    "update_time"     INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS "idx_info_crawl_task_status_update"
    ON "tbl_info_crawl_task" ("status", "update_time");

CREATE INDEX IF NOT EXISTS "idx_info_crawl_task_page_task_status_sort"
    ON "tbl_info_crawl_task_page" ("task_id", "status", "sort");

CREATE INDEX IF NOT EXISTS "idx_info_crawl_run_task_time"
    ON "tbl_info_crawl_run" ("task_id", "create_time");

CREATE INDEX IF NOT EXISTS "idx_info_crawl_run_page_run_task_page"
    ON "tbl_info_crawl_run_page" ("run_id", "task_page_id");
