-- tbl_chrome_devtools_config: Chrome DevTools MCP 调试端口配置
CREATE TABLE IF NOT EXISTS "tbl_chrome_devtools_config" (
    "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT NOT NULL DEFAULT '',
    "port"        INTEGER NOT NULL DEFAULT 0,
    "remark"      TEXT NOT NULL DEFAULT '',
    "is_used"     INTEGER NOT NULL DEFAULT 0,
    "create_time" INTEGER NOT NULL DEFAULT 0,
    "update_time" INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS "idx_chrome_devtools_config_port"
    ON "tbl_chrome_devtools_config" ("port");
