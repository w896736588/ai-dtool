package base_module

func Register(global *Global, register *RegisterStruct) {
	//if len(register.RedisConfigList) > 0 {
	//	for _, value := range register.RedisConfigList {
	//		global.RedisSetConfig(value)
	//	}
	//}
	if len(register.MysqlConfigList) > 0 {
		for _, value := range register.MysqlConfigList {
			global.MysqlSetConfig(value)
		}
	}
	//if len(register.ShellConfigList) > 0 {
	//	for _, gsShell := range register.ShellConfigList {
	//		gsShell.GsSlog = global.logger
	//		connectErr := gsShell.ConnectAuthPassword()
	//		if connectErr != nil {
	//			Logger.Errof(`连接失败 %s`, connectErr.Error())
	//			return
	//		}
	//		_, err := gsShell.RunCommandWait(`pwd`)
	//		if err != nil {
	//			Logger.Errof(`执行命令失败 %s`, err.Error())
	//			return
	//		}
	//		//设置回调
	//		gsShell.SetFuncBefore(func(command string) string {
	//			return `■■ ` + command
	//		})
	//		gsShell.SetCombineNum(1)
	//		gsShell.CloseFirstReceiveMsg()
	//		global.ShellSet(gsShell)
	//	}
	//}
	if register.EncryptIv != `` && register.EncryptKey != `` {
		global.SetEncrypt(register.EncryptKey, register.EncryptIv)
	}
}
