-- E2E 录制会话 ws_token 列：删除 UNIQUE 索引。
--
-- 背景：v7 方案下线了 ws_token HTTP 通道，但表上仍然保留 ws_token 列与
-- UNIQUE INDEX idx_tbl_e2e_record_session_token，每次新建 record_session 写入
-- sentinel 都会触发 UNIQUE 冲突。直接保留 UNIQUE 没有业务意义，反而制造障碍。
--
-- 修复策略：删除 UNIQUE INDEX（保留 ws_token 列与普通 INDEX），让 v7 可以安心
-- 把 ws_token 写为空字符串或固定 sentinel。
DROP INDEX IF EXISTS "idx_tbl_e2e_record_session_token";
