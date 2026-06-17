package main

import (
	"context"
	dtool_butler "dev_tool/internal/app/dtool-butler"
	"dev_tool/internal/app/dtool-butler/bot"
	"dev_tool/internal/app/dtool-butler/butler"
	"dev_tool/internal/app/dtool-butler/define"
	"dev_tool/internal/app/dtool/common"
	"flag"
	"fmt"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

var ConfigFile string

func main() {
	flag.StringVar(&ConfigFile, `ConfigFile`, `config`, `配置文件名（不含扩展名）`)
	flag.Parse()

	// 初始化环境与数据库
	dtool_butler.Env = dtool_butler.InitEnv(ConfigFile)
	dtool_butler.DbMain = dtool_butler.InitSqlite(dtool_butler.Env)
	// 执行管家自己的 migration
	dtool_butler.RunMigration(dtool_butler.DbMain, dtool_butler.Env)

	// 加载管家配置（取 status=1 的第一条）
	butlerConfig, err := loadButlerConfig(dtool_butler.DbMain)
	if err != nil {
		panic(fmt.Sprintf(`[butler] 加载管家配置失败 %s`, err.Error()))
	}
	gstool.FmtPrintlnLogTime(`[butler] 管家配置: name=%s role_id=%d bot_config_id=%d`,
		butlerConfig.Name, butlerConfig.RoleId, butlerConfig.BotConfigId)

	// 加载角色
	role, err := loadRole(dtool_butler.DbMain, butlerConfig.RoleId)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 加载角色失败 %s，使用默认空角色`, err.Error())
		role = &define.RoleItem{}
	}
	gstool.FmtPrintlnLogTime(`[butler] 角色: name=%s persona=%s`, role.Name, role.Persona)

	// 加载机器人配置
	botConfig, err := loadBotConfig(dtool_butler.DbMain, butlerConfig.BotConfigId)
	if err != nil {
		panic(fmt.Sprintf(`[butler] 加载机器人配置失败 %s`, err.Error()))
	}
	gstool.FmtPrintlnLogTime(`[butler] 机器人: platform=%s name=%s`, botConfig.Platform, botConfig.Name)

	// 创建消息通道
	msgChan := make(chan bot.IncomingMessage, 128)

	// 创建并启动钉钉网关
	gateway := bot.NewDingTalkGateway(botConfig, msgChan)
	if err := gateway.Start(); err != nil {
		panic(fmt.Sprintf(`[butler] 钉钉网关启动失败 %s`, err.Error()))
	}
	dtool_butler.BotGateway = gateway

	// 创建并启动管家核心
	core := butler.NewCore(dtool_butler.DbMain, butlerConfig, role, gateway, msgChan)
	dtool_butler.ButlerCore = core
	core.Start()

	// 阻塞等待退出信号
	gstool.SignalDefault()
	gstool.FmtPrintlnLogTime(`[butler] 开始停止`)
	core.Stop()
	gateway.Close()
	_ = context.Background()
}

// loadButlerConfig 从共用库读取启用的管家配置（status=1 的第一条）。
func loadButlerConfig(db *common.CSqlite) (*define.ButlerConfigItem, error) {
	row, err := db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_config WHERE status = 1 ORDER BY id ASC`,
	).One()
	if err != nil {
		return nil, err
	}
	if len(row) == 0 {
		return nil, fmt.Errorf(`未找到启用的管家配置，请在 dtool 中配置 tbl_butler_config`)
	}
	return &define.ButlerConfigItem{
		Id:                   cast.ToInt(row[`id`]),
		Name:                 cast.ToString(row[`name`]),
		RoleId:               cast.ToInt(row[`role_id`]),
		ModelId:              cast.ToInt(row[`model_id`]),
		FcModelId:            cast.ToInt(row[`fc_model_id`]),
		AgentCliId:           cast.ToInt(row[`agent_cli_id`]),
		BotConfigId:          cast.ToInt(row[`bot_config_id`]),
		ActiveTimeoutMinutes: cast.ToInt(row[`active_timeout_minutes`]),
		MaxHistory:           cast.ToInt(row[`max_history`]),
		AutoCleanOnNewTopic:  cast.ToInt(row[`auto_clean_on_new_topic`]),
		IndexDocPath:         cast.ToString(row[`index_doc_path`]),
		AutoInitOnStart:      cast.ToInt(row[`auto_init_on_start`]),
		Status:               cast.ToInt(row[`status`]),
	}, nil
}

// loadRole 根据 roleId 读取角色配置。
func loadRole(db *common.CSqlite, roleId int) (*define.RoleItem, error) {
	row, err := db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_role WHERE id = ? AND status = 1`, roleId,
	).One()
	if err != nil {
		return nil, err
	}
	if len(row) == 0 {
		return nil, fmt.Errorf(`未找到角色 id=%d`, roleId)
	}
	return &define.RoleItem{
		Id:           cast.ToInt(row[`id`]),
		Name:         cast.ToString(row[`name`]),
		Persona:      cast.ToString(row[`persona`]),
		Tone:         cast.ToString(row[`tone`]),
		SystemPrompt: cast.ToString(row[`system_prompt`]),
		InitGreeting: cast.ToString(row[`init_greeting`]),
		Status:       cast.ToInt(row[`status`]),
	}, nil
}

// loadBotConfig 根据 botConfigId 读取机器人配置。
func loadBotConfig(db *common.CSqlite, botConfigId int) (*define.BotConfigItem, error) {
	row, err := db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_bot_config WHERE id = ? AND status = 1`, botConfigId,
	).One()
	if err != nil {
		return nil, err
	}
	if len(row) == 0 {
		return nil, fmt.Errorf(`未找到机器人配置 id=%d`, botConfigId)
	}
	return &define.BotConfigItem{
		Id:         cast.ToInt(row[`id`]),
		Platform:   cast.ToString(row[`platform`]),
		Name:       cast.ToString(row[`name`]),
		AppKey:     cast.ToString(row[`app_key`]),
		AppSecret:  cast.ToString(row[`app_secret`]),
		RobotCode:  cast.ToString(row[`robot_code`]),
		WebhookUrl: cast.ToString(row[`webhook_url`]),
		Secret:     cast.ToString(row[`secret`]),
		Status:     cast.ToInt(row[`status`]),
	}, nil
}
