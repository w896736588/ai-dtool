-- E2E 录制会话与 smart_link 绑定（v6）
-- 在 v5 基础上新增：smart_link_id / user_name / ws_token / link_id / recorder_url
-- 兼容策略：默认 0/空，老 record_session 不参与 /api/e2e/record/by_token/* 鉴权。
-- 主键遵循项目约定：tbl_e2e_record_session 的自增主键为 row_id。
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "smart_link_id" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "user_name" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "ws_token" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "link_id" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "tbl_e2e_record_session" ADD COLUMN "recorder_url" TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_token"
    ON "tbl_e2e_record_session" ("ws_token");
CREATE INDEX IF NOT EXISTS "idx_tbl_e2e_record_session_smart_link"
    ON "tbl_e2e_record_session" ("smart_link_id");