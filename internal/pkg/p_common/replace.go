package p_common

import (
	"dev_tool/internal/pkg/p_sse"
	"fmt"
	"regexp"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// Replace 替换变量
func Replace(data string, replaceList map[string]string) string {
	//处理特殊情况
	for replaceKey, replaceVal := range replaceList {
		// 替换键可能包含正则元字符，先转义后再参与 `%数字` 模式匹配，避免 panic。
		//取模
		matchSubList := gstool.RegexMatchSubString(data, regexp.QuoteMeta(replaceKey)+`%(\d+)`)
		if len(matchSubList) >= 2 {
			data = gstool.SReplaces(data, map[string]string{
				matchSubList[0]: cast.ToString(cast.ToInt64(replaceVal) % cast.ToInt64(matchSubList[1])),
			})
		}
	}
	data = gstool.SReplaces(data, replaceList)
	return data
}

type StructOpenAiBody struct {
	Temperature float64 `json:"temperature"`
	Model       string  `json:"model"`
	Messages    []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

func ReplaceOpenAiBody(body string, replaceList map[string]string, sse *p_sse.SseShell) (string, bool, error) {
	if TBaseClient.IsAiCurl(body) {
		sse.Send(`openai格式curl请求`)
		openAiBody := StructOpenAiBody{}
		err := gstool.JsonDecode(body, &openAiBody)
		if err != nil {
			sse.Send(`解析openai格式失败 ` + body + "\n")
			return ``, true, err
		}
		for messageIndex, message := range openAiBody.Messages {
			openAiBody.Messages[messageIndex].Content = Replace(message.Content, replaceList)
		}
		body = gstool.JsonEncode(openAiBody)
		return body, true, nil
	}
	return body, false, nil
}

type OpenAiResult struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
		TotalTokens      int64 `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Index   int64 `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// 提取openapi结果中的消息
func ExtractOpenAiMessage(result string) string {
	openAiResult := OpenAiResult{}
	err := gstool.JsonDecode(result, &openAiResult)
	if err != nil {
		return fmt.Sprintf(`提取结果失败,from->%s,err->%s`, result, err.Error())
	}
	if len(openAiResult.Choices) >= 1 {
		return openAiResult.Choices[0].Message.Content
	}
	return ``
}
