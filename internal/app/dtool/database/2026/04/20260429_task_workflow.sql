CREATE TABLE "tbl_task_workflow"
(
    "id"                      INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "home_task_id"            INTEGER NOT NULL DEFAULT 0,
    "status"                  TEXT    NOT NULL DEFAULT '',
    "current_stage"           TEXT    NOT NULL DEFAULT '',
    "requirement_fragment_id" TEXT    NOT NULL DEFAULT '',
    "dev_plan_fragment_id"    TEXT    NOT NULL DEFAULT '',
    "latest_plan_run_id"      INTEGER NOT NULL DEFAULT 0,
    "latest_test_run_id"      INTEGER NOT NULL DEFAULT 0,
    "base_branch"             TEXT    NOT NULL DEFAULT '',
    "feature_branch"          TEXT    NOT NULL DEFAULT '',
    "last_error"              TEXT    NOT NULL DEFAULT '',
    "create_time"             INTEGER NOT NULL DEFAULT 0,
    "update_time"             INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX "idx_task_workflow_home_task"
    ON "tbl_task_workflow" ("home_task_id" ASC);
