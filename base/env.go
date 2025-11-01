package base

import (
	"os"
	"path/filepath"

	"gitee.com/Sxiaobai/gs/gstool"
)

type Env struct {
	RootPath           string   //项目根目录
	PkgPath            string   //pkg目录
	DbName             string   //数据库文件名
	AppName            string   //项目名称
	ConfigPath         string   //配置文件目录
	DatabaseUpPath     string   //数据库升级目录
	LogPath            string   //日志目录
	ViewPath           string   //前端目录
	DbPath             string   //配置数据库目录
	NodePath           string   //node js可执行程序目录
	WebkitDriverPath   string   //浏览器核心目录
	WebkitDownloadPath string   //浏览器核心下载临时文件数据目录
	WebkitDataPath     string   //浏览器核心用户数据目录
	Ports              []string //gin支持的端口
}

func (h *Env) Init(appName, dbPath, DbName, ViewPath string) {
	if h.RootPath == `` {
		panic(`root_path不能为空`)
	}
	h.DbName = DbName
	h.AppName = appName
	dbFileName := h.AppName + `.db`
	if DbName != `` {
		dbFileName = DbName
	}
	//配置文件目录
	if dbPath != `` {
		h.DbPath = filepath.Join(dbPath, dbFileName)
	} else {
		h.DbPath = filepath.Join(h.RootPath, `config`, h.AppName, dbFileName)
	}
	//前端目录
	if ViewPath == `` {
		h.ViewPath = filepath.Join(filepath.Dir(h.RootPath), `devtool`, `dist`)
	} else {
		h.ViewPath = ViewPath
	}
	//基础
	h.ConfigPath = filepath.Join(Component.Env.RootPath, `config`, Component.Env.AppName)
	//配置初始化
	Component.ConfigViper.AddConfigPath(h.ConfigPath)
	Component.ConfigViper.SetConfigName(`config`)
	Component.ConfigViper.SetConfigType(`ini`)
	if readErr := Component.ConfigViper.ReadInConfig(); readErr != nil {
		panic(readErr.Error())
	}
	h.PkgPath = filepath.Join(h.RootPath, `internal`, `pkg`)
	h.LogPath = filepath.Join(h.RootPath, `logs`)
	//webkit
	h.NodePath = gstool.SReplaces(Component.ConfigViper.GetString(`path.webkit_node_path`), map[string]string{
		`{PKG_PATH}`: h.PkgPath,
	})
	//判断是否存在D盘如果没有那么就改为C盘
	drive := ``
	drivePath := string(`D`) + ":\\"
	_, err := os.Stat(drivePath)
	if err == nil {
		drive = `D`
	} else {
		drive = `C`
	}
	h.WebkitDriverPath = Component.ConfigViper.GetString(`path.webkit_driver_path`)
	h.WebkitDataPath = Component.ConfigViper.GetString(`path.webkit_data_path`)
	h.WebkitDownloadPath = Component.ConfigViper.GetString(`path.webkit_download_path`)
	h.WebkitDataPath = gstool.SReplaces(h.WebkitDataPath, map[string]string{
		`{DRIVE}`: drive,
	})
	h.WebkitDownloadPath = gstool.SReplaces(h.WebkitDownloadPath, map[string]string{
		`{DRIVE}`: drive,
	})
	h.WebkitDriverPath = gstool.SReplaces(h.WebkitDriverPath, map[string]string{
		`{DRIVE}`: drive,
	})
	//创建目录
	_ = gstool.DirCreatePath(h.LogPath)
	_ = gstool.DirCreatePath(h.ViewPath)
	_ = gstool.DirCreatePath(h.DbPath)
	_ = gstool.DirCreatePath(h.WebkitDataPath)
	_ = gstool.DirCreatePath(h.WebkitDriverPath)
	_ = gstool.DirCreatePath(h.WebkitDownloadPath)
	gstool.FmtPrintlnLogTime(`输出配置：`)
	gstool.FmtPrintlnLogTime(gstool.JsonFormat(h))
}
