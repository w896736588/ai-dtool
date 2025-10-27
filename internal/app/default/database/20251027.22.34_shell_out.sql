CREATE TABLE "tbl_shell_out"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "command"     text    NOT NULL DEFAULT '',
    "name"        text    NOT NULL DEFAULT '',
    "desc"        text    NOT NULL DEFAULT '',
    "ssh_id"      integer NOT NULL DEFAULT 0,
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);