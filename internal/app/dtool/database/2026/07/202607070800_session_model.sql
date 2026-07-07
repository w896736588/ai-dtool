-- 为 tbl_agent_v2_session 添加 model_name 字段，支持每个会话记住最后使用的模型

ALTER TABLE tbl_agent_v2_session ADD COLUMN model_name TEXT NOT NULL DEFAULT '';
