-- E2E 自动化测试 - 录制会话表
-- 用于录制时暂存用户操作的步骤序列

CREATE TABLE IF NOT EXISTS "tbl_e2e_record_session" (
    "id"              TEXT PRIMARY KEY,
    "case_id"         INTEGER NOT NULL DEFAULT 0,
    "env_url"         TEXT NOT NULL DEFAULT '',
    "env_base_url"    TEXT NOT NULL DEFAULT '',
    "name"            TEXT NOT NULL DEFAULT '',
    "steps"           TEXT NOT NULL DEFAULT '[]',
    "created_at"      INTEGER NOT NULL DEFAULT 0,
    "updated_at"      INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_case" ON "tbl_e2e_record_session" ("case_id");
