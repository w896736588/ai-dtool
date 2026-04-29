CREATE TABLE "tbl_task_test_run"
(
    "id"                      INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "workflow_id"             INTEGER NOT NULL DEFAULT 0,
    "run_no"                  TEXT    NOT NULL DEFAULT '',
    "run_type"                TEXT    NOT NULL DEFAULT '',
    "status"                  TEXT    NOT NULL DEFAULT '',
    "trigger_source"          TEXT    NOT NULL DEFAULT '',
    "requirement_snapshot_md" TEXT    NOT NULL DEFAULT '',
    "dev_plan_snapshot_md"    TEXT    NOT NULL DEFAULT '',
    "diff_snapshot_text"      TEXT    NOT NULL DEFAULT '',
    "coverage_report_json"    TEXT    NOT NULL DEFAULT '',
    "test_plan_json"          TEXT    NOT NULL DEFAULT '',
    "test_report_json"        TEXT    NOT NULL DEFAULT '',
    "summary_md"              TEXT    NOT NULL DEFAULT '',
    "started_at"              INTEGER NOT NULL DEFAULT 0,
    "finished_at"             INTEGER NOT NULL DEFAULT 0,
    "create_time"             INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX "idx_task_test_run_workflow"
    ON "tbl_task_test_run" ("workflow_id" ASC, "id" DESC);
