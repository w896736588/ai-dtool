CREATE TABLE "tbl_api"
(
    "id"                integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"              text    NOT NULL DEFAULT '',
    "folder_id"         integer NOT NULL DEFAULT 0,
    "collection_id"     integer NOT NULL DEFAULT 0,
    "method" text       NOT NULL DEFAULT '',
    "url" text          NOT NULL DEFAULT '',
    "protocol" text     NOT NULL DEFAULT '',
    "desc" text         NOT NULL DEFAULT '',
    "headers" text      NOT NULL DEFAULT '',
    "query_params" text  NOT NULL DEFAULT '',
    "content_type" text NOT NULL DEFAULT '',
    "body_form" text    NOT NULL DEFAULT '',
    "body_json" text    NOT NULL DEFAULT '',
    "create_time"       integer NOT NULL DEFAULT 0,
    "update_time"       integer NOT NULL DEFAULT 0
);