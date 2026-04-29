-- 为 tbl_smart_link_client 添加 agent_token 列
-- Agent WebSocket 连接鉴权使用
ALTER TABLE "tbl_smart_link_client" ADD COLUMN "agent_token" text NOT NULL DEFAULT '';
