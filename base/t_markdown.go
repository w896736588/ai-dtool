package base

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
)

type TMarkDown struct {
}

func (h *TMarkDown) Code(str, lang string) string {
	tip := `#`
	if lang == `sql` {
		tip = `-`
	} else if lang == `json` {
		tip = `//`
	}
	return fmt.Sprintf("```%s\n%s%s\n%s\n```", lang, tip, lang, str)
}

func (h *TMarkDown) Json(data any) string {
	str := gstool.JsonFormat(data)
	return h.Code(str, `json`)
}

func (h *TMarkDown) BlockQuote(str string) string {
	return fmt.Sprintf("> %s", str)
}

func (h *TMarkDown) Enter(str string) string {
	return fmt.Sprintf("%s  \n", str)
}
