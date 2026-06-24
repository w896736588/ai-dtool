-- 消息表新增 FC 循环中间消息所需字段
-- tool_calls: 存储 assistant 消息中的 tool_calls JSON（仅 FC 中间消息有值）
-- tool_call_id: 存储 tool 结果消息的调用 ID（仅 FC tool 消息有值）
ALTER TABLE "tbl_butler_message" ADD COLUMN "tool_calls" TEXT NOT NULL DEFAULT '';
ALTER TABLE "tbl_butler_message" ADD COLUMN "tool_call_id" TEXT NOT NULL DEFAULT '';
