-- AgentCli 专用分组表
CREATE TABLE IF NOT EXISTS "tbl_agent_cli_group" (
    "id"         INTEGER PRIMARY KEY AUTOINCREMENT,
    "name"       TEXT    NOT NULL DEFAULT '',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "created_at" INTEGER NOT NULL DEFAULT 0,
    "updated_at" INTEGER NOT NULL DEFAULT 0
);

-- AgentCli 与分组的多对多关联表
CREATE TABLE IF NOT EXISTS "tbl_agent_cli_group_rel" (
    "id"           INTEGER PRIMARY KEY AUTOINCREMENT,
    "agent_cli_id" INTEGER NOT NULL DEFAULT 0,
    "group_id"     INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_agent_cli_group_rel_cli ON tbl_agent_cli_group_rel(agent_cli_id);
CREATE INDEX IF NOT EXISTS idx_agent_cli_group_rel_group ON tbl_agent_cli_group_rel(group_id);
