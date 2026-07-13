-- E2E 自动化测试 - 执行记录相关表
-- 用于存储每次用例执行的步骤、断言、捕获请求、截图等详细记录

-- E2E 执行记录表
CREATE TABLE IF NOT EXISTS "tbl_e2e_run" (
    "id"              INTEGER PRIMARY KEY AUTOINCREMENT,
    "case_id"         INTEGER NOT NULL,
    "group_id"        INTEGER NOT NULL DEFAULT 0,
    "status"          TEXT NOT NULL DEFAULT 'pending',
    "total_steps"     INTEGER NOT NULL DEFAULT 0,
    "passed_steps"    INTEGER NOT NULL DEFAULT 0,
    "failed_steps"    INTEGER NOT NULL DEFAULT 0,
    "total_asserts"   INTEGER NOT NULL DEFAULT 0,
    "passed_asserts"  INTEGER NOT NULL DEFAULT 0,
    "failed_asserts"  INTEGER NOT NULL DEFAULT 0,
    "started_at"      INTEGER NOT NULL DEFAULT 0,
    "finished_at"     INTEGER NOT NULL DEFAULT 0,
    "duration_ms"     INTEGER NOT NULL DEFAULT 0,
    "error_message"   TEXT,
    "log_stream"      TEXT,
    "trigger_type"    TEXT NOT NULL DEFAULT 'manual',
    "created_at"      INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY ("case_id") REFERENCES "tbl_e2e_case"("id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_run_case" ON "tbl_e2e_run" ("case_id");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_run_group" ON "tbl_e2e_run" ("group_id");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_run_status" ON "tbl_e2e_run" ("status");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_run_started" ON "tbl_e2e_run" ("started_at" DESC);

-- E2E 执行步骤记录表
CREATE TABLE IF NOT EXISTS "tbl_e2e_run_step" (
    "id"              INTEGER PRIMARY KEY AUTOINCREMENT,
    "run_id"          INTEGER NOT NULL,
    "step_index"      INTEGER NOT NULL DEFAULT 0,
    "step_id"         TEXT NOT NULL DEFAULT '',
    "step_type"       TEXT NOT NULL DEFAULT '',
    "step_version"    TEXT NOT NULL DEFAULT '1.0',
    "step_config"     TEXT NOT NULL DEFAULT '{}',
    "description"     TEXT NOT NULL DEFAULT '',
    "status"          TEXT NOT NULL DEFAULT 'pending',
    "error_message"   TEXT,
    "screenshot_path" TEXT,
    "extracted_vars"  TEXT,
    "duration_ms"     INTEGER NOT NULL DEFAULT 0,
    "executed_at"     INTEGER NOT NULL DEFAULT 0,
    "created_at"      INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY ("run_id") REFERENCES "tbl_e2e_run"("id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_run_step_run" ON "tbl_e2e_run_step" ("run_id");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_run_step_step_id" ON "tbl_e2e_run_step" ("run_id", "step_id");

-- E2E 断言执行记录表
CREATE TABLE IF NOT EXISTS "tbl_e2e_run_assertion" (
    "id"                   INTEGER PRIMARY KEY AUTOINCREMENT,
    "run_id"               INTEGER NOT NULL,
    "run_step_id"          INTEGER NOT NULL DEFAULT 0,
    "assertion_id"         TEXT NOT NULL DEFAULT '',
    "assertion_type"       TEXT NOT NULL DEFAULT '',
    "assertion_version"    TEXT NOT NULL DEFAULT '1.0',
    "assertion_config"     TEXT NOT NULL DEFAULT '{}',
    "status"               TEXT NOT NULL DEFAULT 'pending',
    "expected"             TEXT,
    "actual"               TEXT,
    "error_message"        TEXT,
    "matched_request_url"  TEXT,
    "matched_request_id"   TEXT,
    "executed_at"          INTEGER NOT NULL DEFAULT 0,
    "created_at"           INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY ("run_id") REFERENCES "tbl_e2e_run"("id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_run_assertion_run" ON "tbl_e2e_run_assertion" ("run_id");

-- E2E 捕获请求表
-- 用于存储每次执行期间通过 Playwright route 捕获到的所有 XHR/Fetch 请求
CREATE TABLE IF NOT EXISTS "tbl_e2e_captured_request" (
    "id"                TEXT PRIMARY KEY,
    "run_id"            INTEGER NOT NULL,
    "run_step_id"       INTEGER NOT NULL DEFAULT 0,
    "url"               TEXT NOT NULL DEFAULT '',
    "method"            TEXT NOT NULL DEFAULT '',
    "request_headers"   TEXT NOT NULL DEFAULT '{}',
    "request_body"      TEXT,
    "response_status"   INTEGER NOT NULL DEFAULT 0,
    "response_headers"  TEXT NOT NULL DEFAULT '{}',
    "response_body"     TEXT,
    "response_time_ms"  INTEGER NOT NULL DEFAULT 0,
    "matched"           INTEGER NOT NULL DEFAULT 0,
    "matched_by"        TEXT,
    "captured_at"       INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY ("run_id") REFERENCES "tbl_e2e_run"("id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_captured_request_run" ON "tbl_e2e_captured_request" ("run_id");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_captured_request_step" ON "tbl_e2e_captured_request" ("run_step_id");
