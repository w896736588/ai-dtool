package butler

import (
	"context"
	"dev_tool/internal/app/dtool/butler/bot"
	"dev_tool/internal/app/dtool/butler/index"
	"dev_tool/internal/app/dtool/butler/worker"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"fmt"
	"strings"
	"time"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// Core 管家核心，负责消息消费、激活态管理、命令路由、AI 回复、休眠巡检。
type Core struct {
	db              *common.CSqlite
	config          *define.ButlerConfigItem
	env             *define.ButlerEnv
	role            *define.RoleItem
	systemPrompt    string
	gatewayProvider bot.GatewayProvider // 多机器人场景下的网关提供者
	history         *History
	sessions        *SessionManager
	msgChan         <-chan bot.IncomingMessage
	replier         *chatbot.ChatbotReplier
	stopCh          chan struct{}
	indexPath       string          // 索引文档目录路径
	skillsRoot      string          // skills 目录绝对路径
	greetedSessions map[string]bool // 已发送过打招呼语的会话 ID，确保每次启动后每会话仅发送一次
	// 归档提交回调，由业务层注入。主管家任务完成后将文件+对话异步提交到归档管家。
	// 返回新创建或更新后的归档记录 ID。
	archiveSubmit       func(configId, taskId int, sessionId string, files []string, conversation string) int
	lastFilesWritten    []string            // 最近一次 FC 循环产生的文件路径
	sessionFilesWritten map[string][]string // 会话级累积文件路径（key=conversationId）
}

// NewCore 创建管家核心。msgChan 为机器人网关投递的消息通道。
// gatewayProvider 为网关提供者，用于多机器人场景下获取 Gateway 实例。
func NewCore(
	db *common.CSqlite,
	config *define.ButlerConfigItem,
	env *define.ButlerEnv,
	role *define.RoleItem,
	gatewayProvider bot.GatewayProvider,
	msgChan <-chan bot.IncomingMessage,
) *Core {
	timeout := time.Duration(config.ActiveTimeoutMinutes) * time.Minute
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	// 历史存储上限默认 100
	if config.MaxHistoryStore <= 0 {
		config.MaxHistoryStore = 100
	}
	// Loop 上限默认 10
	if config.MaxLoop <= 0 {
		config.MaxLoop = 10
	}
	// 预构建 system prompt，避免每条消息重复拼装
	systemPrompt := BuildSystemPrompt(role)
	// 解析索引路径
	indexPath := index.ResolveIndexPath(config, env)
	skillsRoot := index.GetSkillsRoot()
	// 设置 worker 包的 skills 根目录，供文件工具路径解析使用
	worker.SetSkillsRoot(skillsRoot)
	// 设置 worker 包的 dtool API 基地址，供 http_call 工具使用
	worker.SetDtoolBaseURL(env.DtoolBaseURL)
	return &Core{
		db:                  db,
		config:              config,
		env:                 env,
		role:                role,
		systemPrompt:        systemPrompt,
		gatewayProvider:     gatewayProvider,
		history:             NewHistory(db, config.BotConfigId),
		sessions:            NewSessionManager(timeout),
		msgChan:             msgChan,
		replier:             chatbot.NewChatbotReplier(),
		stopCh:              make(chan struct{}),
		indexPath:           indexPath,
		skillsRoot:          skillsRoot,
		greetedSessions:     make(map[string]bool),
		sessionFilesWritten: make(map[string][]string),
	}
}

// SetArchiveSubmit 注入归档提交回调（由业务层调用）。仅主管家生效。
// 回调返回归档记录 ID（新建或更新后）。
func (c *Core) SetArchiveSubmit(fn func(configId, taskId int, sessionId string, files []string, conversation string) int) {
	c.archiveSubmit = fn
}

// Start 启动管家主循环：发打招呼 → 自动初始化索引 → 消费消息 → 定时巡检休眠。非阻塞。
func (c *Core) Start() {
	// 启动打招呼
	c.sendGreeting()
	// 自动初始化索引（auto_init_on_start=1 时）
	if c.config.AutoInitOnStart == 1 {
		c.autoInitIndex()
	}
	// 启动消息消费循环
	go c.consumeLoop()
	// 启动休眠巡检（每 1min）
	go c.timeoutLoop()
	gstool.FmtPrintlnLogTime(`[butler-core] 管家已启动，激活态超时=%v`, time.Duration(c.config.ActiveTimeoutMinutes)*time.Minute)
}

// autoInitIndex 自动初始化索引文档。索引已存在时跳过。
func (c *Core) autoInitIndex() {
	if c.indexPath == `` {
		gstool.FmtPrintlnLogTime(`[butler-core] 索引路径未配置，跳过自动初始化`)
		return
	}
	if index.IndexExists(c.indexPath, index.ScriptsFileName) {
		gstool.FmtPrintlnLogTime(`[butler-core] scripts.md 已存在，跳过自动初始化`)
		return
	}
	content, err := index.InitIndex(c.skillsRoot, c.indexPath)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 自动初始化索引失败 %s`, err.Error())
		return
	}
	lineCount := strings.Count(content, "\n") + 1
	gstool.FmtPrintlnLogTime(`[butler-core] 自动初始化索引完成，scripts.md 共 %d 行`, lineCount)
}

// Stop 停止管家主循环。
func (c *Core) Stop() {
	close(c.stopCh)
}

// sendGreeting 启动时发送打招呼消息。
// 纯流式机器人模式下，没有 userId 无法主动推送，仅在首次收到消息时发送打招呼。
// 此处仅记录打招呼语，实际发送在 handleMessage 中首次激活时触发。
func (c *Core) sendGreeting() {
	if c.role == nil || c.role.InitGreeting == `` {
		gstool.FmtPrintlnLogTime(`[butler-core] 角色未配置打招呼语，跳过`)
		return
	}
	gstool.FmtPrintlnLogTime(`[butler-core] 打招呼语已就绪，将在首次收到消息时发送`)
}

// buildGreeting 构建完整打招呼语：角色打招呼 + 内置命令说明。
// 每次启动后每会话仅发送一次。
func (c *Core) buildGreeting() string {
	if c.role == nil || c.role.InitGreeting == `` {
		return ``
	}
	return c.role.InitGreeting + `\n\n` + builtinCommandsHelp()
}

// consumeLoop 消费消息通道，处理每条消息。
func (c *Core) consumeLoop() {
	for {
		select {
		case <-c.stopCh:
			return
		case msg, ok := <-c.msgChan:
			if !ok {
				return
			}
			c.handleMessage(msg)
		}
	}
}

// timeoutLoop 定时巡检超时会话，触发休眠通知。
// 纯流式模式下无法主动推送（没有 userId），仅记录日志。
// 实际休眠通知将在下次收到消息时，通过 SessionManager 的状态判断来触发。
func (c *Core) timeoutLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			timedOut := c.sessions.CheckTimeout()
			for _, conversationId := range timedOut {
				gstool.FmtPrintlnLogTime(`[butler-core] 会话超时休眠 %s（纯流式模式无法主动推送休眠通知）`, conversationId)
			}
		}
	}
}

// handleMessage 处理单条消息：打招呼 → 激活会话 → 存历史 → 命令路由 → 意图分析 → AI 回复。
func (c *Core) handleMessage(msg bot.IncomingMessage) {
	// 激活会话（刷新最后活跃时间）
	justActivated := c.sessions.Activate(msg.ConversationId)
	if justActivated {
		gstool.FmtPrintlnLogTime(`[butler-core] 会话已激活 %s`, msg.ConversationId)
		// 每次启动后每会话仅发送一次打招呼语（纯流式模式下只能在有消息上下文时推送）
		if !c.greetedSessions[msg.ConversationId] {
			greeting := c.buildGreeting()
			if greeting != `` {
				if err := c.reply(msg, greeting); err != nil {
					gstool.FmtPrintlnLogTime(`[butler-core] 打招呼发送失败 %s`, err.Error())
				}
				gstool.FmtPrintlnLogTime(`[butler-core] 已发送打招呼给 %s`, msg.SenderNick)
			}
			c.greetedSessions[msg.ConversationId] = true
		}
	}
	// 存历史（用户消息），使用消息来源机器人的 botConfigId
	if err := c.history.Append(msg.ConversationId, define.ButlerRoleUser, msg.Text, msg.BotConfigId); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 存用户消息失败 %s`, err.Error())
	}
	// 1. 命令路由
	cmdCtx := &CommandContext{
		IndexPath:  c.indexPath,
		SkillsRoot: c.skillsRoot,
	}
	cmdResult := ParseCommand(msg.Text, c.sessions, c.history, msg.ConversationId, c.config.MaxHistoryStore, cmdCtx)
	if cmdResult.Handled {
		if err := c.reply(msg, cmdResult.Text); err != nil {
			gstool.FmtPrintlnLogTime(`[butler-core] 命令回复失败 %s`, err.Error())
		}
		return
	}
	// 2. 意图分析
	intent := c.analyzeIntent(msg)
	if intent != nil && !intent.Clear && len(intent.Questions) > 0 {
		// 意图不清晰 → 直接返回澄清提问，不进入 AI 主回复
		questionsText := formatClarifyingQuestions(intent.Questions)
		if err := c.reply(msg, questionsText); err != nil {
			gstool.FmtPrintlnLogTime(`[butler-core] 澄清提问回复失败 %s`, err.Error())
		}
		// 存历史（管家追问）
		if err := c.history.AppendWithTopic(msg.ConversationId, define.ButlerRoleAssistant, questionsText, intent.Topic, msg.BotConfigId); err != nil {
			gstool.FmtPrintlnLogTime(`[butler-core] 存追问失败 %s`, err.Error())
		}
		return
	}
	// 3. FC 循环回复（支持 Function Calling 工具调用）
	aiReply, toolsUsed := c.fcReply(msg)
	if err := c.reply(msg, aiReply); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] AI 回复失败 %s`, err.Error())
		return
	}
	// 存历史（管家回复），附带话题标记
	topic := ``
	if intent != nil {
		topic = intent.Topic
	}
	if err := c.history.AppendWithTopic(msg.ConversationId, define.ButlerRoleAssistant, aiReply, topic, msg.BotConfigId); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 存管家回复失败 %s`, err.Error())
	}
	// 回填之前消息的主题（如果主题为空且 intent 有 topic）
	if intent != nil && intent.Topic != `` {
		if err := c.history.UpdateTopicBySession(msg.ConversationId, intent.Topic); err != nil {
			gstool.FmtPrintlnLogTime(`[butler-core] 回填主题失败 %s`, err.Error())
		}
	}
	// 历史存储上限自动清理：超过配置上限时自动删除最旧消息
	if err := c.history.TrimBySession(msg.ConversationId, c.config.MaxHistoryStore); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 历史自动 trim 失败 %s`, err.Error())
	}
	// 有工具调用 → 创建任务记录 → 提交/更新会话级归档
	if len(toolsUsed) > 0 {
		taskId := c.saveTaskRecord(msg.ConversationId, msg.Text, aiReply, toolsUsed)
		// 主管家任务完成后 → 异步提交归档（将对话+文件交给归档管家评估自进化）
		if c.config.ButlerType == define.ButlerTypeMain && c.archiveSubmit != nil {
			// 累积本次 FC 循环产生的文件到会话级别（去重合并）
			c.sessionFilesWritten[msg.ConversationId] = mergeUniqueStrings(c.sessionFilesWritten[msg.ConversationId], c.lastFilesWritten)
			// 获取会话完整对话历史
			conversation := c.getSessionConversation(msg.ConversationId)
			// 同一会话后续轮次会更新已有归档记录，而非创建新记录
			go func(sessionId, conv string, accumulatedFiles []string) {
				archiveId := c.archiveSubmit(c.config.Id, taskId, sessionId, accumulatedFiles, conv)
				gstool.FmtPrintlnLogTime(`[butler-core] 已提交归档 task_id=%d session=%s archive_id=%d files=%d`, taskId, sessionId, archiveId, len(accumulatedFiles))
			}(msg.ConversationId, conversation, c.sessionFilesWritten[msg.ConversationId])
		}
	}
}

// analyzeIntent 对当前消息进行意图分析。使用 fc_model_id（轻量模型），为 0 时回落 model_id。
func (c *Core) analyzeIntent(msg bot.IncomingMessage) *IntentResult {
	intentModelId := c.config.FcModelId
	if intentModelId <= 0 {
		intentModelId = c.config.ModelId
	}
	if intentModelId <= 0 {
		return nil
	}
	// 获取最近对话主题
	recentTopic, err := c.history.GetRecentTopic(msg.ConversationId)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 获取最近主题失败 %s`, err.Error())
		recentTopic = `` // 查询失败视为无历史
	}
	result, err := AnalyzeIntent(c.db, intentModelId, msg.Text, recentTopic)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 意图分析失败 %s，跳过`, err.Error())
		return nil
	}
	return result
}

// formatClarifyingQuestions 将澄清问题列表格式化为回复文本。
func formatClarifyingQuestions(questions []string) string {
	if len(questions) == 0 {
		return ``
	}
	lines := make([]string, 0, len(questions)+1)
	lines = append(lines, `您的意图不太明确，请帮忙澄清：`)
	for i, q := range questions {
		lines = append(lines, fmt.Sprintf(`%d. %s`, i+1, q))
	}
	return strings.Join(lines, `\n`)
}

// fcReply 调用 FC 循环或 Agent CLI 生成回复。
// 先通过 dispatcher 判断任务路由：简单→FC，复杂→Agent CLI。
// 使用 fc_model_id（Function Calling 用模型），为 0 时回落 model_id。
// 返回回复文本和使用过的工具名称列表。
func (c *Core) fcReply(msg bot.IncomingMessage) (string, []string) {
	fcModelId := c.config.FcModelId
	if fcModelId <= 0 {
		fcModelId = c.config.ModelId
	}
	if fcModelId <= 0 {
		gstool.FmtPrintlnLogTime(`[butler-core] 管家模型未配置，回退固定回复`)
		return fmt.Sprintf(`已收到：%s`, msg.Text), nil
	}
	// 任务路由：简单→FC，复杂→Agent CLI
	dispatchResult := worker.Dispatch(c.db, fcModelId, msg.Text, c.config.AgentCliId)
	if dispatchResult.TaskType == worker.TaskTypeAgentCli {
		return c.agentCliReply(msg)
	}
	// FC 循环路径
	return c.fcLoopReply(msg, fcModelId)
}

// fcLoopReply 执行 FC 循环生成回复。
// 分为两个阶段：①规划（检索资源→制定计划→单独发送）②执行（按计划执行→返回结果）。
func (c *Core) fcLoopReply(msg bot.IncomingMessage, fcModelId int) (string, []string) {
	// 加载历史消息（最近 MaxHistoryStore 条）
	historyMessages, err := c.history.ListBySession(msg.ConversationId, c.config.MaxHistoryStore)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 加载历史失败 %s，使用无历史对话`, err.Error())
		historyMessages = nil
	}
	fcHistory := historyToFcMessages(historyMessages)

	// ========== 阶段一：资源检索 + 制定执行计划（单独发送给用户） ==========
	planPrompt := c.systemPrompt + `
## 当前阶段：资源检索与计划制定
收到用户任务后，**只做检索和计划**，不要执行任务。

步骤：
1. 用 file_read 读取 skills/dtool-butler/index/scripts.md，逐条检查是否有可复用的脚本（dtool-butler 节优先于模块通用节）
2. 仅当 scripts.md 中无匹配脚本时，才用 file_read 读取 apis.md 查看可用接口
3. 根据检索结果，制定执行计划并以标准格式输出：

📋 执行计划：
- 任务：<一句话描述>
- 复用脚本：<脚本路径，如无则写"无">
- 调用接口：<接口路径，如无则写"无">
- 新建脚本：<是/否，若是则说明原因>

**输出计划后立即停止，不要调用任何执行类工具（http_call/file_write/run_script 等）。**`

	planResult := worker.RunFCLoop(c.db, fcModelId, planPrompt, fcHistory, msg.Text, 2)
	if planResult.Content != `` {
		if err := c.reply(msg, planResult.Content); err != nil {
			gstool.FmtPrintlnLogTime(`[butler-core] 发送执行计划失败 %s`, err.Error())
		} else {
			gstool.FmtPrintlnLogTime(`[butler-core] 已发送执行计划`)
		}
	}

	// ========== 阶段二：执行任务 ==========
	execPrompt := c.systemPrompt + fcSystemPromptSuffix + `

---
**当前阶段：执行任务**

执行计划已单独发送给用户，现在**直接执行任务**：
- 优先使用已检索到的脚本（无需重复读取 scripts.md）
- 无脚本时调用 API 完成
- 完成后输出结果汇总

**不要重复输出"📋 执行计划"。**`

	if err := c.reply(msg, `正在执行，请稍候...`); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 发送执行提示失败 %s`, err.Error())
	}

	result := worker.RunFCLoop(c.db, fcModelId, execPrompt, fcHistory, msg.Text, c.config.MaxLoop)
	c.lastFilesWritten = result.FilesWritten
	if result.Content == `` {
		return `我暂时无法回复，请稍后再试。`, result.ToolUsed
	}
	// 附加 LLM 用量统计 + 脚本清单
	if result.LLMCalls > 0 {
		usageInfo := "\n\n---\n\n" + fmt.Sprintf("📊 LLM 调用 %d 次 ｜ 输入 %d token ｜ 输出 %d token", result.LLMCalls, result.InputTokens, result.OutputTokens)
		if result.CacheTokens > 0 {
			usageInfo += fmt.Sprintf(" ｜ 缓存命中 %d token", result.CacheTokens)
		}
		if len(result.ScriptsRun) > 0 {
			usageInfo += "\n\n" + fmt.Sprintf("📜 执行脚本：%s", strings.Join(result.ScriptsRun, `, `))
		}
		if len(result.ScriptsCreated) > 0 {
			usageInfo += "\n\n" + fmt.Sprintf("📝 新建脚本：%s", strings.Join(result.ScriptsCreated, `, `))
		}
		return result.Content + usageInfo, result.ToolUsed
	}
	return result.Content, result.ToolUsed
}

// agentCliReply 使用 Agent CLI 执行复杂任务并返回结果。
func (c *Core) agentCliReply(msg bot.IncomingMessage) (string, []string) {
	gstool.FmtPrintlnLogTime(`[butler-core] 任务路由到 Agent CLI，开始执行`)
	// 构建 Agent CLI 的 prompt（包含角色信息 + 用户消息）
	agentPrompt := msg.Text
	if c.systemPrompt != `` {
		agentPrompt = fmt.Sprintf(`[角色设定] %s\n\n[用户任务] %s`, c.systemPrompt, msg.Text)
	}
	// 执行 Agent CLI
	result := worker.RunAgentCli(c.db, c.config.AgentCliId, agentPrompt)
	// 记录任务
	toolsUsed := []string{`agent_cli`}
	if !result.Success {
		// Agent CLI 执行失败 → 创建失败任务记录
		c.saveTaskRecordWithStatus(msg.ConversationId, msg.Text, result.Content, toolsUsed, define.ButlerTaskStatusFailed, `agent_cli`)
		return fmt.Sprintf(`任务执行遇到问题：\n%s`, result.Content), toolsUsed
	}
	// 成功 → 创建完成任务记录
	c.saveTaskRecord(msg.ConversationId, msg.Text, result.Content, toolsUsed)
	return result.Content, toolsUsed
}

// historyToFcMessages 将历史消息列表转换为 FC 循环的 []map[string]string 格式。
func historyToFcMessages(messages []define.ButlerHistoryMessage) []map[string]string {
	result := make([]map[string]string, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == define.ButlerRoleUser || msg.Role == define.ButlerRoleAssistant {
			result = append(result, map[string]string{
				`role`:    msg.Role,
				`content`: msg.Content,
			})
		}
	}
	return result
}

// saveTaskRecord 创建管家任务记录到 tbl_butler_task（状态为 done）。
func (c *Core) saveTaskRecord(sessionId, title, result string, toolsUsed []string) int {
	return c.saveTaskRecordWithStatus(sessionId, title, result, toolsUsed, define.ButlerTaskStatusDone, `fc`)
}

// saveTaskRecordWithStatus 创建管家任务记录到 tbl_butler_task，指定状态和执行器。返回新记录 ID。
func (c *Core) saveTaskRecordWithStatus(sessionId, title, result string, toolsUsed []string, status, executor string) int {
	_, err := c.db.Client.QuickCreate(`tbl_butler_task`, map[string]any{
		`session_id`: sessionId,
		`title`:      title,
		`status`:     status,
		`plan`:       strings.Join(toolsUsed, `,`),
		`result`:     result,
		`executor`:   executor,
		`created_at`: time.Now().Unix(),
		`updated_at`: time.Now().Unix(),
	}).Exec()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 创建任务记录失败 %s`, err.Error())
		return 0
	}
	// 获取自增 ID
	one, _ := c.db.Client.QueryBySql(`SELECT last_insert_rowid() as id`).One()
	taskId := 0
	if len(one) > 0 {
		taskId = cast.ToInt(one[`id`])
	}
	gstool.FmtPrintlnLogTime(`[butler-core] 已创建任务记录 id=%d title=%s executor=%s tools=%v`, taskId, truncateForLog(title, 50), executor, toolsUsed)
	return taskId
}

// getSessionConversation 获取会话的完整对话历史并格式化为文本，供归档提交。
func (c *Core) getSessionConversation(sessionId string) string {
	messages, err := c.history.ListBySession(sessionId, 0) // 0=不限制数量
	if err != nil || len(messages) == 0 {
		return ``
	}
	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", msg.Role, msg.Content))
	}
	return sb.String()
}

// fcSystemPromptSuffix FC 循环的 system prompt 补充说明，指导 AI 使用工具。
const fcSystemPromptSuffix = `

## 可用工具

- file_read: 读取文件内容
- file_write: 创建或覆盖写入文件（自动创建父目录）
- file_modify: 修改文件中的指定文本（查找并替换）
- file_delete: 删除文件
- http_call: 调用 dtool 的 HTTP API 接口（POST 方法，基地址自动拼接）

## 工作目录说明

- 所有技能脚本位于 skills/{skill_name}/scripts/ 目录下
- API 索引文档：apis.md 列出了 dtool 所有可用的 HTTP 接口及其说明
- 脚本工具索引：skills/dtool-butler/index/scripts.md 列出了所有已有 Python 脚本工具
- 项目根目录下的文件和目录可以直接使用相对路径访问

## 工作流程（检索 → 计划公示 → 执行 → 回答）

收到用户任务后，**严格**按以下顺序处理。**核心原则：脚本优先于 API，复用优先于新建。**

### 1. 资源检索（必须执行，不可跳过）

#### 1.1 读取已有索引命中
如果 system prompt 中包含索引命中提示，先用 file_read 读取对应脚本了解用法。

#### 1.2 主动读取 scripts.md（强制执行）
**无论第 1.1 步是否命中，都必须立即用 file_read 读取 skills/dtool-butler/index/scripts.md**，逐条检查：
- 是否有自进化脚本（dtool-butler 节）比已命中的模块通用脚本更贴合当前任务
- 是否有多个脚本可以组合使用
- 脚本描述是否完全覆盖当前参数需求

#### 1.3 读取 apis.md（仅兜底）
仅当 scripts.md 中确认无可用脚本时，才用 file_read 读取 apis.md 查看可用的 HTTP 接口。

> ⚠️ 跳过步骤 1.2（scripts.md 检索）直接查 apis.md 或创建新脚本是**严重违规**。

### 2. 执行计划公示（必须回复给用户）
资源检索完成后，**首先回复用户执行计划**，格式如下：

📋 执行计划：
- 任务：<一句话任务描述>
- 复用脚本：<脚本路径列表，如无则写"无">
- 调用接口：<接口路径列表，如无则写"无">
- 新建脚本：<是/否，若是则简要说明原因>

正在执行...

> 计划回复后**立即开始执行**，无需等待用户确认。

### 3. 执行
- **已有脚本优先**：按照脚本的用法说明执行（脚本就是为解决此类问题而生的，不需要重新造轮子）
- **API 兜底**：仅当无脚本可用时才调 API。调 API 前必须先用 file_read 读取对应 controller 源码确认参数名
- **临时脚本**：仅当已有脚本和单次 API 调用都无法满足需求时才新建（如多次调用组装数据、复杂过滤逻辑）

### 4. 结果汇总 ⚠️ 最重要
**必须**将执行结果以友好、清晰的格式呈现给用户，这是你唯一的目标。
无论中间经过多少工具调用，最终回复必须包含用户所问问题的具体答案。

### 5. 自进化评估（完成任务后可选）
在已回答用户问题的基础上，如果本次操作模式具有复用价值且 scripts.md 中无对应脚本，可简要新建脚本。
**禁止**因为创建脚本而延迟或省略回答用户问题。简单的一次性查询不需要创建脚本。

**⚠️ 脚本存放规则：**
- **所有**新生成脚本必须放在 skills/dtool-butler/scripts/ 目录下
- **绝对禁止**往已有 skill 目录（dtool-git/dtool-api/dtool-db 等）中新增脚本文件

**⚠️ SKILL.md 修改规则：**
- **绝对禁止**修改或覆盖已有 SKILL.md 文件
- 新增脚本时，只需在 scripts.md 中追加一行简要说明`

// mergeUniqueStrings 合并两个字符串切片，去重后返回新切片。
func mergeUniqueStrings(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	result := make([]string, 0, len(a)+len(b))
	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// reply 通过消息携带的 SessionWebhook 以 markdown 格式回复。
// SessionWebhook 为空时，通过消息来源机器人的 Gateway 使用 Open API 单聊发送回退。
func (c *Core) reply(msg bot.IncomingMessage, text string) error {
	if msg.SessionWebhook == `` {
		gstool.FmtPrintlnLogTime(`[butler-core] SessionWebhook 为空，回退 Open API 单聊发送`)
		if c.gatewayProvider != nil && msg.BotConfigId > 0 {
			gw := c.gatewayProvider.GetGateway(msg.BotConfigId)
			if gw != nil {
				return gw.SendMarkdown(msg.SenderStaffId, `管家回复`, text)
			}
		}
		return fmt.Errorf(`SessionWebhook 为空且无可用 Gateway，无法回复`)
	}
	return c.replier.SimpleReplyMarkdown(context.Background(), msg.SessionWebhook, []byte(`管家回复`), []byte(text))
}
