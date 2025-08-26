package base

import (
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gstool"
	"regexp"
	"strings"
)

type TAi struct {
	log *gstool.GsSlog
}

func (h *TAi) Init() {
	h.log = gstool.NewSlog3(Component.Env.LogPath, `ai`)
	_ = h.log.CleanOldLogs(2)
}

// ParseStream 解析流式数据
// 支持纳米AI
// 支持deepseek
func (h *TAi) ParseStream(url, msg string) []byte {
	h.log.Debugf(`%s`, msg)
	msgList := strings.Split(msg, "\n")
	resBytes := make([]byte, 0)
	for _, msgVal := range msgList {
		if !strings.HasPrefix(msgVal, `data: `) {
			continue
		}
		if strings.Contains(url, `/api/assistant/chat`) { //纳米AI
			h.ParseBot(msgVal, &resBytes)
		} else if gstool.SContains(url, []string{`/api/v0/chat/completion`, `/api/GitLab`, `basic`}) { //git日志
			h.ParseDeepseek(msgVal, &resBytes)
		} else if gstool.SContains(url, []string{`/completion/stream`}) { //kimi
			h.ParseKimi(msgVal, &resBytes)
		}
	}
	return resBytes
}

func (h *TAi) ParseStreamJson(url, msg string, sendFunc func(string)) {
	gstool.FmtPrintlnLogTime(`--收到###%v###`, msg)
	re := regexp.MustCompile(`\s{4}.\{`)
	parts := re.Split(msg, -1)
	for _, part := range parts {
		if strings.Trim(part, ` `) == `` {
			continue
		}
		//在按照
		secondList := regexp.MustCompile(`\s{3}.{2}\{`)
		for _, secondPart := range secondList.Split(part, -1) {
			//再按照!{进行切割
			threeList := regexp.MustCompile(`[\x00-\x1F]`)
			for _, threePart := range threeList.Split(secondPart, -1) {
				if threePart == `` {
					continue
				}
				if threePart[0:1] != `{` {
					threePart = threePart[1:]
				}
				gstool.FmtPrintlnLogTime(`解析--%v--`, threePart)
				realMsgObj := _struct.StreamJson{}
				decodeErr := gstool.JsonDecode(threePart, &realMsgObj)
				if decodeErr != nil {

				} else {
					sendFunc(realMsgObj.Block.Text.Content)
				}
			}

		}
	}
}

// ParseBot 纳米AI格式
func (h *TAi) ParseBot(msg string, resBytes *[]byte) {
	msg = gstool.SReplaces(msg, map[string]string{
		`data: `: ``,
	})
	if strings.HasPrefix(msg, `MESSAGEID`) {
		*resBytes = append(*resBytes, []byte(msg+"  \n")...)
	} else {
		if msg == `` { //纳米AI 空可能表示这个消息结束
			*resBytes = append(*resBytes, []byte("\n")...)
		} else {
			*resBytes = append(*resBytes, []byte(msg)...)
		}
	}
}

// ParseKimi kimi格式
func (h *TAi) ParseKimi(msg string, resBytes *[]byte) {
	msg = gstool.SReplaces(msg, map[string]string{
		`data: `: ``,
	})
	data := _struct.Kimi{}
	err := gstool.JsonDecode(msg, &data)
	if err != nil {
		Component.GsLog.Errof(`解析kimi内容失败 --%s--`, msg)
	} else {
		if data.Event == `all_done` {
			*resBytes = append(*resBytes, []byte("\n")...)
			return
		} else if data.Event == `cmpl` { //回复的文字 其实还有其他乱七八糟的事件 这里不管
			*resBytes = append(*resBytes, []byte(data.Text)...)
		}
	}
}

func (h *TAi) ParseDeepseek(msg string, resBytes *[]byte) {
	msg = gstool.SReplaces(msg, map[string]string{
		`data: `: ``,
	})
	if msg == "[DONE]" {
		*resBytes = append(*resBytes, []byte("\n")...)
		return
	}
	data := _struct.StreamData{}
	err := gstool.JsonDecode(msg, &data)
	if err != nil {
		Component.GsLog.Errof(`解析deepseek内容失败 --%s--`, msg)
	} else {
		for _, choice := range data.Choices {
			*resBytes = append(*resBytes, []byte(choice.Delta.Content)...)
		}
	}
}
