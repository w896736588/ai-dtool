package p_common

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pion/stun"
	"github.com/w896736588/go-tool/gsencrypt"
	"github.com/w896736588/go-tool/gstool"
)

var TBaseClient *TBase

var AesGcmClient *gsencrypt.AesGcm

var TOsClient = gstool.NewGsOs()

type TBase struct {
	StartMillUnix int64
	LogPath       string
}

// GetCombineKey 组装key
func (h *TBase) GetCombineKey(params ...any) string {
	strParamsList := gstool.Array2Str(&params)
	return strings.Join(strParamsList, `#`)
}

// ExplainCombineKey 分解key
func (h *TBase) ExplainCombineKey(uniqueKey string) []string {
	return strings.Split(uniqueKey, `#`)
}

func (h *TBase) GetUnique(prefix string) string {
	h.StartMillUnix += 1
	return fmt.Sprintf(`%s%d`, prefix, h.StartMillUnix)
}

func (h *TBase) GetPublicIPWithSTUN() (string, error) {
	// 1. 创建UDP连接
	conn, err := net.Dial("udp", "stun.l.google.com:19302") // Google公共STUN服务器
	if err != nil {
		return "", fmt.Errorf("创建UDP连接失败: %v", err)
	}
	defer conn.Close()

	// 2. 设置超时
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return "", fmt.Errorf("设置超时失败: %v", err)
	}

	// 3. 创建STUN客户端
	client, err := stun.NewClient(conn)
	if err != nil {
		return "", fmt.Errorf("创建STUN客户端失败: %v", err)
	}
	defer client.Close()

	// 4. 构建STUN请求
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// 5. 处理响应
	var publicIP string
	err = client.Do(message, func(res stun.Event) {
		if res.Error != nil {
			return
		}

		// 解析XOR-MAPPED-ADDRESS属性
		var xorAddr stun.XORMappedAddress
		if err := xorAddr.GetFrom(res.Message); err != nil {
			return
		}
		publicIP = xorAddr.IP.String()
	})

	if err != nil {
		return "", fmt.Errorf("STUN请求失败: %v", err)
	}

	if publicIP == "" {
		return "", fmt.Errorf("未从STUN响应中获取到IP地址")
	}

	return publicIP, nil
}

func (h *TBase) DiffText(text1, text2 string) string {
	// 创建临时文件
	tmpFileName1 := h.GetUnique(`diff`) + `.md`
	tmpFile1 := filepath.Join(h.LogPath, tmpFileName1)

	defer func() {
		_ = gstool.FileDelete(tmpFile1)
	}()

	err := gstool.FileCreate(h.LogPath, tmpFileName1, text1)
	if err != nil {
		return ``
	}

	//defer gstool.FileDelete(tmpFile1)
	tmpFileName2 := h.GetUnique(`diff`) + `.md`
	tmpFile2 := filepath.Join(h.LogPath, tmpFileName2)

	defer func() {
		_ = gstool.FileDelete(tmpFile2)
	}()

	err = gstool.FileCreate(h.LogPath, tmpFileName2, text2)
	if err != nil {
		return ``
	}

	// 执行 git diff --numstat file1 file2
	cmd := exec.Command("git", "diff", "--no-index", "--shortstat", tmpFile1, tmpFile2)
	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")
	stats := ""
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			stats = lines[i]
			break
		}
	}
	return stats

}

// 过滤终端控制符，标准化换行
// 修复：避免过度匹配导致核心文本被删除
func (h *TBase) FilterTerminalChars(msg string) string {
	// 1. 过滤常规 ANSI 颜色/控制符（\x1b[数字;字母 格式）
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	msg = ansiRegex.ReplaceAllString(msg, "")

	// 2. 过滤特殊终端模式控制符（如 \x1b[?2004h/l）
	// 优化：精准匹配 \x1b[? 开头 + 数字 + 字母，避免过度匹配
	cleanRegex := regexp.MustCompile(`\x1b\[\?[0-9]+[a-zA-Z]`)
	msg = cleanRegex.ReplaceAllString(msg, "")

	// 3. 修复：精准过滤 \x1b] 开头的OSC命令（仅匹配命令本身，不吞后续文本）
	// 匹配规则：
	// - \x1b] 开头
	// - [0-2] 标题相关参数
	// - ;? 可选的分隔符
	// - ([^\x07\n\r]*)? 可选的标题内容（仅匹配非换行/非BEL的字符，避免跨行）
	// - \x07? 可选的结束符
	titleRegex := regexp.MustCompile(`\x1b\][0-2];?([^\x07\n\r]*)?\x07?`)
	msg = titleRegex.ReplaceAllString(msg, "")

	// 4. 统一换行格式：保留核心文本的换行逻辑，仅清理杂乱回车
	newlineRegex := regexp.MustCompile(`\r\n|\r`)
	msg = newlineRegex.ReplaceAllString(msg, "\n")
	// 优化：只合并3个及以上连续换行（避免把有效换行也合并），可根据需求调整
	multiNewlineRegex := regexp.MustCompile(`\n{3,}`)
	msg = multiNewlineRegex.ReplaceAllString(msg, "\n")

	// 5. 可选：清理首尾多余的换行（如果需要）
	// msg = strings.Trim(msg, "\n")

	return msg
}

func (h *TBase) GetGitPromptHosts(input string) string {
	// 正则表达式解释：
	// 1. (Username|Password) for ' ：匹配固定前缀
	// 2. http(s)?:// ：匹配 http:// 或 https://（s可选）
	// 3. (?:[^@']+@)? ：非捕获组，匹配可选的用户名@部分
	// 4. ([^']+) ：捕获组，提取核心域名（直到单引号结束）
	re := regexp.MustCompile(`(Username|Password) for 'http(s)?://(?:[^@']+@)?([^']+)`)

	matches := re.FindStringSubmatch(input)
	if len(matches) < 4 {
		return ""
	}

	// 拼接完整地址：协议（http/https） + 核心域名
	protocol := "http"
	if matches[2] == "s" { // matches[2] 是可选的 "s" 字符
		protocol = "https"
	}
	fullURL := fmt.Sprintf("%s://%s", protocol, matches[3])

	return fullURL
}

func (h *TBase) FilterGitPromptHosts(input, promptTitle string) string {
	re := regexp.MustCompile(fmt.Sprintf(`%s 'http(s)?://(?:[^@']+@)?([^']+)`, promptTitle))
	return re.ReplaceAllString(input, ``)
}

func (h *TBase) IsAiCurl(body string) bool {
	return strings.Contains(body, `model`) && strings.Contains(body, `messages`)
}

func (h *TBase) FillConst(replaces map[string]string) {
	replaces[`{current_date_Y/m/d}`] = gstool.TimeNowUnixToString(`Y/m/d`)
	replaces[`{current_date_Y_m_d}`] = gstool.TimeNowUnixToString(`Y_m_d`)
	replaces[`{current_date_Ymd}`] = gstool.TimeNowUnixToString(`Ymd`)
}
