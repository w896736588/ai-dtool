CREATE TABLE IF NOT EXISTS "tbl_memory_config"
(
    "id"          integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    "key"         text    NOT NULL DEFAULT '',
    "value"       text    NOT NULL DEFAULT '',
    "name"        text    NOT NULL DEFAULT '',
    "desc"        text    NOT NULL DEFAULT '',
    "create_time" integer NOT NULL DEFAULT 0,
    "update_time" integer NOT NULL DEFAULT 0
);

-- 从 tbl_global 迁移已有记忆整理配置
INSERT INTO tbl_memory_config (key, value, name, desc, create_time, update_time)
SELECT 'memory_arrange_prompt', COALESCE(value, ''), '记忆整理提示词', '知识片段 AI 整理提示词',
       strftime('%s', 'now'), strftime('%s', 'now')
FROM tbl_global WHERE key = 'memory_arrange_prompt'
WHERE NOT EXISTS (SELECT 1 FROM tbl_memory_config WHERE key = 'memory_arrange_prompt');

INSERT INTO tbl_memory_config (key, value, name, desc, create_time, update_time)
SELECT 'memory_arrange_model_id', COALESCE(value, ''), '记忆整理模型', '知识片段 AI 整理所用模型 id',
       strftime('%s', 'now'), strftime('%s', 'now')
FROM tbl_global WHERE key = 'memory_arrange_model_id'
WHERE NOT EXISTS (SELECT 1 FROM tbl_memory_config WHERE key = 'memory_arrange_model_id');

-- 清理 tbl_global 中已迁移的记忆整理配置
DELETE FROM tbl_global WHERE key = 'memory_arrange_prompt';
DELETE FROM tbl_global WHERE key = 'memory_arrange_model_id';
