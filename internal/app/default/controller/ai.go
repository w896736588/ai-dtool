package controller

import (
	"dev_tool/base"
	"dev_tool/internal/pkg/ai/ai_bailian"
	"dev_tool/internal/pkg/ai/ai_define"
	"dev_tool/internal/pkg/ai/ai_parse"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"strings"
)

func AiRun(c *gin.Context) {
	data, err := getAiComponent(c)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gstool.FmtPrintlnLogTime(`data %s`, gstool.JsonEncode(data))
	opList := data[`opList`].([]any)
	retList := make([]string, 0)
	for _, op := range opList {
		data[`op`] = op
		parse := ai_parse.NewParse(data)
		messageList, tools, parseErr := parse.Parse()
		if parseErr != nil {
			gsgin.GinResponseError(c, parseErr.Error(), nil)
			return
		}
		var ret = ``
		var retErr error = nil
		model := data[`model`].(string)
		switch model {
		case `qwen2.5-coder-32b-instruct`:
			ai := ai_bailian.NewBailian(model, `sk-938dc32c6e394fe089e64aac7ee6443f`, true, func(s string, err error) {
				gstool.FmtPrintlnLogTime(`收到消息 %s %v`, s, err)
				if err != nil {
					base.Component.TSocket.SendMsg(`code`, `执行失败:`+err.Error())
				} else {
					s = gstool.StringReplaces(s, map[string]string{
						`data: `: ``,
					})
					streamData := ai_define.StreamData{}
					_ = gstool.JsonDecode(s, &streamData)
					for _, val := range streamData.Choices {
						base.Component.TSocket.SendMsgReal(`0#code`, val.Delta.Content)
					}
				}
			})
			ret, retErr = ai.Api(messageList, tools)
		}
		if retErr != nil {
			retList = append(retList, `执行失败：`+retErr.Error())
			continue
		} else {
			retList = append(retList, ret)
		}
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`ret`: strings.Join(retList, "\n"),
	})
}

func getAiComponent(c *gin.Context) (map[string]interface{}, error) {
	reqMap := make(map[string]interface{})
	err := gsgin.GinPostBody(c, &reqMap)
	if err != nil {
		return nil, err
	}
	return reqMap, nil
}
