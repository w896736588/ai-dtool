-- 钉钉机器人配置（Stream 模式所需）
CREATE TABLE IF NOT EXISTS "tbl_butler_bot_config" (
  "id"           INTEGER PRIMARY KEY AUTOINCREMENT,
  "platform"     TEXT NOT NULL DEFAULT 'dingtalk',
  "name"         TEXT NOT NULL DEFAULT '',
  "app_key"      TEXT NOT NULL DEFAULT '',
  "app_secret"   TEXT NOT NULL DEFAULT '',
  "robot_code"   TEXT NOT NULL DEFAULT '',
  "webhook_url"  TEXT NOT NULL DEFAULT '',
  "secret"       TEXT NOT NULL DEFAULT '',
  "status"       INTEGER NOT NULL DEFAULT 1,
  "created_at"   INTEGER NOT NULL DEFAULT 0,
  "updated_at"   INTEGER NOT NULL DEFAULT 0
);

-- 管家角色
CREATE TABLE IF NOT EXISTS "tbl_butler_role" (
  "id"             INTEGER PRIMARY KEY AUTOINCREMENT,
  "name"           TEXT NOT NULL DEFAULT '',
  "persona"        TEXT NOT NULL DEFAULT '',
  "tone"           TEXT NOT NULL DEFAULT '',
  "system_prompt"  TEXT NOT NULL DEFAULT '',
  "init_greeting"  TEXT NOT NULL DEFAULT '',
  "status"         INTEGER NOT NULL DEFAULT 1,
  "created_at"     INTEGER NOT NULL DEFAULT 0,
  "updated_at"     INTEGER NOT NULL DEFAULT 0
);

-- 管家运行参数
CREATE TABLE IF NOT EXISTS "tbl_butler_config" (
  "id"                       INTEGER PRIMARY KEY AUTOINCREMENT,
  "name"                     TEXT NOT NULL DEFAULT '',
  "role_id"                  INTEGER NOT NULL DEFAULT 0,
  "model_id"                 INTEGER NOT NULL DEFAULT 0,
  "fc_model_id"              INTEGER NOT NULL DEFAULT 0,
  "agent_cli_id"             INTEGER NOT NULL DEFAULT 0,
  "bot_config_id"            INTEGER NOT NULL DEFAULT 0,
  "active_timeout_minutes"   INTEGER NOT NULL DEFAULT 30,
  "max_history"              INTEGER NOT NULL DEFAULT 100,
  "auto_clean_on_new_topic"  INTEGER NOT NULL DEFAULT 1,
  "index_doc_path"           TEXT NOT NULL DEFAULT '',
  "auto_init_on_start"       INTEGER NOT NULL DEFAULT 1,
  "status"                   INTEGER NOT NULL DEFAULT 1,
  "created_at"               INTEGER NOT NULL DEFAULT 0,
  "updated_at"               INTEGER NOT NULL DEFAULT 0
);

-- 会话历史
CREATE TABLE IF NOT EXISTS "tbl_butler_message" (
  "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
  "session_id"  TEXT NOT NULL DEFAULT '',
  "role"        TEXT NOT NULL DEFAULT '',
  "content"     TEXT NOT NULL DEFAULT '',
  "token_count" INTEGER NOT NULL DEFAULT 0,
  "topic"       TEXT NOT NULL DEFAULT '',
  "created_at"  INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS "idx_butler_msg_session" ON "tbl_butler_message"("session_id", "id");

-- 管家任务记录
CREATE TABLE IF NOT EXISTS "tbl_butler_task" (
  "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
  "session_id"  TEXT NOT NULL DEFAULT '',
  "title"       TEXT NOT NULL DEFAULT '',
  "status"      TEXT NOT NULL DEFAULT 'pending',
  "plan"        TEXT NOT NULL DEFAULT '',
  "result"      TEXT NOT NULL DEFAULT '',
  "executor"    TEXT NOT NULL DEFAULT '',
  "created_at"  INTEGER NOT NULL DEFAULT 0,
  "updated_at"  INTEGER NOT NULL DEFAULT 0
);
