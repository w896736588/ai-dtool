package service

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"time"
)

// QueryWechatAppid 查询应用信息
func QueryWechatAppid(mysqlCli *gsdb.GsMysql, wechatAppId string) map[string]interface{} {
	data, err := mysqlCli.QueryBySql(`select _id,app_id,app_type,user_id,app_name from tbl_wechatapp where (app_id = ? or _id = ?) and user_id > 0`, wechatAppId, cast.ToInt(wechatAppId)).One()
	if err != nil {
		return make(map[string]interface{})
	}
	return data
}

// GetAdminUserId 拿到用户信息
func GetAdminUserId(mysqlCli *gsdb.GsMysql, account string) map[string]interface{} {
	data, err := mysqlCli.QueryBySql(`select _id,user_name from tbl_user where (user_name = ? or _id = ?)`, account, cast.ToInt(account)).One()
	if err != nil {
		gstool.FmtPrintlnLog(`查询出错 %s`, err.Error())
		return make(map[string]interface{})
	}
	return data
}

// QueryEnvWechatKefuList 查询微信客服
func QueryEnvWechatKefuList(mysqlCli *gsdb.GsMysql, adminUserId string) []map[string]interface{} {
	dataList, err := mysqlCli.QueryBySql(`select app_name,app_id,app_type from tbl_wechatapp where user_id = ? and app_type = ?`, adminUserId, `wechat_kefu`).All()
	if err != nil {
		return []map[string]interface{}{}
	}
	return dataList
}

// UpdateVip 变更vip版本
func UpdateVip(mysqlCli *gsdb.GsMysql, adminUserId, expiredDay, systemType, vipLevel string) (int64, error) {
	vipTable := `tbl_kefu_vip`
	if systemType == `1` { //客服系统
		vipTable = `tbl_kefu_vip`
	} else {
		vipTable = `tbl_official_account_vip`
	}
	//时间
	t := time.Now().Unix()
	t += 86400 * cast.ToInt64(expiredDay)
	expiredTime := time.Unix(t, 0).Format(`2006-01-02 15:04:05`)
	sqlStr := `update ` + vipTable + ` set expired_time = ? , vip_type = ? where user_id =?`
	return mysqlCli.ExecBySql(sqlStr, expiredTime, vipLevel, adminUserId).Exec()
}

// QueryVip 查询VIP信息
func QueryVip(mysqlCli *gsdb.GsMysql, adminUserId, systemType string) (map[string]interface{}, error) {
	vipTable := `tbl_kefu_vip`
	if systemType == `1` { //客服系统
		vipTable = `tbl_kefu_vip`
	} else {
		vipTable = `tbl_official_account_vip`
	}
	return mysqlCli.QueryBySql(`select vip_type,expired_time from `+vipTable+` where user_id = ?`, adminUserId).One()
}

// QueryOneWechatAppIdChannelId 拿一个应用ID和渠道
func QueryOneWechatAppIdChannelId(mysqlCli *gsdb.GsMysql, userId int) (string, string, error) {
	appInfo, err := mysqlCli.QueryBySql(`select wechatapp_id from tbl_staff_wechatapp_relation where user_id = ?`, userId).One()
	fmt.Println(fmt.Sprintf(`%#v`, appInfo))
	if err != nil {
		return ``, ``, err
	}
	channelInfo, err := mysqlCli.QueryBySql(`select channel_id from tbl_channel_user_rel where user_id = ? and wechatapp_id = ? and status = 1`, userId, cast.ToInt(appInfo[`wechatapp_id`])).One()
	if err != nil {
		return ``, ``, err
	}
	return cast.ToString(appInfo[`wechatapp_id`]), cast.ToString(channelInfo[`channel_id`]), nil
}
