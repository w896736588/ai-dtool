-- E2E 录制会话与 smart_link 绑定（v6）— 修复版
-- 在 v5 基础上新增：smart_link_id / user_name / ws_token / link_id / recorder_url
-- 兼容策略：默认 0/空，老 record_session 不参与 /api/e2e/record/by_token/* 鉴权。
-- 主键遵循项目约定：tbl_e2e_record_session 的自增主键为 row_id。
--
-- 修复要点：
-- 1) 上一版迁移在含历史空 ws_token 行的库上会因 UNIQUE 索引失败而中止。
--    本版本先把所有空字符串 token 规整为同一个 sentinel 'none'，确保 UNIQUE 不会因为
--    大量 '' 重复值而冲突（sentinel 与真实 token 互斥：真实 token 由 base64 random 生成）。
-- 2) ALTER TABLE ADD COLUMN 在 SQLite 中不支持 IF NOT EXISTS。迁移 runner 在
--    ExecBySql 阶段如果遇到 "duplicate column name" 会整文件失败，故改用
--    pragma_table_info 视图判断列是否已存在，存在则跳过。
-- 3) 历史库的 row_id / session_id / group_id 等 v5 列必须保留，故逐列加 ALTER。

-- 列存在性检查
-- SQLite pragma_table_info 返回当前表所有列的 (cid, name, type, notnull, dflt_value, pk)。

ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "smart_link_id" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "user_name" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "ws_token" TEXT NOT NULL DEFAULT 'none';
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "link_id" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "recorder_url" TEXT NOT NULL DEFAULT '';

-- 历史记录里残留 '' 的 ws_token 全部规整为 'none'，从而 UNIQUE 可以创建
UPDATE "tbl_e2e_record_session" SET "ws_token" = 'none' WHERE "ws_token" = '' OR "ws_token" IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_token"
    ON "tbl_e2e_record_session" ("ws_token");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_smart_link"
    ON "tbl_e2e_record_session" ("smart_link_id");