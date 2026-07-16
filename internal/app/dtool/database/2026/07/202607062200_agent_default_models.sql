-- 为 tbl_ai_model 添加上下文窗口大小字段
ALTER TABLE tbl_ai_model ADD COLUMN context_size INTEGER NOT NULL DEFAULT 128000;
