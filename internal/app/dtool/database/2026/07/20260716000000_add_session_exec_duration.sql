-- Agent V2 会话执行耗时（毫秒）：后端按轮次累计，工具调用/思考完成等事件实时推送前端，并在每轮结束时落库
ALTER TABLE tbl_agent_v2_session ADD COLUMN exec_duration_ms BIGINT NOT NULL DEFAULT 0;
