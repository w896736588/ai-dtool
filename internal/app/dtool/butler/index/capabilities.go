package index

import (
	"strings"
)

// GenerateCapabilitiesIndex 生成 capabilities.md 内容——管家总能力清单。
// 基于管家已知的内置命令、FC 工具、索引能力和会话管理，生成静态描述。
func GenerateCapabilitiesIndex() string {
	var sb strings.Builder
	sb.WriteString("# 管家能力清单\n\n")
	sb.WriteString("dtool-butler 管家的核心能力如下：\n\n")

	sb.WriteString("## 内置命令\n\n")
	sb.WriteString("| 命令 | 说明 |\n|------|------|\n")
	sb.WriteString("| /clean | 清除当前会话历史 |\n")
	sb.WriteString("| /init | 初始化索引文档（step.md、capabilities.md、apis.md） |\n")
	sb.WriteString("| /status | 查询管家与当前会话状态 |\n")
	sb.WriteString("| /help | 显示帮助信息 |\n\n")

	sb.WriteString("## Function Calling 工具\n\n")
	sb.WriteString("| 工具名 | 说明 |\n|--------|------|\n")
	sb.WriteString("| file_read | 读取文件内容 |\n")
	sb.WriteString("| file_write | 创建或覆盖写入文件（自动创建父目录） |\n")
	sb.WriteString("| file_modify | 查找并替换文件中的指定文本 |\n")
	sb.WriteString("| file_delete | 删除文件 |\n")
	sb.WriteString("| http_call | 调用 dtool 的 HTTP API 接口（自动拼接基地址） |\n\n")

	sb.WriteString("## 索引与自进化\n\n")
	sb.WriteString("- step.md：管家自进化积累的可复用操作步骤索引（一行一个步骤文件 + 任务说明）\n")
	sb.WriteString("- capabilities.md：管家总能力清单（本文件）\n")
	sb.WriteString("- apis.md：dtool HTTP 接口索引\n")
	sb.WriteString("- 检索：任务执行前先查 step.md 是否有可复用的步骤文件\n")
	sb.WriteString("- 自进化：任务完成后归档管家分析工具调用过程，总结为 .md 步骤文件，追加到 step.md\n\n")

	sb.WriteString("## 会话管理\n\n")
	sb.WriteString("- 激活态：收到消息后激活，定时器重置\n")
	sb.WriteString("- 休眠巡检：超时无消息自动休眠并通知\n")
	sb.WriteString("- 历史管理：对话存库、新话题自动清历史、溢出提示\n\n")

	sb.WriteString("## 任务路由\n\n")
	sb.WriteString("- 简单任务（文件操作）→ Function Calling 工具循环\n")
	sb.WriteString("- 复杂任务（开发/重构）→ Agent CLI（Claude Code CLI / Codex CLI）\n\n")

	sb.WriteString("## 意图分析\n\n")
	sb.WriteString("- 模糊问题 → 自动追问 2-3 个澄清提问\n")
	sb.WriteString("- 明确意图 → 进入任务执行\n")
	sb.WriteString("- 新话题检测 → 自动清除旧历史\n")
	return sb.String()
}
