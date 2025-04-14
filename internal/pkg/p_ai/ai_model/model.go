package ai_model

import _struct "dev_tool/base/struct"

type AiModel interface {
	Api(messageList []_struct.Message, tools []_struct.Tool) (string, error)
	MessageList() []_struct.Message
}
