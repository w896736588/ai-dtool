CREATE TABLE IF NOT EXISTS "tbl_ai_provider"
(
    "id"            INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"          TEXT    NOT NULL DEFAULT '',
    "provider_type" TEXT    NOT NULL DEFAULT 'openai',
    "base_url"      TEXT    NOT NULL DEFAULT '',
    "api_key"       TEXT    NOT NULL DEFAULT '',
    "status"        INTEGER NOT NULL DEFAULT 1,
    "create_time"   INTEGER NOT NULL DEFAULT 0,
    "update_time"   INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "tbl_ai_model"
(
    "id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "provider_id" INTEGER NOT NULL DEFAULT 0,
    "name"        TEXT    NOT NULL DEFAULT '',
    "model"       TEXT    NOT NULL DEFAULT '',
    "status"      INTEGER NOT NULL DEFAULT 1,
    "create_time" INTEGER NOT NULL DEFAULT 0,
    "update_time" INTEGER NOT NULL DEFAULT 0
);

