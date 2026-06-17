package butler

import (
	"dev_tool/internal/app/dtool-butler/define"
	"dev_tool/internal/app/dtool/common"
	"time"

	"github.com/spf13/cast"
)

// History 历史对话存储，读写 tbl_butler_message。
type History struct {
	db *common.CSqlite
}

// NewHistory 创建历史存储实例。
func NewHistory(db *common.CSqlite) *History {
	return &History{db: db}
}

// Append 追加一条消息记录。
func (h *History) Append(sessionId, role, content string) error {
	_, err := h.db.Client.QuickCreate(`tbl_butler_message`, map[string]any{
		`session_id`:  sessionId,
		`role`:        role,
		`content`:     content,
		`token_count`: 0,
		`topic`:       ``,
		`created_at`:  time.Now().Unix(),
	}).Exec()
	return err
}

// CountBySession 返回指定会话的消息条数。
func (h *History) CountBySession(sessionId string) (int, error) {
	rows, err := h.db.Client.QuickQuery(`tbl_butler_message`, `id`, map[string]any{
		`session_id`: sessionId,
	}).All()
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}

// CleanBySession 清除指定会话的全部历史消息。
func (h *History) CleanBySession(sessionId string) error {
	_, err := h.db.Client.ExecBySql(
		`DELETE FROM tbl_butler_message WHERE session_id = ?`, sessionId,
	).Exec()
	return err
}

// ListBySession 返回指定会话的历史消息（按 id 升序），最多 limit 条。
func (h *History) ListBySession(sessionId string, limit int) ([]define.HistoryMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := h.db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_message WHERE session_id = ? ORDER BY id ASC LIMIT ?`,
		sessionId, limit,
	).All()
	if err != nil {
		return nil, err
	}
	result := make([]define.HistoryMessage, 0, len(rows))
	for _, row := range rows {
		result = append(result, define.HistoryMessage{
			Id:        cast.ToInt(row[`id`]),
			SessionId: cast.ToString(row[`session_id`]),
			Role:      cast.ToString(row[`role`]),
			Content:   cast.ToString(row[`content`]),
			Topic:     cast.ToString(row[`topic`]),
			CreatedAt: cast.ToInt64(row[`created_at`]),
		})
	}
	return result, nil
}
