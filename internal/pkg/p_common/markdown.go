package p_common

import (
	"fmt"
	"github.com/w896736588/go-tool/gstool"
)

var TMarkDownClient *TMarkDown

type TMarkDown struct {
}

func (h *TMarkDown) Code(str, lang string) string {
	return fmt.Sprintf("```%s\n\n%s\n```", lang, str)
}

func (h *TMarkDown) Json(data any) string {
	str := gstool.JsonFormat(data)
	return h.Code(str, `json`)
}

func (h *TMarkDown) BlockQuote(str string) string {
	return fmt.Sprintf("> %s\n", str)
}

func (h *TMarkDown) Enter(str string) string {
	return fmt.Sprintf("%s  \n", str)
}

func (h *TMarkDown) Bold(str string) string {
	return fmt.Sprintf("**%s**", str)
}
