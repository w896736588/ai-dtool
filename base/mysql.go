package base

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"redis_manager/define"
	"time"
)

var Conn map[string]*sql.DB

// QueryWechatAppid 拿到应用的AppId https://www.cnblogs.com/jasonminghao/p/12386580.html
// @auth frog
// @date 2023-01-17 12:25:31
func QueryWechatAppid(wechatAppId string, dbConfig define.MysqlConfig) define.TblWechatapp {
	var appInfo define.TblWechatapp
	dbConfig.Dbname = `xkf_test`
	db := GetDbConn(dbConfig)
	if db == nil {
		return appInfo
	}
	err := db.QueryRow(`select app_id,app_type,user_id from tbl_wechatapp where (app_id = ? or _id = ?) and user_id > 0`, wechatAppId, cast.ToInt(wechatAppId)).Scan(&appInfo.Appid, &appInfo.AppType, &appInfo.UserId)
	if err != nil {
		log.Errorf(`执行sql出错 %s`, err.Error())
		return appInfo
	}
	return appInfo
}

// GetDbConn 拿到连接
// @auth frog
// @date 2023-01-17 14:09:48
func GetDbConn(dbConfig define.MysqlConfig) *sql.DB {
	if len(Conn) == 0 {
		Conn = make(map[string]*sql.DB)
	}
	dsn := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s`, dbConfig.Username, dbConfig.Password, dbConfig.Host, cast.ToString(dbConfig.Port), dbConfig.Dbname)
	log.Debugf(`链接串 %s`, dsn)
	if Conn[dsn] != nil {
		return Conn[dsn]
	}
	db, err := sql.Open(`mysql`, dsn) // 不会校验用户名和密码是否正确
	if err != nil {                   // dsn 格式不正确的时候会报错
		log.Error(`连接mysql报错 %s`, err.Error())
		return nil
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Minute)
	Conn[dsn] = db
	return db
}

// GetAdminUserId 拿到用户信息
// @auth frog
// @date 2023-01-17 15:49:51
func GetAdminUserId(account string, dbConfig define.MysqlConfig) define.TblUser {
	dbConfig.Dbname = `appurl_test`
	db := GetDbConn(dbConfig)

	var userInfo define.TblUser
	if db == nil {
		return userInfo
	}
	err := db.QueryRow(`select _id,user_name from tbl_user where (user_name = ? or _id = ?)`, account, cast.ToInt(account)).Scan(&userInfo.Id, &userInfo.Username)
	if err != nil {
		log.Errorf(`执行sql出错 %s`, err.Error())
		return userInfo
	}
	return userInfo
}

// UpdateVip 变更vip版本
// @auth frog
// @date 2023-01-17 16:03:11
func UpdateVip(adminUserId, expiredDay, systemType, vipLevel string, dbConfig define.MysqlConfig) string {
	dbConfig.Dbname = `xkf_test`
	db := GetDbConn(dbConfig)
	if db == nil {
		return `连接失败`
	}
	vipTable := `tbl_kefu_vip`
	if systemType == `1` { //客服系统
		vipTable = `tbl_kefu_vip`
	} else {
		vipTable = `tbl_official_account_vip`
	}
	//时间
	t := time.Now().Unix()
	t += 86400 * cast.ToInt64(expiredDay)
	var ret sql.Result
	var err error
	expiredTime := time.Unix(t, 0).Format(`2006-01-02 15:04:05`)
	sqlStr := `update ` + vipTable + ` set expired_time = ? , vip_type = ? where user_id =?`
	log.Debugf(`更新vip表 %s %s %s %s %s`, vipTable, sqlStr, vipLevel, expiredTime, adminUserId)
	ret, err = db.Exec(sqlStr, expiredTime, vipLevel, adminUserId)
	if err != nil {
		log.Errorf(`更新%s失败 %s`, vipTable, err.Error())
		return `更新失败`
	}
	n, err := ret.RowsAffected()
	if err != nil {
		log.Errorf(`更新%s失败 %s`, vipTable, err.Error())
		return `更新成功 ` + cast.ToString(n)
	}

	return `成功`
}

// QueryVip 查询VIP信息
// @auth frog
// @date 2023-03-16 09:31:34
func QueryVip(adminUserId, systemType string, dbConfig define.MysqlConfig) *define.TblVip {
	var vipInfo = &define.TblVip{}
	dbConfig.Dbname = `xkf_test`
	db := GetDbConn(dbConfig)
	if db == nil {
		return vipInfo
	}
	vipTable := `tbl_kefu_vip`
	if systemType == `1` { //客服系统
		vipTable = `tbl_kefu_vip`
	} else {
		vipTable = `tbl_official_account_vip`
	}

	errQuery := db.QueryRow(`select vip_type,expired_time from `+vipTable+` where user_id = ?`, adminUserId).Scan(&vipInfo.VipType, &vipInfo.ExpiredTime)
	if errQuery != nil {
		log.Errorf(`执行sql出错 %s`, errQuery.Error())
		return vipInfo
	}
	return vipInfo
}
