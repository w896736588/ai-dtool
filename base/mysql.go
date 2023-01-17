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
func UpdateVip(adminUserId, systemType, vipLevel string, dbConfig define.MysqlConfig) string {
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

	var ret sql.Result
	var err error
	if vipLevel == `0` { //免费版
		expiredTime := time.Now().Format(`2006-01-02 15:04:05`)
		sqlStr := `update ` + vipTable + ` set expired_time = ? where user_id =?`
		log.Debugf(`更新vip表 %s %s %s %s`, vipTable, sqlStr, expiredTime, adminUserId)
		ret, err = db.Exec(sqlStr, expiredTime, adminUserId)

	} else {
		expiredTime := time.Unix(cast.ToInt64(time.Now().Unix()+365*86400), 0).Format(`2006-01-02 15:04:05`)
		sqlStr := `update ` + vipTable + ` set expired_time = ? , vip_type = ? where user_id =?`
		log.Debugf(`更新vip表 %s %s %s %s %s`, vipTable, sqlStr, vipLevel, expiredTime, adminUserId)
		ret, err = db.Exec(sqlStr, expiredTime, vipLevel, adminUserId)
	}
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
