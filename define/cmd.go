package define

//  命令对应
var Cmd = map[string]string{}

//  supervisor 配置文件
var SupervisorList = map[string]CmdStruct{
	`AiHistoryExportTask`: {
		ConfigFile: `/var/www/dockerfiles/dev_test/docker_volumes/supervisor/etc/supervisor/conf.d/historyExportTask.conf`,
		Name:       `Ai历史会话导出`,
	},
}
