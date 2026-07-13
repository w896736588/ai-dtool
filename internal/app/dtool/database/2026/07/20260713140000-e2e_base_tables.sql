-- E2E 自动化测试 - 基础表结构
-- 核心原则：所有版本共用同一张表，通过 JSON 中的 type + version 区分
-- 数据库不变原则：未来新增步骤/断言类型只需注册解析器和执行器，无需迁移

-- E2E 分组表
CREATE TABLE IF NOT EXISTS "tbl_e2e_group" (
    "id"                  INTEGER PRIMARY KEY AUTOINCREMENT,
    "name"                TEXT NOT NULL DEFAULT '',
    "workflow_task_id"    INTEGER NULL,
    "notification_enabled" INTEGER NOT NULL DEFAULT 0,
    "webhook_config_id"   INTEGER NOT NULL DEFAULT 0,
    "created_at"          INTEGER NOT NULL DEFAULT 0,
    "updated_at"          INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_group_workflow" ON "tbl_e2e_group" ("workflow_task_id");

-- E2E 用例表
-- steps/assertions/variables 都是 JSON 字段，允许灵活扩展
CREATE TABLE IF NOT EXISTS "tbl_e2e_case" (
    "id"                  INTEGER PRIMARY KEY AUTOINCREMENT,
    "group_id"            INTEGER NOT NULL,
    "name"                TEXT NOT NULL DEFAULT '',
    "env_url"             TEXT NOT NULL DEFAULT '',
    "env_base_url"        TEXT NOT NULL DEFAULT '',
    "steps"               TEXT NOT NULL DEFAULT '[]',
    "assertions"          TEXT NOT NULL DEFAULT '[]',
    "variables"           TEXT NOT NULL DEFAULT '{}',
    "notification_enabled" INTEGER NOT NULL DEFAULT 0,
    "timeout_seconds"     INTEGER NOT NULL DEFAULT 600,
    "tags"                TEXT NOT NULL DEFAULT '',
    "created_at"          INTEGER NOT NULL DEFAULT 0,
    "updated_at"          INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY ("group_id") REFERENCES "tbl_e2e_group"("id") ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_case_group" ON "tbl_e2e_case" ("group_id");
