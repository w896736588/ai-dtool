package base

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
)

type TMarkDown struct {
}

func (h *TMarkDown) Code(str, lang string) string {
	return fmt.Sprintf("```%s\n%s\n%s\n```", lang, lang, str)
}

func (h *TMarkDown) Json(data any) string {
	str := gstool.JsonFormat(data)
	return fmt.Sprintf("```%s\n%s\n```", `json`, str)
}

func (h *TMarkDown) BlockQuote(str string) string {
	return fmt.Sprintf("> %s", str)
}
