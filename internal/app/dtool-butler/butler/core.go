package butler

import (
	"dev_tool/internal/app/dtool-butler/bot"
	"dev_tool/internal/app/dtool-butler/define"
	"dev_tool/internal/app/dtool/common"
	"fmt"
	"time"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/w896736588/go-tool/gstool"
)

// Core 管家核心，负责消息消费、激活态管理、回复、休眠巡检。
type Core struct {
	db         *common.CSqlite
	config     *define.ButlerConfigItem
	role       *define.RoleItem
	botGateway bot.Gateway
	history    *History
	sessions   *SessionManager
	msgChan    <-chan bot.IncomingMessage
	replier    *chatbot.ChatbotReplier
	stopCh     chan struct{}
}

// NewCore 创建管家核心。msgChan 为机器人网关投递的消息通道。
func NewCore(
	db *common.CSqlite,
	config *define.ButlerConfigItem,
	role *define.RoleItem,
	botGateway bot.Gateway,
	msgChan <-chan bot.IncomingMessage,
) *Core {
	timeout := time.Duration(config.ActiveTimeoutMinutes) * time.Minute
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	return &Core{
		db:         db,
		config:     config,
		role:       role,
		botGateway: botGateway,
		history:    NewHistory(db),
		sessions:   NewSessionManager(timeout),
		msgChan:    msgChan,
		replier:    chatbot.NewChatbotReplier(),
		stopCh:     make(chan struct{}),
	}
}

// Start 启动管家主循环：发打招呼 → 消费消息 → 定时巡检休眠。非阻塞。
func (c *Core) Start() {
	// 启动打招呼
	c.sendGreeting()
	// 启动消息消费循环
	go c.consumeLoop()
	// 启动休眠巡检（每 1min）
	go c.timeoutLoop()
	gstool.FmtPrintlnLogTime(`[butler-core] 管家已启动，激活态超时=%v`, time.Duration(c.config.ActiveTimeoutMinutes)*time.Minute)
}

// Stop 停止管家主循环。
func (c *Core) Stop() {
	close(c.stopCh)
}

// sendGreeting 启动时发送打招呼消息。钉钉 Stream 模式下无目标会话时通过 webhook 推送。
func (c *Core) sendGreeting() {
	if c.role == nil || c.role.InitGreeting == `` {
		gstool.FmtPrintlnLogTime(`[butler-core] 角色未配置打招呼语，跳过`)
		return
	}
	// P1 阶段：通过群机器人 webhook 主动推送打招呼
	if err := c.botGateway.SendText(``, c.role.InitGreeting); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 打招呼发送失败 %s`, err.Error())
	}
	gstool.FmtPrintlnLogTime(`[butler-core] 已发送打招呼`)
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
				gstool.FmtPrintlnLogTime(`[butler-core] 会话超时休眠 %s`, conversationId)
				// 休眠通知通过 webhook 推送（无 incoming 的 SessionWebhook 可用）
				_ = c.botGateway.SendText(conversationId, `我已休眠，需要时随时叫我。`)
			}
		}
	}
}

// handleMessage 处理单条消息：激活会话 → 存历史 → 回复。
// P1 阶段：固定回复"已收到：{消息}"，不接 AI。
func (c *Core) handleMessage(msg bot.IncomingMessage) {
	// 激活会话（刷新最后活跃时间）
	justActivated := c.sessions.Activate(msg.ConversationId)
	if justActivated {
		gstool.FmtPrintlnLogTime(`[butler-core] 会话已激活 %s`, msg.ConversationId)
	}
	// 存历史（用户消息）
	if err := c.history.Append(msg.ConversationId, define.RoleUser, msg.Text); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 存用户消息失败 %s`, err.Error())
	}
	// P1 固定回复
	replyText := fmt.Sprintf(`已收到：%s`, msg.Text)
	if err := c.reply(msg, replyText); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 回复失败 %s`, err.Error())
		return
	}
	// 存历史（管家回复）
	if err := c.history.Append(msg.ConversationId, define.RoleAssistant, replyText); err != nil {
		gstool.FmtPrintlnLogTime(`[butler-core] 存管家回复失败 %s`, err.Error())
	}
}

// reply 通过消息携带的 SessionWebhook 回复。
func (c *Core) reply(msg bot.IncomingMessage, text string) error {
	if msg.SessionWebhook == `` {
		gstool.FmtPrintlnLogTime(`[butler-core] SessionWebhook 为空，回退 webhook 发送`)
		return c.botGateway.SendText(msg.ConversationId, text)
	}
	return c.replier.SimpleReplyText(nil, msg.SessionWebhook, []byte(text))
}
