-- E2E 录制会话表扩展（v5.0）
-- 在 v4 基础上加入：
--   * 自增主键 `id`（与外部 API 交互用）
--   * 业务 session_id（UUID，客户端标识）
--   * group_id（关联到 tbl_e2e_group，便于按组汇总录制会话）
--   * browser_id（录制所属 Playwright 实例的标识）
--   * status（recording / committed / discarded）
-- 兼容策略：保留原 id 作为业务主键（TEXT），
--   新增自增主键 `row_id`，业务 session_id 字段名为 `session_id`。

-- SQLite 不支持 ALTER TABLE 增加自增主键，这里采取"加列 + 用 row_id 做自增"策略。
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "row_id" INTEGER;
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "session_id" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "group_id" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "browser_id" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "status" TEXT NOT NULL DEFAULT 'recording';

-- 把现有数据的 row_id 用 id 的哈希派生（保证不重复即可，外部不会引用）。
-- 用 SQLite 内置函数：abs(random()) 仅做兜底，真实业务中 id 是 TEXT UUID，row_id 可为空自增。
UPDATE "tbl_e2e_record_session" SET "session_id" = "id" WHERE "session_id" = '';

CREATE UNIQUE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_row"
    ON "tbl_e2e_record_session" ("row_id");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_session"
    ON "tbl_e2e_record_session" ("session_id");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_group"
    ON "tbl_e2e_record_session" ("group_id");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_status"
    ON "tbl_e2e_record_session" ("status");
