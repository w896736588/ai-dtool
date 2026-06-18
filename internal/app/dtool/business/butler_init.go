package business

import (
	"dev_tool/internal/app/dtool/butler"
	"dev_tool/internal/app/dtool/butler/bot"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// ButlerRuntime 管家运行时，管理所有已启用的机器人网关和管家核心实例。
// 支持多机器人同时连接：每个 status=1 的机器人配置对应一个 Gateway，
// 所有 Gateway 的消息合并投递到管家核心的统一消息通道。
// 同时实现 bot.GatewayProvider 接口，供管家核心获取 Gateway 实例。
type ButlerRuntime struct {
	db         *common.CSqlite
	butlerEnv  *define.ButlerEnv
	core       *butler.Core
	msgChan    chan bot.IncomingMessage
	gateways   map[int]bot.Gateway           // botConfigId → Gateway
	botConfigs map[int]*define.BotConfigItem // botConfigId → BotConfigItem
	mu         sync.Mutex
}

// NewButlerRuntime 创建管家运行时实例，db 为管家数据库，butlerEnv 为管家运行时环境。
func NewButlerRuntime(db *common.CSqlite, butlerEnv *define.ButlerEnv) *ButlerRuntime {
	return &ButlerRuntime{
		db:         db,
		butlerEnv:  butlerEnv,
		msgChan:    make(chan bot.IncomingMessage, 128),
		gateways:   make(map[int]bot.Gateway),
		botConfigs: make(map[int]*define.BotConfigItem),
	}
}

// Start 启动管家运行时：先连接所有已启用的机器人，再尝试加载管家配置并启动核心。
// 管家配置不存在时仅跳过核心启动，机器人仍可正常连接/断开。
func (r *ButlerRuntime) Start() error {
	// 1. 优先连接所有已启用的机器人（不依赖管家配置）
	r.connectAllEnabledBots()

	// 2. 加载管家配置（取 status=1 的第一条）
	butlerConfig, err := r.loadButlerConfig()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 管家配置未就绪 %s，跳过AI核心启动（机器人仍可用）`, err.Error())
		return nil
	}
	gstool.FmtPrintlnLogTime(`[butler] 管家配置: name=%s role_id=%d bot_config_id=%d`,
		butlerConfig.Name, butlerConfig.RoleId, butlerConfig.BotConfigId)

	// 3. 加载角色
	role, err := r.loadRole(butlerConfig.RoleId)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 加载角色失败 %s，使用默认空角色`, err.Error())
		role = &define.RoleItem{}
	}
	gstool.FmtPrintlnLogTime(`[butler] 角色: name=%s persona=%s`, role.Name, role.Persona)

	// 4. 创建管家核心（传入 ButlerRuntime 作为 GatewayProvider）
	r.core = butler.NewCore(r.db, butlerConfig, r.butlerEnv, role, r, r.msgChan)
	r.core.Start()

	gstool.FmtPrintlnLogTime(`[butler] 管家运行时已启动，已连接 %d 个机器人`, len(r.gateways))
	return nil
}

// RestartCore 根据最新管家配置重新启动核心。
// 先停止已有核心（如有），再重新加载配置并启动新核心。
// 管家配置不存在时仅停止旧核心，不报错。
func (r *ButlerRuntime) RestartCore() {
	r.mu.Lock()
	if r.core != nil {
		r.core.Stop()
		r.core = nil
	}
	r.mu.Unlock()

	butlerConfig, err := r.loadButlerConfig()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] RestartCore: 管家配置未就绪 %s，核心已停止`, err.Error())
		return
	}
	gstool.FmtPrintlnLogTime(`[butler] RestartCore: 管家配置 name=%s role_id=%d bot_config_id=%d`,
		butlerConfig.Name, butlerConfig.RoleId, butlerConfig.BotConfigId)

	role, err := r.loadRole(butlerConfig.RoleId)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] RestartCore: 加载角色失败 %s，使用默认空角色`, err.Error())
		role = &define.RoleItem{}
	}
	gstool.FmtPrintlnLogTime(`[butler] RestartCore: 角色 name=%s persona=%s`, role.Name, role.Persona)

	r.mu.Lock()
	r.core = butler.NewCore(r.db, butlerConfig, r.butlerEnv, role, r, r.msgChan)
	r.mu.Unlock()
	r.core.Start()
	gstool.FmtPrintlnLogTime(`[butler] RestartCore: 核心已重启，已连接 %d 个机器人`, len(r.gateways))
}

// Stop 停止管家运行时：断开所有机器人连接 → 停止管家核心。
func (r *ButlerRuntime) Stop() {
	gstool.FmtPrintlnLogTime(`[butler] 开始停止管家运行时`)
	if r.core != nil {
		r.core.Stop()
	}
	r.mu.Lock()
	for id, gw := range r.gateways {
		gw.Close()
		r.updateBotConnStatus(id, define.ConnStatusDisconnected, ``)
	}
	r.gateways = nil
	r.botConfigs = nil
	r.mu.Unlock()
	gstool.FmtPrintlnLogTime(`[butler] 管家运行时已停止`)
}

// ConnectBot 连接指定机器人配置。如果该机器人已连接则跳过。
// 成功连接后更新数据库连接状态为 ConnStatusConnected。
func (r *ButlerRuntime) ConnectBot(botConfig *define.BotConfigItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.gateways[botConfig.Id]; exists {
		gstool.FmtPrintlnLogTime(`[butler] 机器人 %s (id=%d) 已连接，跳过`, botConfig.Name, botConfig.Id)
		return nil
	}

	gateway := bot.NewDingTalkGateway(botConfig, r.msgChan)
	if err := gateway.Start(); err != nil {
		r.updateBotConnStatus(botConfig.Id, define.ConnStatusFailed, err.Error())
		return err
	}
	r.gateways[botConfig.Id] = gateway
	r.botConfigs[botConfig.Id] = botConfig
	r.updateBotConnStatus(botConfig.Id, define.ConnStatusConnected, ``)
	gstool.FmtPrintlnLogTime(`[butler] 机器人 %s (id=%d) 连接成功`, botConfig.Name, botConfig.Id)
	return nil
}

// DisconnectBot 断开指定机器人的连接，并更新数据库连接状态。
func (r *ButlerRuntime) DisconnectBot(botConfigId int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	gw, exists := r.gateways[botConfigId]
	if !exists {
		return
	}
	gw.Close()
	delete(r.gateways, botConfigId)
	delete(r.botConfigs, botConfigId)
	r.updateBotConnStatus(botConfigId, define.ConnStatusDisconnected, ``)
	gstool.FmtPrintlnLogTime(`[butler] 机器人 (id=%d) 已断开连接`, botConfigId)
}

// IsBotConnected 查询指定机器人是否已连接。
func (r *ButlerRuntime) IsBotConnected(botConfigId int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.gateways[botConfigId]
	return exists
}

// GetGateway 获取指定机器人的网关实例（bot.GatewayProvider 接口实现）。
func (r *ButlerRuntime) GetGateway(botConfigId int) bot.Gateway {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.gateways[botConfigId]
}

// GetAllGateways 返回所有已连接的机器人网关（bot.GatewayProvider 接口实现）。
func (r *ButlerRuntime) GetAllGateways() map[int]bot.Gateway {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 返回副本避免外部并发修改
	result := make(map[int]bot.Gateway, len(r.gateways))
	for id, gw := range r.gateways {
		result[id] = gw
	}
	return result
}

// connectAllEnabledBots 扫描 tbl_butler_bot_config 中所有 status=1 的记录并逐一连接。
func (r *ButlerRuntime) connectAllEnabledBots() {
	rows, err := r.db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_bot_config WHERE status = 1 ORDER BY id ASC`,
	).All()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 查询已启用的机器人配置失败 %s`, err.Error())
		return
	}
	for _, row := range rows {
		botConfig := r.rowToBotConfigItem(row)
		if err := r.ConnectBot(botConfig); err != nil {
			gstool.FmtPrintlnLogTime(`[butler] 机器人 %s (id=%d) 连接失败 %s`, botConfig.Name, botConfig.Id, err.Error())
			// 连接失败不 panic，继续连接其他机器人
		}
	}
}

// loadButlerConfig 从共用库读取启用的管家配置（status=1 的第一条）。
func (r *ButlerRuntime) loadButlerConfig() (*define.ButlerConfigItem, error) {
	row, err := r.db.Client.QueryBySql(
		`SELECT * FROM tbl_butler_config WHERE status = 1 ORDER BY id ASC`,
	).One()
	if err != nil {
		return nil, err
	}
	if len(row) == 0 {
		return nil, fmt.Errorf(`未找到启用的管家配置，请在 dtool 中配置 tbl_butler_config`)
	}
	return r.rowToButlerConfigItem(row), nil
}

// loadRole 根据 roleId 读取角色配置。
func (r *ButlerRuntime) loadRole(roleId int) (*define.RoleItem, error) {
	row, err := r.db.Client.QueryBySql(
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

// rowToBotConfigItem 将数据库行映射为 BotConfigItem 结构体。
func (r *ButlerRuntime) rowToBotConfigItem(row map[string]any) *define.BotConfigItem {
	return &define.BotConfigItem{
		Id:        cast.ToInt(row[`id`]),
		Platform:  cast.ToString(row[`platform`]),
		Name:      cast.ToString(row[`name`]),
		AppKey:    cast.ToString(row[`app_key`]),
		AppSecret: cast.ToString(row[`app_secret`]),
		RobotCode: cast.ToString(row[`robot_code`]),
		Status:    cast.ToInt(row[`status`]),
	}
}

// rowToButlerConfigItem 将数据库行映射为 ButlerConfigItem 结构体。
func (r *ButlerRuntime) rowToButlerConfigItem(row map[string]any) *define.ButlerConfigItem {
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
	}
}

// updateBotConnStatus 更新机器人连接状态到 tbl_butler_bot_config。
func (r *ButlerRuntime) updateBotConnStatus(botConfigId, status int, errMsg string) {
	_, err := r.db.Client.QuickUpdate(`tbl_butler_bot_config`, map[string]any{`id`: botConfigId}, map[string]any{
		`conn_status`:    status,
		`conn_status_at`: time.Now().Unix(),
		`conn_error`:     errMsg,
	}).Exec()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 更新连接状态失败 %s`, err.Error())
	}
}

// InitButlerRuntime 构建管家运行时环境并启动管家，连接所有已启用的机器人。
// 供 dtool config.go 的 initButlerRuntime 调用。
func InitButlerRuntime() {
	butlerEnv := buildButlerEnv()
	runtime := NewButlerRuntime(component.DbMain, butlerEnv)
	component.ButlerRuntime = runtime
	if err := runtime.Start(); err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 管家运行时启动失败 %s，管家功能不可用`, err.Error())
		// 不 panic，允许 dtool 在无管家的情况下继续运行
		return
	}
}

// GetButlerRuntime 获取管家运行时实例的完整类型，供 controller 调用 ConnectBot 等方法。
// 返回 nil 表示管家运行时未初始化。
func GetButlerRuntime() *ButlerRuntime {
	if component.ButlerRuntime == nil {
		return nil
	}
	runtime, ok := component.ButlerRuntime.(*ButlerRuntime)
	if !ok {
		return nil
	}
	return runtime
}

// buildButlerEnv 从 dtool Env 构建 butler 的 define.ButlerEnv，供管家内部包使用。
// 管家表已合并到主库，不再使用独立数据库文件。
func buildButlerEnv() *define.ButlerEnv {
	env := component.EnvClient
	memoryDbPath := common.ResolveDefaultDToolDir(env.ConfigBase.MemoryDBPath)
	// 构建 dtool API 基地址：默认取第一个 API 端口
	baseURL := `http://localhost:17170`
	if len(env.ApiPorts) > 0 {
		baseURL = fmt.Sprintf(`http://localhost:%s`, env.ApiPorts[0])
	}
	return &define.ButlerEnv{
		RootPath:      env.RootPath,
		ConfigPath:    env.ConfigPath,
		ConfigFile:    env.ConfigFile,
		DbPath:        env.DbConfig.DbPath,
		DbName:        env.DbConfig.DbName,
		LogDbPath:     env.LogDbConfig.DbPath,
		MemoryDbPath:  memoryDbPath,
		DatabaseUpDir: env.DatabaseUpPath,
		LogPath:       env.LogPath,
		DtoolBaseURL:  baseURL,
	}
}

// DingtalkSendSingleChatMsg 通过钉钉 Open API 发送单聊文本消息。
// appKey/appSecret 用于获取 access_token，robotCode 为机器人编码，userId 为接收者内部 ID。
// 供 controller 层测试连通性时使用。
func DingtalkSendSingleChatMsg(appKey, appSecret, robotCode, userId, text string) error {
	return bot.SendDingtalkSingleChatMsg(appKey, appSecret, robotCode, userId, text)
}

// DingtalkGetAccessToken 获取钉钉 access_token（供测试接口获取管理员列表时使用）。
func DingtalkGetAccessToken(appKey, appSecret string) (string, error) {
	return bot.GetDingtalkAccessToken(appKey, appSecret)
}

// DingtalkGetAppAdmins 获取钉钉应用管理员列表（供测试接口获取接收者时使用）。
func DingtalkGetAppAdmins(accessToken, appKey string) ([]string, error) {
	return bot.GetDingtalkAppAdmins(accessToken, appKey)
}
