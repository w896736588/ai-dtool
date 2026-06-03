package p_curl

import (
	"dev_tool/internal/app/dtool/define"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/w896736588/go-tool/gstool"
)

type KeyValue struct {
	Description string `json:"description"` //描述
	Field       string `json:"field"`       //字段key
	Type        string `json:"type"`        //字段类型 string int  float bool file
	Value       string `json:"value"`       //字段值，如果是文件则存储文件地址,如果是二进制内容则空着
}

const (
	FieldTypeString = "string"
	FieldTypeInt    = "int"
	FieldTypeFloat  = "float"
	FieldTypeBool   = "bool"
	FieldTypeFile   = "file"
)

// CurlStruct 表示解析后的 curl 命令结构
type CurlStruct struct {
	Method      string            `json:"method"`       //POST GET
	Url         string            `json:"url"`          //域名部分 例如 baidu.com
	Protocol    string            `json:"protocol"`     //协议 http https
	ContentType string            `json:"content_type"` //请求头Content-Type application/x-www-form-urlencoded application/json multipart/form-data
	Headers     map[string]string `json:"headers"`      //请求头
	QueryParams []KeyValue        `json:"query_params"` //url的请求参数
	BodyForm    []KeyValue        `json:"body_form"`    // 请求数据 当content-type为application/x-www-form-urlencoded或者multipart/form-data
	BodyJson    string            `json:"body_json"`    // 请求数据当content-type为application/json或raw时
	BodyRaw     string            `json:"body_raw"`     // 请求数据当content-type为text/plain或raw时
}

type ParseCurl struct {
	CurlStruct CurlStruct
	Curl       string
}

// NewParseCurl 解析curl命令字符串
func NewParseCurl(curl string) ParseCurl {
	return ParseCurl{
		CurlStruct: CurlStruct{
			Headers:     map[string]string{},
			BodyForm:    []KeyValue{},
			QueryParams: []KeyValue{},
		},
		Curl: curl,
	}
}

func (h *ParseCurl) Parse() error {
	lines := strings.Split(h.Curl, "\n")
	for _, line := range lines {
		line = strings.TrimLeft(line, ` `)
		//解析地址
		if strings.HasPrefix(line, `curl`) && h.CurlStruct.Url == `` {
			h.GetHostScheme(line)
			continue
		}
		//解析header
		if strings.HasPrefix(line, `-H`) {
			h.GetH(line)
			//通过header解析content-type
			h.contentType(`Content-Type`)
			h.contentType(`content-type`)
			continue
		}
		//解析header
		if strings.HasPrefix(line, `-b`) {
			h.GetB(line)
			continue
		}
		//解析header
		if strings.HasPrefix(line, `--header`) {
			h.GetHeader(line)
			//通过header解析content-type
			h.contentType(`Content-Type`)
			h.contentType(`content-type`)
			continue
		}
		//解析 data
		if strings.HasPrefix(line, `--data-raw`) {
			if h.CurlStruct.ContentType == define.ContentTypeJson {
				h.GetDataRaw(line)
			} else if h.CurlStruct.ContentType == define.ContentTypeForm {
				h.GetDataRawUrls(line)
			} else {
				h.GetDataRawForm(line)
			}
			continue
		}
		//解析 form表单
		if strings.Contains(line, `--form`) {
			h.GetDataForm(line)
			continue
		}
	}
	if h.CurlStruct.Method == `` {
		h.CurlStruct.Method = http.MethodGet
	}
	return nil
}

func (h *ParseCurl) GetDataForm(sLine string) {
	if strings.Contains(sLine, "=@") {
		line := strings.TrimPrefix(sLine, "\t--form '")
		line = strings.TrimSuffix(line, "\"' \\")
		lineParams := strings.Split(line, "=@\"")
		if len(lineParams) > 1 {
			keyVal := KeyValue{}
			keyVal.Field = lineParams[0]
			keyVal.Value = lineParams[1]
			keyVal.Type = FieldTypeFile
			h.CurlStruct.BodyForm = append(h.CurlStruct.BodyForm, keyVal)
		} else {
			gstool.FmtPrintlnLogTime(`解析文件错误 %s`, sLine)
		}
	} else {
		line := strings.TrimPrefix(sLine, "\t--form '")
		line = strings.TrimSuffix(line, "\"' \\")
		lineParams := strings.Split(line, "=\"")
		if len(lineParams) > 1 {
			keyVal := KeyValue{}
			keyVal.Field = lineParams[0]
			keyVal.Value = lineParams[1]
			keyVal.Type = FieldTypeString
			h.CurlStruct.BodyForm = append(h.CurlStruct.BodyForm, keyVal)
		} else {
			gstool.FmtPrintlnLogTime(`解析form属性错误 %s`, sLine)
		}
	}
}

func (h *ParseCurl) GetDataRaw(line string) {
	line = strings.TrimPrefix(line, "--data-raw '")
	line = strings.TrimSuffix(line, "' \\")
	h.CurlStruct.BodyJson = line
}

func (h *ParseCurl) GetDataRawUrls(line string) {
	line = strings.TrimPrefix(line, "--data-raw '")
	line = strings.TrimSuffix(line, "' \\")
	values, err := url.ParseQuery(line)
	if err != nil {
		gstool.FmtPrintlnLogTime(`解析url参数错误 %s`, line)
	} else {
		for key, value := range values {
			keyVal := KeyValue{}
			keyVal.Field = key
			keyVal.Value = value[0]
			keyVal.Type = FieldTypeString
			h.CurlStruct.BodyForm = append(h.CurlStruct.BodyForm, keyVal)
		}
	}
}

func (h *ParseCurl) GetDataRawForm(line string) {
	params := strings.Split(line, `\r\n`)
	for index, param := range params {
		if strings.Contains(param, `Content-Disposition`) {
			paramChilds := strings.Split(param, `;`)
			keyVal := KeyValue{}
			for _, paramChild := range paramChilds {
				if strings.HasPrefix(strings.TrimLeft(paramChild, ` `), `name`) { //字段名
					matchs := gstool.RegexMatchSubString(paramChild, `name="([^"]*)"`)
					if len(matchs) > 1 {
						keyVal.Field = matchs[1]
					}
				}
				if strings.Contains(paramChild, `filename`) { //文件名
					matchs := gstool.RegexMatchSubString(paramChild, `filename="([^"]*)"`)
					if len(matchs) > 1 {
						keyVal.Value = matchs[1]
					}
					keyVal.Type = FieldTypeFile
				}
			}
			//如果没有找到文件 那么往下两行是值
			if keyVal.Type == `` {
				keyVal.Type = FieldTypeString
				if len(params) > index+2 {
					keyVal.Value = params[index+2]
				}
			}
			h.CurlStruct.BodyForm = append(h.CurlStruct.BodyForm, keyVal)
		}
	}
}

func (h *ParseCurl) GetH(line string) {
	re := regexp.MustCompile(`-H\s+'([^:]+):\s*([^']+)'`)
	line = strings.TrimSpace(line)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 3 {
		key := strings.TrimSpace(matches[1])
		value := strings.TrimSpace(matches[2])
		h.CurlStruct.Headers[key] = value
	}
}

func (h *ParseCurl) GetB(line string) {
	line = strings.TrimPrefix(line, "-b '")
	line = strings.TrimSuffix(line, "'")
	h.CurlStruct.Headers[`Cookie`] = line
}

func (h *ParseCurl) GetHeader(line string) {
	re := regexp.MustCompile(`header\s+'([^:]+):\s*([^']+)'`)
	line = strings.TrimSpace(line)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 3 {
		key := strings.TrimSpace(matches[1])
		value := strings.TrimSpace(matches[2])
		h.CurlStruct.Headers[key] = value
	}
}

func (h *ParseCurl) contentType(contentTypeKey string) {
	if _, ok := h.CurlStruct.Headers[contentTypeKey]; ok {
		if strings.Contains(h.CurlStruct.Headers[contentTypeKey], define.ContentTypeForm) {
			h.CurlStruct.Headers[`Content-Type`] = define.ContentTypeForm
			h.CurlStruct.Method = http.MethodPost
			h.CurlStruct.ContentType = define.ContentTypeForm
		} else if strings.Contains(h.CurlStruct.Headers[contentTypeKey], define.ContentTypeJson) {
			h.CurlStruct.Headers[`Content-Type`] = define.ContentTypeJson
			h.CurlStruct.Method = http.MethodPost
			h.CurlStruct.ContentType = define.ContentTypeJson
		} else if strings.Contains(h.CurlStruct.Headers[contentTypeKey], define.ContentTypeMultiForm) {
			h.CurlStruct.Headers[`Content-Type`] = define.ContentTypeMultiForm
			h.CurlStruct.Method = http.MethodPost
			h.CurlStruct.ContentType = define.ContentTypeMultiForm
		} else if strings.Contains(h.CurlStruct.Headers[contentTypeKey], define.ContentTypeRaw) {
			h.CurlStruct.Headers[`Content-Type`] = define.ContentTypeRaw
			h.CurlStruct.ContentType = define.ContentTypeRaw
			h.CurlStruct.Method = http.MethodPost
		} else {
			h.CurlStruct.Method = http.MethodGet
		}
	}
}

func (h *ParseCurl) GetHostScheme(line string) {
	if strings.Contains(line, `POST`) {
		h.CurlStruct.Method = `POST`
	} else if strings.Contains(line, `GET`) {
		h.CurlStruct.Method = `GET`
	}
	re := regexp.MustCompile(`https?://[^\s"'\\]+`)
	h.CurlStruct.Protocol, h.CurlStruct.Url = gstool.URLGetBase(re.FindString(line))
	params, _ := gstool.UrlParseParams(re.FindString(line))
	if len(params) > 0 {
		for _, param := range params {
			h.CurlStruct.QueryParams = append(h.CurlStruct.QueryParams, KeyValue{
				Field: param[`key`],
				Type:  FieldTypeString,
				Value: param[`value`],
			})
		}
	}
}
