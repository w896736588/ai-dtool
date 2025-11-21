CREATE TABLE "tbl_api_collection"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        text    NOT NULL DEFAULT '',
    "desc"        text    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);