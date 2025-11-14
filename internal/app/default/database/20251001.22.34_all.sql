CREATE TABLE "tbl_account"
(
    "id"               integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "username"         text    NOT NULL DEFAULT '',
    "password"         text    NOT NULL DEFAULT 0,
    "account_group_id" integer NOT NULL DEFAULT 0,
    "create_time"      integer NOT NULL DEFAULT 0,
    "update_time"      integer NOT NULL DEFAULT 0
);


CREATE TABLE "tbl_cmd"
(
    "id"          INTEGER NOT NULL,
    "name"        TEXT    NOT NULL DEFAULT '',
    "command"     TEXT    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL,
    "update_time" integer NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "tbl_docker_compose"
(
    "id"               INTEGER PRIMARY KEY AUTOINCREMENT,
    "name"             TEXT          DEFAULT '',
    "compose_yml_path" TEXT          DEFAULT '',
    "status"           integer       DEFAULT 1,
    "create_time"      integer,
    "update_time"      integer,
    "ssh_id"           INTEGER       DEFAULT 0,
    "docker_cmd"       TEXT          DEFAULT 'docker compose',
    "env_file"         TEXT          DEFAULT ''
);

CREATE TABLE "tbl_git"
(
    "id"           integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"         text    NOT NULL DEFAULT '',
    "ssh_id"       integer NOT NULL DEFAULT 0,
    "code_path"    text    NOT NULL DEFAULT '',
    "git_group_id" integer NOT NULL DEFAULT 0,
    "create_time"  integer NOT NULL DEFAULT 0,
    "update_time"  integer NOT NULL DEFAULT 0,
    "assign_check" integer NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_gitlab_token"
(
    "id"           integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"         text    NOT NULL DEFAULT '',
    "url"          text    NOT NULL DEFAULT '',
    "access_token" text    NOT NULL DEFAULT '',
    "create_time"  integer NOT NULL DEFAULT 0,
    "update_time"  integer NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_global"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        text    NOT NULL DEFAULT '',
    "key"         text    NOT NULL DEFAULT '',
    "value"       text    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_group"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0,
    "type"        integer NOT NULL
);

CREATE TABLE "tbl_markdown"
(
    "id"            INTEGER PRIMARY KEY AUTOINCREMENT,
    "name"          TEXT    NOT NULL DEFAULT '',
    "content"       text    NOT NULL DEFAULT '',
    "create_time"   integer NOT NULL DEFAULT 0,
    "update_time"   integer NOT NULL DEFAULT 0,
    "markdown_type" TEXT    NOT NULL DEFAULT 'normal',
    "weight"        integer NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_markdown_history"
(
    "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
    "markdown_id" INTEGER NOT NULL,
    "old_content" TEXT    NOT NULL,
    "new_content" TEXT    NOT NULL,
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0,
    "change_desc" TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE "tbl_mysql"
(
    "id"          integer NOT NULL DEFAULT '' PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT    NOT NULL DEFAULT '',
    "host"        TEXT    NOT NULL DEFAULT '',
    "port"        TEXT    NOT NULL DEFAULT '',
    "username"    TEXT    NOT NULL DEFAULT '',
    "password"    TEXT    NOT NULL DEFAULT '',
    "ssh_id"      INTEGER NOT NULL DEFAULT 0,
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0,
    "dbname"      TEXT    NOT NULL
);

CREATE TABLE "tbl_redis"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT    NOT NULL DEFAULT '',
    "host"        TEXT    NOT NULL DEFAULT '',
    "port"        TEXT    NOT NULL DEFAULT '',
    "username"    TEXT    NOT NULL DEFAULT '',
    "password"    TEXT    NOT NULL DEFAULT '',
    "ssh_id"      INTEGER NOT NULL DEFAULT 0,
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_replace_param"
(
    "id"          INTEGER NOT NULL,
    "name"        TEXT    NOT NULL DEFAULT '',
    "type"        integer NOT NULL DEFAULT 0,
    "config"      TEXT    NOT NULL,
    "create_time" integer NOT NULL,
    "update_time" integer NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "tbl_smart_link"
(
    "id"                  INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"                TEXT,
    "smart_link_group_id" INTEGER,
    "links"               TEXT,
    "open_num"            INTEGER,
    "open_type"           TEXT,
    "process"             TEXT,
    "weight"              integer,
    "combine_type"        integer,
    "status"              integer          DEFAULT 1,
    "value"               TEXT,
    "create_time"         integer,
    "update_time"         integer,
    "download_finds"      TEXT             DEFAULT '',
    "auto_close_second"   integer          DEFAULT 0,
    "channel"             TEXT             DEFAULT '',
    "show_cookies"        TEXT             DEFAULT '',
    "process_id"          INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_smart_link_last"
(
    "id"              INTEGER NOT NULL,
    "smart_link_id"   INTEGER NOT NULL DEFAULT 0,
    "user_name"       TEXT    NOT NULL DEFAULT '',
    "user_data_index" text    NOT NULL DEFAULT '',
    "domain"          TEXT    NOT NULL DEFAULT '',
    "create_time"     integer NOT NULL DEFAULT 0,
    "update_time"     integer NOT NULL DEFAULT 0,
    PRIMARY KEY ("id")
);

CREATE TABLE "tbl_smart_link_process"
(
    "id"          INTEGER NOT NULL,
    "name"        TEXT    NOT NULL DEFAULT '',
    "status"      integer NOT NULL DEFAULT 1,
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0,
    PRIMARY KEY ("id")
);

CREATE TABLE "tbl_smart_link_process_item"
(
    "id"                    INTEGER NOT NULL,
    "name"                  TEXT    NOT NULL DEFAULT '',
    "smart_link_process_id" INTEGER NOT NULL,
    "type"                  TEXT    NOT NULL DEFAULT '',
    "locator"               TEXT    NOT NULL DEFAULT '',
    "tip"                   TEXT    NOT NULL DEFAULT '',
    "value"                 TEXT    NOT NULL DEFAULT '',
    "out_key"               TEXT    NOT NULL DEFAULT '',
    "check_key"             TEXT    NOT NULL DEFAULT '',
    "weight"                integer NOT NULL DEFAULT 0,
    "domain_limit"          TEXT    NOT NULL DEFAULT '',
    "status"                integer NOT NULL DEFAULT 1,
    "create_time"           integer NOT NULL DEFAULT 0,
    "update_time"           integer NOT NULL DEFAULT 0,
    "wait_mills"            integer NOT NULL DEFAULT 500,
    "append_to_replace"     text    NOT NULL DEFAULT 0,
    "is_async"              TEXT    NOT NULL DEFAULT 0,
    "is_error_continue"     TEXT    NOT NULL DEFAULT '0',
    "next_ids"              text    NOT NULL DEFAULT 0,
    "x"                     integer          DEFAULT 0,
    "y"                     integer          DEFAULT 0,
    PRIMARY KEY ("id")
);

CREATE TABLE "tbl_ssh"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT    NOT NULL DEFAULT '',
    "host"        TEXT    NOT NULL DEFAULT '',
    "port"        TEXT    NOT NULL DEFAULT '',
    "username"    TEXT    NOT NULL DEFAULT '',
    "password"    TEXT    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_star"
(
    "id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT    NOT NULL DEFAULT '',
    "key"         TEXT    NOT NULL,
    "value"       TEXT    NOT NULL DEFAULT '',
    "type"        TEXT    NOT NULL,
    "create_time" integer NOT NULL,
    "update_time" integer NOT NULL
);

CREATE TABLE "tbl_supervisor"
(
    "id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT    NOT NULL,
    "ssh_id"      integer NOT NULL DEFAULT '',
    "docker_name" text    NOT NULL DEFAULT '',
    "config_dir"  TEXT    NOT NULL,
    "create_time" integer NOT NULL,
    "update_time" integer NOT NULL
);

CREATE TABLE "tbl_user"
(
    "id"       INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "username" TEXT    NOT NULL DEFAULT '',
    "password" TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE "tbl_variable"
(
    "id"                INTEGER NOT NULL,
    "name"              TEXT    NOT NULL DEFAULT '',
    "variable_group_id" INTEGER NOT NULL DEFAULT 0,
    "config"            TEXT    NOT NULL DEFAULT '',
    "create_time"       integer NOT NULL DEFAULT 0,
    "update_time"       integer NOT NULL DEFAULT 0,
    "type"              integer NOT NULL DEFAULT 1,
    "key"               TEXT    NOT NULL DEFAULT '',
    "remark"            TEXT    NOT NULL DEFAULT '',
    "status"            integer NOT NULL DEFAULT 1,
    "desc"              TEXT    NOT NULL DEFAULT '',
    PRIMARY KEY ("id")
);

CREATE TABLE "tbl_variable_cmd"
(
    "id"               INTEGER NOT NULL,
    "name"             TEXT    NOT NULL DEFAULT '',
    "variable_id"      INTEGER NOT NULL DEFAULT 0,
    "config"           TEXT    NOT NULL DEFAULT '',
    "type"             integer NOT NULL DEFAULT 0,
    "key"              TEXT    NOT NULL DEFAULT '',
    "remark"           TEXT    NOT NULL DEFAULT '',
    "weight"           integer NOT NULL DEFAULT 0,
    "sql"              TEXT    NOT NULL DEFAULT '',
    "cmd"              TEXT    NOT NULL DEFAULT '',
    "bash"             TEXT    NOT NULL DEFAULT '',
    "result_key"       TEXT    NOT NULL DEFAULT '',
    "options"          TEXT    NOT NULL DEFAULT '',
    "is_pre"           text             DEFAULT '',
    "status"           integer          DEFAULT 1,
    "default"          TEXT    NOT NULL DEFAULT '',
    "smart_link_id"    INTEGER NOT NULL DEFAULT 0,
    "smart_link_label" TEXT    NOT NULL DEFAULT '',
    "create_time"      integer NOT NULL DEFAULT 0,
    "update_time"      integer NOT NULL DEFAULT 0,
    "checks"           TEXT    NOT NULL DEFAULT '',
    "run_type"         text    NOT NULL,
    "next_ids"         INTEGER NOT NULL DEFAULT 0,
    "is_start"         integer NOT NULL DEFAULT 0,
    PRIMARY KEY ("id")
);

-- ----------------------------
-- Indexes structure for table tbl_smart_link_last
-- ----------------------------
CREATE UNIQUE INDEX "idx"
    ON "tbl_smart_link_last" (
                              "domain" ASC,
                              "user_data_index" ASC
        );

