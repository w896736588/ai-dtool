-- tbl_mcp_agent_target: 目标智能体定义（如 claude_code）
CREATE TABLE IF NOT EXISTS "tbl_mcp_agent_target" (
    "id"              INTEGER PRIMARY KEY AUTOINCREMENT,
    "agent_name"      TEXT NOT NULL DEFAULT '',
    "config_filename" TEXT NOT NULL DEFAULT '',
    "config_dir"      TEXT NOT NULL DEFAULT '',
    "create_time"     INTEGER NOT NULL DEFAULT 0,
    "update_time"     INTEGER NOT NULL DEFAULT 0
);

-- tbl_mcp_binding: MCP 类型、目录映射与目标智能体之间的绑定关系
CREATE TABLE IF NOT EXISTS "tbl_mcp_binding" (
    "id"              INTEGER PRIMARY KEY AUTOINCREMENT,
    "mcp_type"        TEXT NOT NULL DEFAULT '',
    "mapping_id"      INTEGER NOT NULL DEFAULT 0,
    "agent_target_id" INTEGER NOT NULL DEFAULT 0,
    "create_time"     INTEGER NOT NULL DEFAULT 0,
    "update_time"     INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_mcp_binding_unique"
    ON "tbl_mcp_binding" ("mcp_type", "mapping_id", "agent_target_id");
