package store

import (
	"dev_tool/internal/app/dtool/common"
	"time"
)

// UpdateSmartLink 绑定 smart_link 登录链路 + 一次性 ws_token。
// 主键遵循项目约定：tbl_e2e_record_session 自增主键为 row_id。
func (s *RecordSessionStore) UpdateSmartLink(id int64, smartLinkID int, userName, wsToken, recorderURL string, linkID int) error {
	_, err := common.DbMain.Client.ExecBySql(`
		UPDATE tbl_e2e_record_session
		SET smart_link_id = ?, user_name = ?, ws_token = ?, recorder_url = ?, link_id = ?, updated_at = ?
		WHERE row_id = ?`,
		smartLinkID, userName, wsToken, recorderURL, linkID, time.Now().Unix(), id,
	).Exec()
	return err
}

// FindByToken 按 ws_token 查询会话。
// 空 token 直接返回 (nil, nil)，避免任何 SQL 副作用。
func (s *RecordSessionStore) FindByToken(token string) (map[string]any, error) {
	if token == "" {
		return nil, nil
	}
	return common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_e2e_record_session WHERE ws_token = ? LIMIT 1`, token,
	).One()
}

// UpdateWSToken 重置一次性 token（续录用）。
func (s *RecordSessionStore) UpdateWSToken(id int64, newToken string) error {
	_, err := common.DbMain.Client.ExecBySql(
		`UPDATE tbl_e2e_record_session SET ws_token = ?, updated_at = ? WHERE row_id = ?`,
		newToken, time.Now().Unix(), id,
	).Exec()
	return err
}

// MarkPaused 标记会话为暂停。
func (s *RecordSessionStore) MarkPaused(id int64) error {
	_, err := common.DbMain.Client.ExecBySql(
		`UPDATE tbl_e2e_record_session SET status = 'paused', updated_at = ? WHERE row_id = ?`,
		time.Now().Unix(), id,
	).Exec()
	return err
}

// MarkRecording 标记会话为录制中。
func (s *RecordSessionStore) MarkRecording(id int64) error {
	_, err := common.DbMain.Client.ExecBySql(
		`UPDATE tbl_e2e_record_session SET status = 'recording', updated_at = ? WHERE row_id = ?`,
		time.Now().Unix(), id,
	).Exec()
	return err
}