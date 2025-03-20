package ai

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"dev_tool/internal/pkg/ai/ai_define"
	"dev_tool/internal/pkg/ai/ai_model"
	"dev_tool/internal/pkg/ai/ai_parse"
	"errors"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

func Ai(data map[string]any) (string, []ai_define.Message, error) {
	ai := getAiModel(data[`model`].(string))
	if ai == nil {
		return ``, nil, errors.New(`不支持的model`)
	}
	//model service action 处理
	opList := data[`opList`].([]any)
	retList := make([]string, 0)
	for _, op := range opList {
		data[`op`] = op
		parse := ai_parse.NewParse(data)
		messageList, tools, parseErr := parse.Parse()
		if parseErr != nil {
			return ``, nil, errors.New(parseErr.Error())
		}
		ret, retErr := ai.Api(messageList, tools)
		if retErr != nil {
			retList = append(retList, `执行失败：`+retErr.Error())
			continue
		} else {
			base.Component.GsLog.Debugf("对话开始 结束：\n %s", ret)
			retList = append(retList, ret)
		}
	}
	//其他选项处理
	parse := ai_parse.NewParse(data)
	messageList, tools, parseErr := parse.ParseOtherSet()
	if parseErr != nil {
		return ``, nil, errors.New(parseErr.Error())
	}
	if len(messageList) > 0 {
		ret, retErr := ai.Api(messageList, tools)
		if retErr != nil {
			retList = append(retList, `执行失败：`+retErr.Error())
		} else {
			base.Component.GsLog.Debugf("对话开始 结束：\n %s", ret)
			retList = append(retList, ret)
		}
	}

	return strings.Join(retList, "\n"), ai.MessageList(), nil
}

func getAiModel(model string) ai_model.AiModel {
	switch model {
	case `qwen2.5-coder-32b-instruct`:
		ai := ai_model.NewBailian(model, `sk-938dc32c6e394fe089e64aac7ee6443f`, true, func(s string, err error) {
			if err != nil {
				sendErr := base.Component.TSse.Send(define.SseAiCode, `执行失败:`+err.Error())
				if sendErr != nil {
					gstool.FmtPrintlnLogTime(`发送0#code失败 %s`, sendErr.Error())
				}
			} else {
				sendErr := base.Component.TSse.Send(define.SseAiCode, s)
				if sendErr != nil {
					gstool.FmtPrintlnLogTime(`发送0#code失败 %s`, sendErr.Error())
				}
			}
		})
		return ai
	default:
		return nil
	}
}
