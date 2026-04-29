CREATE TABLE "tbl_async_task"
(
    "id"              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "task_type"       TEXT    NOT NULL DEFAULT '',
    "task_status"     TEXT    NOT NULL DEFAULT '',
    "title"           TEXT    NOT NULL DEFAULT '',
    "source_id"       TEXT    NOT NULL DEFAULT '',
    "request_payload" TEXT    NOT NULL DEFAULT '',
    "result_payload"  TEXT    NOT NULL DEFAULT '',
    "error_message"   TEXT    NOT NULL DEFAULT '',
    "create_time"     INTEGER NOT NULL DEFAULT 0,
    "start_time"      INTEGER NOT NULL DEFAULT 0,
    "finish_time"     INTEGER NOT NULL DEFAULT 0,
    "update_time"     INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX "idx_async_task_status_time"
    ON "tbl_async_task" (
                           "task_status" ASC,
                           "update_time" DESC
        );

CREATE INDEX "idx_async_task_type_time"
    ON "tbl_async_task" (
                           "task_type" ASC,
                           "create_time" DESC
        );
