-- 自定义网页本地客户端相关表
-- Smart Link Local Client Tables

-- 本地客户端表
CREATE TABLE "tbl_smart_link_client"
(
    "id"               integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "client_id"        text    NOT NULL DEFAULT '',
    "client_name"      text    NOT NULL DEFAULT '',
    "client_version"   text    NOT NULL DEFAULT '',
    "required_version" text    NOT NULL DEFAULT '',
    "status"           text    NOT NULL DEFAULT 'offline',
    "host_name"        text    NOT NULL DEFAULT '',
    "os"               text    NOT NULL DEFAULT '',
    "arch"             text    NOT NULL DEFAULT '',
    "user_name"        text    NOT NULL DEFAULT '',
    "last_seen_time"   integer NOT NULL DEFAULT 0,
    "create_time"      integer NOT NULL DEFAULT 0,
    "update_time"      integer NOT NULL DEFAULT 0
);

-- 本地客户端任务表
CREATE TABLE "tbl_smart_link_task"
(
    "id"              integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "task_id"         text    NOT NULL DEFAULT '',
    "client_id"       text    NOT NULL DEFAULT '',
    "smart_link_id"   integer NOT NULL DEFAULT 0,
    "label"           text    NOT NULL DEFAULT '',
    "status"          text    NOT NULL DEFAULT 'pending',
    "run_mode"        text    NOT NULL DEFAULT 'local_client',
    "request_payload" text    NOT NULL DEFAULT '{}',
    "result_payload"  text    NOT NULL DEFAULT '{}',
    "error_message"   text    NOT NULL DEFAULT '',
    "log_text"        text    NOT NULL DEFAULT '',
    "create_time"     integer NOT NULL DEFAULT 0,
    "start_time"      integer NOT NULL DEFAULT 0,
    "finish_time"     integer NOT NULL DEFAULT 0,
    "update_time"     integer NOT NULL DEFAULT 0
);

-- 创建索引
CREATE INDEX "idx_smart_link_client_client_id" ON "tbl_smart_link_client" ("client_id");
CREATE INDEX "idx_smart_link_client_status" ON "tbl_smart_link_client" ("status");
CREATE INDEX "idx_smart_link_task_task_id" ON "tbl_smart_link_task" ("task_id");
CREATE INDEX "idx_smart_link_task_client_id" ON "tbl_smart_link_task" ("client_id");
CREATE INDEX "idx_smart_link_task_status" ON "tbl_smart_link_task" ("status");
CREATE INDEX "idx_smart_link_task_smart_link_id" ON "tbl_smart_link_task" ("smart_link_id");
