CREATE TABLE IF NOT EXISTS "tbl_task_status"
(
    "id"          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"        TEXT    NOT NULL DEFAULT '',
    "sort_order"  INTEGER NOT NULL DEFAULT 0,
    "create_time" INTEGER NOT NULL DEFAULT 0,
    "update_time" INTEGER NOT NULL DEFAULT 0
);

-- 插入默认状态（按常见开发流程排序）
INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('待开始', 0, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('开发中', 1, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('开发完', 2, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('自测中', 3, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('自测完', 4, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('待对接', 5, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('对接中', 6, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('测试中', 7, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('待测试', 8, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('上线中', 9, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('已上线', 10, strftime('%s', 'now'), strftime('%s', 'now'));

INSERT INTO "tbl_task_status" ("name", "sort_order", "create_time", "update_time")
VALUES ('已废弃', 11, strftime('%s', 'now'), strftime('%s', 'now'));
