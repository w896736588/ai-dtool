CREATE TABLE "tbl_api_dir"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        text    NOT NULL DEFAULT '',
    "collection_id" integer NOT NULL DEFAULT 0,
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);