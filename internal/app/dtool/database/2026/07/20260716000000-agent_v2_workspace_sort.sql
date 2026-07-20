-- Agent V2 工作空间排序：新增 sort_order 字段，支持按对话置顶/拖动顺序持久化排序
ALTER TABLE tbl_agent_v2_workspace ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;
