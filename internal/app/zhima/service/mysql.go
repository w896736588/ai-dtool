package service

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"github.com/spf13/cast"
	"time"
)

//QueryWechatAppid 查询应用信息
func QueryWechatAppid(mysqlCli *gsdb.GsMysql, wechatAppId string) map[string]interface{} {
	data, err := mysqlCli.GetOne(`select _id,app_id,app_type,user_id,app_name from tbl_wechatapp where (app_id = ? or _id = ?) and user_id > 0`, wechatAppId, cast.ToInt(wechatAppId))
	if err != nil {
		return make(map[string]interface{})
	}
	return data
}

// GetAdminUserId 拿到用户信息
func GetAdminUserId(mysqlCli *gsdb.GsMysql, account string) map[string]interface{} {
	data, err := mysqlCli.GetOne(`select _id,user_name from tbl_user where (user_name = ? or _id = ?)`, account, cast.ToInt(account))
	if err != nil {
		return make(map[string]interface{})
	}
	return data
}

// QueryEnvWechatKefuList 查询微信客服
func QueryEnvWechatKefuList(mysqlCli *gsdb.GsMysql, adminUserId string) []map[string]interface{} {
	dataList, err := mysqlCli.GetAll(`select app_name,app_id,app_type from tbl_wechatapp where user_id = ? and app_type = ?`, adminUserId, `wechat_kefu`)
	if err != nil {
		dataList := make([]map[string]interface{}, 0)
		return dataList
	}
	return dataList
}

// UpdateVip 变更vip版本
func UpdateVip(mysqlCli *gsdb.GsMysql, adminUserId, expiredDay, systemType, vipLevel string) (int, error) {
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
	return mysqlCli.Update(sqlStr, expiredTime, vipLevel, adminUserId)
}

// QueryVip 查询VIP信息
func QueryVip(mysqlCli *gsdb.GsMysql, adminUserId, systemType string) (map[string]interface{}, error) {
	vipTable := `tbl_kefu_vip`
	if systemType == `1` { //客服系统
		vipTable = `tbl_kefu_vip`
	} else {
		vipTable = `tbl_official_account_vip`
	}
	return mysqlCli.GetOne(`select vip_type,expired_time from `+vipTable+` where user_id = ?`, adminUserId)
}

//QueryOneWechatAppIdChannelId 拿一个应用ID和渠道
func QueryOneWechatAppIdChannelId(mysqlCli *gsdb.GsMysql, userId int) (string, string, error) {
	appInfo, err := mysqlCli.GetOne(`select wechatapp_id from tbl_staff_wechatapp_relation where user_id = ?`, userId)
	fmt.Println(fmt.Sprintf(`%#v`, appInfo))
	if err != nil {
		return ``, ``, err
	}
	channelInfo, err := mysqlCli.GetOne(`select channel_id from tbl_channel_user_rel where user_id = ? and wechatapp_id = ? and status = 1`, userId, cast.ToInt(appInfo[`wechatapp_id`]))
	if err != nil {
		return ``, ``, err
	}
	return cast.ToString(appInfo[`wechatapp_id`]), cast.ToString(channelInfo[`channel_id`]), nil
}
