package _struct

type CurlResultParse struct {
	Label         string     `json:"label"`          //名字
	Uri           string     `json:"uri"`            //地址 如果是用于playwright 那么表示拦截的uri，如果是自定义脚本那么表示需要请求的url完整地址
	IsStream      int        `json:"is_stream"`      //1流式接收
	ReceiveSignal string     `json:"receive_signal"` //流式接收时按照字符串分割
	ReceiveRegex  string     `json:"receive_regex"`  //流式接收时按照正则分割
	ContentType   string     `json:"content_type"`   //请求的类型 适用于POST
	Method        string     `json:"method"`         //请求的方式POST GET
	TakeJsons     []struct { //从结果中提取json
		Take string `json:"take"` //res.data.token例如
	} `json:"take_jsons"`
	Retry       int `json:"retry"`        //尝试多少次
	RetrySecond int `json:"retry_second"` //每次间隔多少秒
}
