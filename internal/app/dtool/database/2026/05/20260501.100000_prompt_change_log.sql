CREATE TABLE IF NOT EXISTS "tbl_prompt_change_log" (
    "id"            INTEGER PRIMARY KEY AUTOINCREMENT,
    "config_key"    TEXT NOT NULL DEFAULT '',
    "config_name"   TEXT NOT NULL DEFAULT '',
    "old_value"     TEXT NOT NULL DEFAULT '',
    "new_value"     TEXT NOT NULL DEFAULT '',
    "create_time"   INTEGER NOT NULL DEFAULT 0,
    "update_time"   INTEGER NOT NULL DEFAULT 0
);
