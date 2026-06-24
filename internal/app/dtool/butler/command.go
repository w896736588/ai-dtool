package butler

import (
	"dev_tool/internal/app/dtool/butler/index"
	"fmt"
	"strings"
)

// 命令前缀常量
const commandPrefix = `/`

// 内置命令名称常量
const (
	CommandClean  = `clean`  // 清除当前会话历史
	CommandInit   = `init`   // 初始化索引文档
	CommandStatus = `status` // 查询管家状态
	CommandHelp   = `help`   // 显示帮助信息
)

// CommandResult 命令执行结果。
type CommandResult struct {
	Handled bool   // 是否为内置命令
	Text    string // 命令执行后的回复文本
}

// CommandContext 命令执行所需的上下文参数。
type CommandContext struct {
	IndexPath  string // 索引文档目录路径
	SkillsRoot string // skills 目录绝对路径
}

// ParseCommand 解析消息文本是否为内置命令，并执行。
// 返回 CommandResult：Handled=true 表示已处理为命令，Text 为回复内容；
// Handled=false 表示不是命令，应由 AI 处理。
func ParseCommand(text string, sessionManager *SessionManager, history *History, sessionId string, configMaxHistory int, cmdCtx *CommandContext) CommandResult {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, commandPrefix) {
		return CommandResult{Handled: false}
	}
	// 去掉前缀，取命令名
	cmd := strings.ToLower(strings.TrimPrefix(text, commandPrefix))
	// 去掉命令名后面的参数
	cmdParts := strings.SplitN(cmd, ` `, 2)
	cmdName := cmdParts[0]

	switch cmdName {
	case CommandClean:
		return execClean(sessionManager, history, sessionId)
	case CommandInit:
		return execInit(cmdCtx)
	case CommandStatus:
		return execStatus(sessionManager, history, sessionId, configMaxHistory)
	case CommandHelp:
		return execHelp()
	default:
		// 未识别的命令，交由 AI 处理
		return CommandResult{Handled: false}
	}
}

// execClean 执行 /clean 命令：清除当前会话历史。
func execClean(sessionManager *SessionManager, history *History, sessionId string) CommandResult {
	if err := history.CleanBySession(sessionId); err != nil {
		return CommandResult{Handled: true, Text: fmt.Sprintf(`清除历史失败：%s`, err.Error())}
	}
	return CommandResult{Handled: true, Text: `已清除当前会话历史。`}
}

// execInit 执行 /init 命令：扫描 skills/ 生成索引文档。
func execInit(cmdCtx *CommandContext) CommandResult {
	if cmdCtx == nil || cmdCtx.IndexPath == `` {
		return CommandResult{Handled: true, Text: `索引路径未配置，无法初始化。`}
	}
	content, err := index.InitIndex(cmdCtx.SkillsRoot, cmdCtx.IndexPath)
	if err != nil {
		return CommandResult{Handled: true, Text: fmt.Sprintf(`索引初始化失败：%s`, err.Error())}
	}
	lineCount := strings.Count(content, "\n") + 1
	return CommandResult{Handled: true, Text: fmt.Sprintf(`索引初始化完成，已生成 step.md（%d 行）、capabilities.md、apis.md。`, lineCount)}
}

// execStatus 执行 /status 命令：查询管家与当前会话状态。
func execStatus(sessionManager *SessionManager, history *History, sessionId string, configMaxHistory int) CommandResult {
	active := sessionManager.IsActive(sessionId)
	msgCount, _ := history.CountBySession(sessionId)
	statusDesc := `休眠`
	if active {
		statusDesc = `在线`
	}
	return CommandResult{
		Handled: true,
		Text:    fmt.Sprintf(`管家状态：%s\n当前会话消息数：%d / %d`, statusDesc, msgCount, configMaxHistory),
	}
}

// builtinCommandsHelp 返回内置命令的帮助文本，供打招呼语和 /help 命令共用。
func builtinCommandsHelp() string {
	return `📋 内置命令：
- /clean — 清除当前会话历史
- /init — 重新生成索引文档
- /status — 查询管家与当前会话状态
- /help — 显示此帮助信息`
}

// execHelp 执行 /help 命令：显示内置命令帮助。
func execHelp() CommandResult {
	helpText := builtinCommandsHelp() + `\n其他消息将由 AI 管家处理。`
	return CommandResult{Handled: true, Text: helpText}
}
