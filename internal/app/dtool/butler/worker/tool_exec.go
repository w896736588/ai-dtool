package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/w896736588/go-tool/gshttp"
)

const (
	// toolHttpTimeoutSeconds HTTP 调用超时秒数，防止 http_call 工具永久阻塞 consumeLoop。
	toolHttpTimeoutSeconds = 30
	// toolScriptTimeoutSeconds 脚本执行超时秒数，防止 run_script 工具永久阻塞 consumeLoop。
	toolScriptTimeoutSeconds = 60
)

// skillsRoot 技能目录绝对路径，由外部在启动时设置，供路径解析使用。
var skillsRoot string

// dtoolBaseURL dtool API 基地址（如 http://localhost:17170），由外部在启动时设置。
var dtoolBaseURL string

// SetSkillsRoot 设置 skills 目录绝对路径，供文件工具解析相对路径时使用。
func SetSkillsRoot(root string) {
	skillsRoot = root
}

// SetDtoolBaseURL 设置 dtool API 基地址，供 http_call 工具拼接完整 URL。
func SetDtoolBaseURL(baseURL string) {
	dtoolBaseURL = baseURL
}

// argsToStringMap 将 JSON 字节解析为 map[string]string，
// 兼容值为数字等非字符串类型的情况——自动转为字符串。
func argsToStringMap(argumentsJSON string) (map[string]string, error) {
	raw := make(map[string]interface{})
	if err := json.Unmarshal([]byte(argumentsJSON), &raw); err != nil {
		return nil, err
	}
	args := make(map[string]string, len(raw))
	for k, v := range raw {
		switch val := v.(type) {
		case string:
			args[k] = val
		case float64:
			// JSON 数字解析为 float64，整数场景去掉小数点
			if val == float64(int64(val)) {
				args[k] = fmt.Sprintf("%d", int64(val))
			} else {
				args[k] = fmt.Sprintf("%f", val)
			}
		case bool:
			args[k] = fmt.Sprintf("%t", val)
		case nil:
			args[k] = ""
		default:
			args[k] = fmt.Sprintf("%v", val)
		}
	}
	return args, nil
}

// ExecuteTool 执行指定的工具调用，返回执行结果文本。
func ExecuteTool(name string, argumentsJSON string) string {
	args, err := argsToStringMap(argumentsJSON)
	if err != nil {
		return fmt.Sprintf(`参数解析失败：%s`, err.Error())
	}
	switch name {
	case ToolFileRead:
		return execFileRead(args[`path`])
	case ToolFileWrite:
		return `【禁止】file_write 工具在执行阶段不可用，文件创建由归档管家独立处理`
	case ToolFileModify:
		return `【禁止】file_modify 工具在执行阶段不可用，文件修改由归档管家独立处理`
	case ToolFileDelete:
		return `【禁止】file_delete 工具在执行阶段不可用，文件删除由归档管家独立处理`
	case ToolHttpCall:
		return execHttpCall(args[`path`], args[`body`])
	case ToolRunScript:
		return execRunScript(args[`path`], args[`args`], args[`timeout`])
	case ToolAskUser:
		return execAskUser(args[`question`], args[`options`], args[`reason`])
	default:
		return fmt.Sprintf(`未知工具：%s`, name)
	}
}

// resolvePath 解析文件路径：如果是相对路径且直接读取失败，依次在 skills/dtool-butler/index/、skills/dtool-butler/step/ 和 skillsRoot 下查找。
// 优先级：直接路径 > skills/dtool-butler/index/ (索引文件) > skills/dtool-butler/step/ (自进化步骤文件) > skills/*/step/ (内置步骤文件)
func resolvePath(path string) (string, error) {
	if path == `` {
		return ``, fmt.Errorf(`文件路径不能为空`)
	}
	// 绝对路径直接使用
	if filepath.IsAbs(path) {
		return path, nil
	}
	// 先尝试原路径
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if skillsRoot != `` {
		// 1. 优先在 skills/dtool-butler/index/ 下查找（索引文件：apis.md, step.md 等）
		indexCandidate := filepath.Join(skillsRoot, `dtool-butler`, `index`, path)
		if _, err := os.Stat(indexCandidate); err == nil {
			return indexCandidate, nil
		}
		// 2. 在 skills/dtool-butler/step/ 下查找（自进化生成的步骤文件）
		evolvedCandidate := filepath.Join(skillsRoot, `dtool-butler`, `step`, path)
		if _, err := os.Stat(evolvedCandidate); err == nil {
			return evolvedCandidate, nil
		}
		// 3. 回退：在 skillsRoot 下全面查找
		candidate := filepath.Join(skillsRoot, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return path, nil
}

// execFileRead 读取文件内容。相对路径会自动在 skills 目录下查找。
func execFileRead(path string) string {
	resolved, err := resolvePath(path)
	if err != nil {
		return fmt.Sprintf(`读取文件失败：%s`, err.Error())
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return fmt.Sprintf(`读取文件失败：%s`, err.Error())
	}
	return string(data)
}

// execFileWrite 写入文件内容，自动创建父目录。
func execFileWrite(path, content string) string {
	if path == `` {
		return `错误：文件路径不能为空`
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf(`创建目录失败：%s`, err.Error())
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Sprintf(`写入文件失败：%s`, err.Error())
	}
	return `文件写入成功`
}

// execFileModify 查找并替换文件中的文本（仅替换第一个匹配项）。
// 相对路径会自动在 skills 目录下查找。
func execFileModify(path, search, replacement string) string {
	if path == `` {
		return `错误：文件路径不能为空`
	}
	if search == `` {
		return `错误：搜索文本不能为空`
	}
	resolved, err := resolvePath(path)
	if err != nil {
		return fmt.Sprintf(`读取文件失败：%s`, err.Error())
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return fmt.Sprintf(`读取文件失败：%s`, err.Error())
	}
	content := string(data)
	if !strings.Contains(content, search) {
		return `未找到匹配的文本`
	}
	newContent := strings.Replace(content, search, replacement, 1)
	if err := os.WriteFile(resolved, []byte(newContent), 0644); err != nil {
		return fmt.Sprintf(`写入文件失败：%s`, err.Error())
	}
	return `文件修改成功`
}

// execFileDelete 删除文件。相对路径会自动在 skills 目录下查找。
func execFileDelete(path string) string {
	if path == `` {
		return `错误：文件路径不能为空`
	}
	resolved, err := resolvePath(path)
	if err != nil {
		return fmt.Sprintf(`删除文件失败：%s`, err.Error())
	}
	if err := os.Remove(resolved); err != nil {
		return fmt.Sprintf(`删除文件失败：%s`, err.Error())
	}
	return `文件删除成功`
}

// execRunScript 执行本地 Python 脚本并返回 stdout+stderr。
// 脚本路径会自动在 skillsRoot 下查找。
// 支持通过 timeoutStr 参数自定义超时秒数，未指定时使用默认 60s。
func execRunScript(path, argsStr, timeoutStr string) string {
	if path == `` {
		return `错误：脚本路径不能为空`
	}
	// 解析路径
	resolved, err := resolvePath(path)
	if err != nil {
		return fmt.Sprintf(`脚本路径解析失败：%s`, err.Error())
	}
	// 检查文件存在
	if _, err := os.Stat(resolved); err != nil {
		return fmt.Sprintf(`脚本不存在：%s`, resolved)
	}
	// 解析超时：优先使用传入参数，否则用默认值
	timeout := toolScriptTimeoutSeconds
	if timeoutStr != `` {
		if parsed, parseErr := fmt.Sscanf(timeoutStr, "%d", &timeout); parseErr != nil || parsed != 1 || timeout <= 0 {
			timeout = toolScriptTimeoutSeconds
		}
	}
	// 构建命令参数
	cmdArgs := []string{resolved}
	if argsStr != `` {
		cmdArgs = append(cmdArgs, strings.Fields(argsStr)...)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, `python`, cmdArgs...)
	// 设置 UTF-8 环境变量，解决 Windows 上 Python 管道输出 GBK 编码导致中文乱码的问题
	cmd.Env = append(os.Environ(),
		`PYTHONIOENCODING=utf-8`,
		`PYTHONUTF8=1`,
	)
	output, err := cmd.CombinedOutput()
	result := string(output)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf(`脚本执行超时（%d 秒）：%s`, timeout, truncateForLog(result, 2000))
		}
		return fmt.Sprintf(`脚本执行失败：%s\n输出：%s`, err.Error(), truncateForLog(result, 2000))
	}
	if len(result) > 3000 {
		result = result[:3000] + `\n...(输出已截断)`
	}
	return result
}

// execAskUser 向用户发起确认问题。
// 返回特殊格式的标记字符串，供 FC 循环检测并暂停等待用户回复。
func execAskUser(question, options, reason string) string {
	if question == `` {
		return `错误：确认问题不能为空`
	}
	result := map[string]string{
		`marker`:   AskUserMarker,
		`question`: question,
		`options`:  options,
		`reason`:   reason,
	}
	b, _ := json.Marshal(result)
	return string(b)
}

// execHttpCall 调用 dtool 的 HTTP API 接口。
// 自动拼接 dtoolBaseURL 与传入的 path，通过 gshttp 发起 POST 请求（带超时保护）并返回响应文本。
func execHttpCall(path, body string) string {
	if path == `` {
		return `错误：API 路径不能为空`
	}
	if dtoolBaseURL == `` {
		return `错误：dtool API 基地址未配置，无法发起 HTTP 调用`
	}
	// 确保 path 以 / 开头
	if !strings.HasPrefix(path, `/`) {
		path = `/` + path
	}
	fullURL := strings.TrimRight(dtoolBaseURL, `/`) + path

	respBytes, err := gshttp.PostJson(fullURL).
		BodyStr(body).
		Request(toolHttpTimeoutSeconds).
		Result()
	if err != nil {
		return fmt.Sprintf(`HTTP 请求失败：%s`, err.Error())
	}
	result := string(respBytes)
	// 截断过长响应
	if len(result) > 3000 {
		result = result[:3000] + `\n...(响应已截断)`
	}
	return fmt.Sprintf(`HTTP 200 %s → %s`, fullURL, result)
}
