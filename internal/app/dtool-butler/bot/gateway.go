package bot

// Gateway 定义机器人网关接口，预留多平台扩展（当前仅钉钉实现）。
type Gateway interface {
	// Start 建立长连接并开始接收消息，非阻塞，内部起 goroutine。
	Start() error
	// Close 关闭连接。
	Close()
	// SendText 主动发送文本消息到指定会话（用于打招呼/休眠通知等无 incoming 的场景）。
	// conversationId 为钉钉会话 ID；P1 阶段若无法主动推送则返回 nil 由调用方决定。
	SendText(conversationId, text string) error
}
