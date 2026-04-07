package p_define

const (
	SseContentTypeMsg          = `msg`          //消息
	SseContentTypeErrorList    = `error_list`   //错误列表
	SseContentTypeFilterList   = `filter_list`  //拦截数量
	SseContentTypeFilter       = `filter`       //拦截
	SseContentTypeError        = `error`        //错误
	SseContentTypeConnections  = `connections`  //Shell连接状态
)

type SseData struct {
	SseDistributeId string `json:"sse_distribute_id"` //具体接收业务的id 因为公用一个链接
	Data            any    `json:"data"`              //发送的数据
	Type            string `json:"type"`              //数据类型
}
