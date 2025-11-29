package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
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
		} else if gstool.SContains(url, []string{`/api/v0/chat/completion`, `basic`}) {
			h.ParseDeepseek(msgVal, &resBytes)
		} else if gstool.SContains(url, []string{`/completion/stream`}) { //kimi
			h.ParseKimi(msgVal, &resBytes)
		} else if gstool.SContains(url, []string{`/api/GitLab`}) { //gitlab 自定义接口
			h.ParseBaseStruct(msgVal, &resBytes)
		}
	}
	return resBytes
}

// ParseBaseStruct 自定义json 格式 {"sse_client_id":"gitlab","data":"获取red_packet_send_servicecommit 共：20条  \n","type":"msg"}
func (h *TAi) ParseBaseStruct(msg string, resBytes *[]byte) {
	msg = gstool.SReplaces(msg, map[string]string{
		`data: `: ``,
	})
	data := define.SseData{}
	err := gstool.JsonDecode(msg, &data)
	if err != nil {
		Component.GsLog.Errof(`解析内容失败 --%s--`, msg)
		return
	}
	*resBytes = append(*resBytes, []byte(cast.ToString(data.Data))...)
}

func (h *TAi) ParseStreamJson(url, msg string, sendFunc func(string)) {
	Component.TVariable.Log.Debugf(`step one 收到 -%v-`, msg)
	if strings.Index(msg, `{{"op"`) >= 0 { //奇怪的格式
		msg = strings.Replace(msg, `{{"op"`, `{"op"`, 1)
		Component.TVariable.Log.Debugf(`step two 发现奇怪格式 -%v-`, msg)
	}
	jsonLists := gstool.JsonParseFromStr([]byte(msg))
	for _, part := range jsonLists {
		if part == `` {
			continue
		}
		Component.TVariable.Log.Debugf(`step three 解析出json-%v-`, part)
		realMsgObj := _struct.StreamJson{}
		decodeErr := gstool.JsonDecode(part, &realMsgObj)
		if decodeErr != nil {
			gstool.FmtPrintlnLogTime(`解析失败 %s`, decodeErr.Error())
		} else if realMsgObj.Block.Text.Content != `` {
			sendFunc(realMsgObj.Block.Text.Content)
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
	if msg == define.SseDown {
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
