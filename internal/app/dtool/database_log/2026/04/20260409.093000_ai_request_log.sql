-- AI 请求日志表 / AI request log table
CREATE TABLE IF NOT EXISTS "tbl_ai_request_log"
(
    "id"                    INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    -- 服务商ID
    "provider_id"           INTEGER NOT NULL DEFAULT 0,
    -- 服务商名称
    "provider_name"         TEXT    NOT NULL DEFAULT '',
    -- 模型ID
    "model_id"              INTEGER NOT NULL DEFAULT 0,
    -- 模型名称（展示名）
    "model_name"            TEXT    NOT NULL DEFAULT '',
    -- 模型标识（如 gpt-4o-mini）
    "model"                 TEXT    NOT NULL DEFAULT '',
    -- 模型类型：llm 或 embedding
    "model_type"            TEXT    NOT NULL DEFAULT 'llm',
    -- 请求格式：openai
    "request_format"        TEXT    NOT NULL DEFAULT 'openai',
    -- 基础域名
    "base_url"              TEXT    NOT NULL DEFAULT '',
    -- 完整请求地址
    "request_url"           TEXT    NOT NULL DEFAULT '',
    -- 请求方法
    "request_method"        TEXT    NOT NULL DEFAULT 'POST',
    -- 请求参数（JSON字符串）
    "request_params"        TEXT    NOT NULL DEFAULT '',
    -- 请求头（JSON字符串）
    "request_headers"       TEXT    NOT NULL DEFAULT '',
    -- 响应状态码
    "response_status_code"   INTEGER NOT NULL DEFAULT 0,
    -- 响应内容（JSON字符串）
    "response_body"         TEXT    NOT NULL DEFAULT '',
    -- 输入 token 数量
    "input_tokens"          INTEGER NOT NULL DEFAULT 0,
    -- 输出 token 数量
    "output_tokens"         INTEGER NOT NULL DEFAULT 0,
    -- 耗时（毫秒）
    "cost_time_ms"          INTEGER NOT NULL DEFAULT 0,
    -- 是否成功
    "success"               INTEGER NOT NULL DEFAULT 1,
    -- 错误信息
    "error_message"         TEXT    NOT NULL DEFAULT '',
    -- 创建时间
    "create_time"           INTEGER NOT NULL DEFAULT 0
);

-- 创建索引提升查询效率
CREATE INDEX IF NOT EXISTS "idx_ai_request_log_create_time" ON "tbl_ai_request_log" ("create_time" DESC);
CREATE INDEX IF NOT EXISTS "idx_ai_request_log_provider_id" ON "tbl_ai_request_log" ("provider_id");
CREATE INDEX IF NOT EXISTS "idx_ai_request_log_model_id" ON "tbl_ai_request_log" ("model_id");