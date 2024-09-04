package base_module

import (
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsssh"
)

type RegisterStruct struct {
	Unikey          string              `json:"Unikey"`
	MysqlConfigList []*gsdb.MysqlConfig `json:"MysqlConfigList"`
	RedisConfigList []*gsdb.RedisConfig `json:"RedisConfigList"`
	ShellConfigList []*gsssh.SshConfig  `json:"ShellConfigList"`
	EncryptKey      string              `json:"EncryptKey"`
	EncryptIv       string              `json:"EncryptIv"`
}

type LoginStruct struct {
	UserName string `json:"UserName"`
	Password string `json:"Password"`
}

type CodeSearch struct {
	DirPathList []string `json:"dir_path_list"`
	Token       string   `json:"token"`
}
