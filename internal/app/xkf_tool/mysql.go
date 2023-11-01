package xkf_tool

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"time"
)

// QueryWechatAppid 拿到应用的AppId https://www.cnblogs.com/jasonminghao/p/12386580.html
// @auth frog
// @date 2023-01-17 12:25:31
func QueryWechatAppid(wechatAppId string) TblWechatapp {
	var appInfo = TblWechatapp{}
	//goland:noinspection ALL
	data, err := XkfDevMysql.GetOne(`select _id,app_id,app_type,user_id,app_name from tbl_wechatapp where (app_id = ? or _id = ?) and user_id > 0`, wechatAppId, cast.ToInt(wechatAppId))
	if err != nil {
		Logger.Errorf(`执行sql出错 %s`, err.Error())
		return appInfo
	}
	appInfo.AppType = cast.ToString(data[`app_type`])
	appInfo.Appid = cast.ToString(data[`app_id`])
	appInfo.AppName = cast.ToString(data[`app_name`])
	appInfo.UserId = cast.ToString(data[`user_id`])
	appInfo.Id = cast.ToString(data[`_id`])
	return appInfo
}

// GetAdminUserId 拿到用户信息
// @auth frog
// @date 2023-01-17 15:49:51
func GetAdminUserId(account string) TblUser {
	var userInfo = TblUser{}
	//goland:noinspection ALL
	data, err := AppurlDevMysql.GetOne(`select _id,user_name from tbl_user where (user_name = ? or _id = ?)`, account, cast.ToInt(account))
	if err != nil {
		log.Errorf(`执行sql出错 %s`, err.Error())
		return userInfo
	}
	userInfo.Id = cast.ToString(data[`_id`])
	userInfo.Username = cast.ToString(data[`user_name`])
	return userInfo
}

// UpdateVip 变更vip版本
// @auth frog
// @date 2023-01-17 16:03:11
func UpdateVip(adminUserId, expiredDay, systemType, vipLevel string) string {
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
	_, err := XkfDevMysql.Update(sqlStr, expiredTime, vipLevel, adminUserId)
	if err != nil {
		log.Errorf(`更新%s失败 %s`, vipTable, err.Error())
		return `更新失败`
	}
	return `成功`
}

// QueryVip 查询VIP信息
// @auth frog
// @date 2023-03-16 09:31:34
func QueryVip(adminUserId, systemType string) *TblVip {
	var vipInfo = &TblVip{}
	vipTable := `tbl_kefu_vip`
	if systemType == `1` { //客服系统
		vipTable = `tbl_kefu_vip`
	} else {
		vipTable = `tbl_official_account_vip`
	}
	//goland:noinspection SqlNoDataSourceInspection
	data, err := XkfDevMysql.GetOne(`select vip_type,expired_time from `+vipTable+` where user_id = ?`, adminUserId)
	if err != nil {
		log.Errorf(`执行sql出错 %s`, err.Error())
		return vipInfo
	}
	vipInfo.VipType = cast.ToString(data[`vip_type`])
	vipInfo.ExpiredTime = cast.ToString(data[`expired_time`])
	return vipInfo
}

// QueryEnvWechatKefuList 查询微信客服
// @auth frog
// @date 2023-04-12 09:12:46
func QueryEnvWechatKefuList(adminUserId string) string {
	appList := make([]TblWechatapp, 0)
	//goland:noinspection ALL
	dataList, err := XkfDevMysql.GetAll(`select app_name,app_id,app_type from tbl_wechatapp where user_id = ? and app_type = ?`, adminUserId, `wechat_kefu`)
	if err != nil {
		return fmt.Sprintf(`执行sql出错 %s`, err.Error())
	}
	for _, data := range dataList {
		appInfo := TblWechatapp{}
		appInfo.AppType = cast.ToString(data[`app_type`])
		appInfo.Appid = cast.ToString(data[`app_id`])
		appInfo.AppName = cast.ToString(data[`app_name`])
		appList = append(appList, appInfo)
	}
	return gstool.JsonEncode(appList)
}

//QueryOneWechatAppIdChannelId 拿一个应用ID和渠道
func QueryOneWechatAppIdChannelId(userId int) (string, string) {
	appInfo, err := XkfDevMysql.GetOne(`select wechatapp_id from tbl_staff_wechatapp_relation where user_id = ?`, userId)
	fmt.Println(fmt.Sprintf(`%#v`, appInfo))
	if err != nil {
		Logger.Errorf(`执行错误 %s`, err.Error())
	}
	channelInfo, err := XkfDevMysql.GetOne(`select channel_id from tbl_channel_user_rel where user_id = ? and wechatapp_id = ? and status = 1`, userId, cast.ToInt(appInfo[`wechatapp_id`]))
	fmt.Println(fmt.Sprintf(`%#v`, channelInfo))
	if err != nil {
		Logger.Errorf(`执行错误 %s`, err.Error())
	}
	return cast.ToString(appInfo[`wechatapp_id`]), cast.ToString(channelInfo[`channel_id`])
}
