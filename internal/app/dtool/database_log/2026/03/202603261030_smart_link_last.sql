CREATE TABLE IF NOT EXISTS "tbl_smart_link_last"
(
    "id"              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "smart_link_id"   INTEGER NOT NULL DEFAULT 0,
    "user_name"       TEXT    NOT NULL DEFAULT '',
    "user_data_index" INTEGER NOT NULL DEFAULT 0,
    "domain"          TEXT    NOT NULL DEFAULT '',
    "create_time"     INTEGER NOT NULL DEFAULT 0,
    "update_time"     INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS "idx_tbl_smart_link_last_domain_user_data_index"
    ON "tbl_smart_link_last" (
        "domain" ASC,
        "user_data_index" ASC
    );

CREATE INDEX IF NOT EXISTS "idx_tbl_smart_link_last_user_domain"
    ON "tbl_smart_link_last" (
        "user_name" ASC,
        "domain" ASC
    );
