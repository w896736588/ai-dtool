-- 增加归档相关字段：archived 标记是否归档，original_collection_id 记录归档前所属集合
ALTER TABLE tbl_api_dir ADD COLUMN archived integer NOT NULL DEFAULT 0;
ALTER TABLE tbl_api_dir ADD COLUMN original_collection_id integer NOT NULL DEFAULT 0;
