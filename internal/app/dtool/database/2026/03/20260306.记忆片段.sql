CREATE TABLE "tbl_memory_fragment"
(
    "id"            INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "title"         TEXT    NOT NULL DEFAULT '',
    "content"       TEXT    NOT NULL DEFAULT '',
    "content_text"  TEXT    NOT NULL DEFAULT '',
    "is_deleted"    INTEGER NOT NULL DEFAULT 0,
    "index_status"  TEXT    NOT NULL DEFAULT 'pending',
    "index_version" INTEGER NOT NULL DEFAULT 1,
    "create_time"   INTEGER NOT NULL DEFAULT 0,
    "update_time"   INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_memory_fragment_tag"
(
    "id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "fragment_id" INTEGER NOT NULL DEFAULT 0,
    "tag_name"    TEXT    NOT NULL DEFAULT '',
    "create_time" INTEGER NOT NULL DEFAULT 0,
    "update_time" INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_memory_fragment_history"
(
    "id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "fragment_id" INTEGER NOT NULL DEFAULT 0,
    "title_old"   TEXT    NOT NULL DEFAULT '',
    "title_new"   TEXT    NOT NULL DEFAULT '',
    "content_old" TEXT    NOT NULL DEFAULT '',
    "content_new" TEXT    NOT NULL DEFAULT '',
    "tags_old"    TEXT    NOT NULL DEFAULT '[]',
    "tags_new"    TEXT    NOT NULL DEFAULT '[]',
    "change_desc" TEXT    NOT NULL DEFAULT '',
    "create_time" INTEGER NOT NULL DEFAULT 0,
    "update_time" INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_memory_fragment_fts"
(
    "id"           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "fragment_id"  INTEGER NOT NULL DEFAULT 0,
    "title"        TEXT    NOT NULL DEFAULT '',
    "content_text" TEXT    NOT NULL DEFAULT '',
    "tag_text"     TEXT    NOT NULL DEFAULT '',
    "search_text"  TEXT    NOT NULL DEFAULT '',
    "create_time"  INTEGER NOT NULL DEFAULT 0,
    "update_time"  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE "tbl_memory_fragment_vec"
(
    "id"            INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "fragment_id"   INTEGER NOT NULL DEFAULT 0,
    "embedding"     TEXT    NOT NULL DEFAULT '',
    "model_name"    TEXT    NOT NULL DEFAULT '',
    "model_version" TEXT    NOT NULL DEFAULT '',
    "create_time"   INTEGER NOT NULL DEFAULT 0,
    "update_time"   INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX "idx_memory_fragment_deleted_update"
    ON "tbl_memory_fragment" (
                              "is_deleted" ASC,
                              "update_time" DESC
        );

CREATE INDEX "idx_memory_fragment_tag_fragment"
    ON "tbl_memory_fragment_tag" (
                                  "fragment_id" ASC
        );

CREATE INDEX "idx_memory_fragment_tag_name"
    ON "tbl_memory_fragment_tag" (
                                  "tag_name" ASC
        );

CREATE UNIQUE INDEX "idx_memory_fragment_tag_unique"
    ON "tbl_memory_fragment_tag" (
                                  "fragment_id" ASC,
                                  "tag_name" ASC
        );

CREATE INDEX "idx_memory_fragment_history_fragment"
    ON "tbl_memory_fragment_history" (
                                      "fragment_id" ASC,
                                      "id" DESC
        );

CREATE UNIQUE INDEX "idx_memory_fragment_fts_fragment"
    ON "tbl_memory_fragment_fts" (
                                  "fragment_id" ASC
        );

CREATE UNIQUE INDEX "idx_memory_fragment_vec_fragment"
    ON "tbl_memory_fragment_vec" (
                                  "fragment_id" ASC
        );
