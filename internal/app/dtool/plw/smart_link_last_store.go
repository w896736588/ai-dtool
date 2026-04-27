package plw

import (
	"dev_tool/internal/app/dtool/common"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// SmartLinkLastStore hides where smart-link user-data history is persisted.
// Server mode uses sqlite; agent mode injects an HTTP implementation.
type SmartLinkLastStore interface {
	GetLastUserDataIndex(userName, domain string) (int, error)
	ExistDomainUserDataIndex(domain string, userDataIndex int) (bool, error)
	UpsertLastUserDataIndex(smartLinkID int, userName, domain string, userDataIndex int) error
}

type dbSmartLinkLastStore struct{}

// NewDBSmartLinkLastStore 创建服务端默认的 sqlite 历史目录存储实现。
func NewDBSmartLinkLastStore() SmartLinkLastStore {
	return dbSmartLinkLastStore{}
}

// GetLastUserDataIndex 从 log 库读取指定用户和域名上次使用的目录索引。
func (dbSmartLinkLastStore) GetLastUserDataIndex(userName, domain string) (int, error) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		return 0, nil
	}
	sql := `select * from tbl_smart_link_last where user_name = ? and domain = ? `
	row, err := common.DbLog.Client.QueryBySql(sql, userName, domain).One()
	if err != nil {
		return 0, err
	}
	return cast.ToInt(row[`user_data_index`]), nil
}

// ExistDomainUserDataIndex 从 log 库判断指定域名是否已经占用某个目录索引。
func (dbSmartLinkLastStore) ExistDomainUserDataIndex(domain string, userDataIndex int) (bool, error) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		return false, nil
	}
	sql := `select * from tbl_smart_link_last where domain = ? and user_data_index = ? `
	row, err := common.DbLog.Client.QueryBySql(sql, domain, userDataIndex).One()
	if err != nil {
		return false, err
	}
	return len(row) > 0, nil
}

// UpsertLastUserDataIndex 在 log 库记录本次任务实际使用的目录索引。
func (dbSmartLinkLastStore) UpsertLastUserDataIndex(smartLinkID int, userName, domain string, userDataIndex int) error {
	if common.DbLog == nil || common.DbLog.Client == nil {
		return nil
	}
	sql := `select * from tbl_smart_link_last where  smart_link_id = ? and user_name = ? and domain = ?`
	row, err := common.DbLog.Client.QueryBySql(sql, smartLinkID, userName, domain).One()
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	if len(row) > 0 {
		_, err = common.DbLog.Client.QuickUpdate(`tbl_smart_link_last`, map[string]any{
			`smart_link_id`: smartLinkID,
			`user_name`:     userName,
			`domain`:        domain,
		}, map[string]any{
			`user_data_index`: userDataIndex,
			`update_time`:     now,
		}).Exec()
		return err
	}
	_, err = common.DbLog.Client.QuickCreate(`tbl_smart_link_last`, map[string]any{
		`smart_link_id`:   smartLinkID,
		`user_name`:       userName,
		`user_data_index`: userDataIndex,
		`domain`:          domain,
		`create_time`:     now,
		`update_time`:     now,
	}).Exec()
	if err != nil && strings.Contains(err.Error(), `UNIQUE constraint failed: tbl_smart_link_last.domain, tbl_smart_link_last.user_data_index`) {
		// 并发或历史脏数据导致占用键冲突时，保留唯一占用关系并更新归属信息。
		_, err = common.DbLog.Client.QuickUpdate(`tbl_smart_link_last`, map[string]any{
			`domain`:          domain,
			`user_data_index`: userDataIndex,
		}, map[string]any{
			`smart_link_id`: smartLinkID,
			`user_name`:     userName,
			`update_time`:   now,
		}).Exec()
	}
	return err
}
