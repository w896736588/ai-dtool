package define

const (
	SseAiCode       = `ai_code`    //废弃
	SseIdDistribute = `distribute` //固定全局唯一
)

var SseClientIds []string

func RegisterDistributeSseId(sseClientId string) {
	if SseClientIds == nil {
		SseClientIds = make([]string, 0)
	}
	SseClientIds = append(SseClientIds, sseClientId)
}

func UnRegisterDistributeSseId(sseClientId string) {
	if SseClientIds == nil {
		return
	}
	for k, v := range SseClientIds {
		if v == sseClientId {
			SseClientIds = append(SseClientIds[:k], SseClientIds[k+1:]...)
			break
		}
	}
}

const (
	SseEventClean = `[CLEAN]`                   //清除前端的数据
	SseEventLogin = `[LOGIN_USERNAME_PASSWORD]` //通知前端弹窗输入账号密码
	SseDown       = `[DONE]`                    //前端换个行
	SseConnect    = `[CONNECT]`                 //链接已建立
)

type SseData struct {
	SseClientId string `json:"sse_client_id"`
	Data        any    `json:"data"`
	Type        string `json:"type"`
}

type SseEvent string

const (
	SseContentTypeMsg       = `msg`        //消息
	SseContentTypeErrorList = `error_list` //错误列表
	SseContentTypeError     = `error`      //错误
)
