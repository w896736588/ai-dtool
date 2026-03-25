CREATE TABLE IF NOT EXISTS "tbl_home_task"
(
    "id"               INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"             TEXT    NOT NULL DEFAULT '',
    "task_status"      TEXT    NOT NULL DEFAULT '待开始',
    "is_archived"      INTEGER NOT NULL DEFAULT 0,
    "start_time"       INTEGER NOT NULL DEFAULT 0,
    "last_operated_at" INTEGER NOT NULL DEFAULT 0,
    "create_time"      INTEGER NOT NULL DEFAULT 0,
    "update_time"      INTEGER NOT NULL DEFAULT 0
);
