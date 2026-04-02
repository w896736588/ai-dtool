package common

import "dev_tool/internal/pkg/p_common"

func GetCall() *p_common.Call {
	return &p_common.Call{
		AllGlobal: func() ([]map[string]any, error) {
			return DbMain.AllGlobal()
		},
		GetSshConfig: func(s any) (map[string]any, error) {
			return DbMain.GetSshConfig(s)
		},
		GetAllSshConfig: func() ([]map[string]any, error) {
			return DbMain.GetAllSshConfig()
		},
		QueryGroupInfo: func(m map[string]any) (map[string]any, error) {
			return DbMain.Client.QuickQuery(`tbl_group`, `*`, m).One()
		},
		QueryAccountList: func(m map[string]any) ([]map[string]any, error) {
			return DbMain.Client.QuickQuery(`tbl_account`, `*`, m).All()
		},
		QueryMysqlConfig: func(s int) (map[string]any, error) {
			return DbMain.GetMysqlConfig(s)
		},
		QueryGlobalConfig: func(m map[string]any) (map[string]any, error) {
			return DbMain.Client.QuickQuery(`tbl_global`, `*`, m).One()
		},
		GetRedisConfig: func(s any) (map[string]any, error) {
			return DbMain.GetRedisConfig(s)
		},
		CmdInfo: func(s any) (map[string]any, error) {
			return DbMain.CmdInfo(s)
		},
		CreateSmartLastRecord: func(d map[string]any) (int64, error) {
			return DbLog.Client.QuickCreate(`tbl_smart_link_last`, d).Exec()
		},
		UpdateSmartLastRecord: func(m map[string]any, d map[string]any) (int64, error) {
			return DbLog.Client.QuickUpdate(`tbl_smart_link_last`, m, d).Exec()
		},
	}
}
