package base_module

import "gitee.com/Sxiaobai/gs/gstool"

func Register(global *Global, register *RegisterStruct) {
	if len(register.RedisConfigList) > 0 {
		for _, value := range register.RedisConfigList {
			global.RedisSetConfig(value)
		}
	}
	if len(register.MysqlConfigList) > 0 {
		for _, value := range register.MysqlConfigList {
			global.MysqlSetConfig(value)
		}
	}
	if len(register.ShellConfigList) > 0 {
		for _, value := range register.ShellConfigList {
			global.ShellSetConfig(value)
			//初始化client
			gstool.FmtPrintlnLog(`注册服务时获取client %s`, value.Name)
			_, err := global.ShellPushGetClient(value.Name)
			if err != nil {
				return
			}
		}
	}
	if register.EncryptIv != `` && register.EncryptKey != `` {
		global.SetEncrypt(register.EncryptKey, register.EncryptIv)
	}
}
