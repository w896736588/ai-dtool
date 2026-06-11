CREATE TABLE IF NOT EXISTS "tbl_agent_cli_prompt_template" (
    "id"                INTEGER PRIMARY KEY AUTOINCREMENT,
    "name"              TEXT    NOT NULL DEFAULT '',
    "content"           TEXT    NOT NULL DEFAULT '',
    "apply_all_dirs"    INTEGER NOT NULL DEFAULT 0,
    "sort_order"        INTEGER NOT NULL DEFAULT 0,
    "created_at"        INTEGER NOT NULL DEFAULT 0,
    "updated_at"        INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "tbl_agent_cli_prompt_template_dir_rel" (
    "id"           INTEGER PRIMARY KEY AUTOINCREMENT,
    "template_id"  INTEGER NOT NULL DEFAULT 0,
    "local_dir"    TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_agent_cli_prompt_template_sort
    ON tbl_agent_cli_prompt_template(sort_order, id);

CREATE INDEX IF NOT EXISTS idx_agent_cli_prompt_template_dir_rel_template
    ON tbl_agent_cli_prompt_template_dir_rel(template_id);

CREATE INDEX IF NOT EXISTS idx_agent_cli_prompt_template_dir_rel_dir
    ON tbl_agent_cli_prompt_template_dir_rel(local_dir);
