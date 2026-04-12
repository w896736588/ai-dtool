CREATE TABLE "tbl_shell_out_rule_set"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        text    NOT NULL DEFAULT '',
    "description" text    NOT NULL DEFAULT '',
    "is_enabled"  integer NOT NULL DEFAULT 1,
    "match_mode"  text    NOT NULL DEFAULT 'line',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_shell_out_rule_item"
(
    "id"              integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "rule_set_id"     integer NOT NULL DEFAULT 0,
    "name"            text    NOT NULL DEFAULT '',
    "rule_type"       text    NOT NULL DEFAULT '',
    "match_type"      text    NOT NULL DEFAULT '',
    "pattern"         text    NOT NULL DEFAULT '',
    "exclude_pattern" text    NOT NULL DEFAULT '',
    "priority"        integer NOT NULL DEFAULT 0,
    "is_enabled"      integer NOT NULL DEFAULT 1,
    "stop_on_match"   integer NOT NULL DEFAULT 0,
    "config_json"     text    NOT NULL DEFAULT '{}',
    "create_time"     integer NOT NULL DEFAULT 0,
    "update_time"     integer NOT NULL DEFAULT 0
);

ALTER TABLE "tbl_shell_out"
    ADD COLUMN "rule_set_id" integer NOT NULL DEFAULT 0;
