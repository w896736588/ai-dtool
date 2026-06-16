-- 创建新表 smart_link，将老表 tbl_smart_link 中 links JSON 展开为独立行
-- 将老表的 name 作为分组名自动创建 tbl_group
-- Create new smart_link table with links JSON fields flattened into top-level columns

CREATE TABLE "smart_link"
(
    "id"                    INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "label"                 TEXT    NOT NULL DEFAULT '',
    "link"                  TEXT    NOT NULL DEFAULT '',
    "smart_link_group_id"   INTEGER NOT NULL DEFAULT 0,
    "account_list"          TEXT    NOT NULL DEFAULT '',
    "browser_auth_username" TEXT    NOT NULL DEFAULT '',
    "browser_auth_password" TEXT    NOT NULL DEFAULT '',
    "cookie"                TEXT    NOT NULL DEFAULT '',
    "headers"               TEXT    NOT NULL DEFAULT '',
    "open_num"              INTEGER,
    "open_type"             TEXT,
    "process"               TEXT,
    "weight"                integer,
    "combine_type"          integer,
    "status"                integer          DEFAULT 1,
    "value"                 TEXT,
    "create_time"           integer,
    "update_time"           integer,
    "download_finds"        TEXT             DEFAULT '',
    "auto_close_second"     integer          DEFAULT 0,
    "channel"               TEXT             DEFAULT '',
    "show_cookies"          TEXT             DEFAULT '',
    "process_id"            INTEGER NOT NULL DEFAULT 0,
    "filter_uris"           TEXT    NOT NULL DEFAULT ''
);

-- 将老表 name 作为分组名创建 tbl_group
-- Create groups from old table names
INSERT OR IGNORE INTO tbl_group (name, type, create_time, update_time)
SELECT DISTINCT old.name, 4, MIN(old.create_time), MIN(old.update_time)
FROM tbl_smart_link old
WHERE old.status = 1
  AND old.name IS NOT NULL
  AND old.name != ''
GROUP BY old.name;

-- 从老表 tbl_smart_link 展开 links JSON 迁移数据
-- Migrate data by expanding links JSON, join with newly created groups
INSERT INTO smart_link (
    label, link, smart_link_group_id, account_list,
    browser_auth_username, browser_auth_password, cookie, headers,
    open_num, open_type, process, weight, combine_type, status, value,
    create_time, update_time, download_finds, auto_close_second, channel,
    show_cookies, process_id, filter_uris
)
SELECT
    json_extract(link.value, '$.label'),
    json_extract(link.value, '$.link'),
    COALESCE(
        (SELECT g.id FROM tbl_group g WHERE g.name = old.name AND g.type = 4 LIMIT 1),
        old.smart_link_group_id,
        0
    ),
    json_extract(link.value, '$.account_list'),
    COALESCE(json_extract(link.value, '$.browser_auth_username'), ''),
    COALESCE(json_extract(link.value, '$.browser_auth_password'), ''),
    COALESCE(json_extract(link.value, '$.cookie'), ''),
    COALESCE(json_extract(link.value, '$.headers'), ''),
    old.open_num,
    old.open_type,
    old.process,
    old.weight,
    old.combine_type,
    old.status,
    old.value,
    old.create_time,
    old.update_time,
    old.download_finds,
    old.auto_close_second,
    old.channel,
    old.show_cookies,
    COALESCE(NULLIF(json_extract(link.value, '$.process_id'), 0), old.process_id, 0),
    old.filter_uris
FROM tbl_smart_link AS old,
     json_each(old.links) AS link
WHERE old.status = 1
  AND json_extract(link.value, '$.label') IS NOT NULL
  AND json_extract(link.value, '$.label') != ''
  AND NOT EXISTS (
      SELECT 1 FROM smart_link existing
      WHERE existing.link = json_extract(link.value, '$.link')
        AND existing.label = json_extract(link.value, '$.label')
        AND existing.status = 1
  );
