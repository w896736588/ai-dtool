package ai_parse

import (
	"dev_tool/internal/pkg/ai/ai_define"
	"dev_tool/internal/pkg/ai/ai_parse/ai_model"
	"errors"
)

type Parse struct {
	Data map[string]any
}

func NewParse(data map[string]any) *Parse {
	return &Parse{
		Data: data,
	}
}

func (h *Parse) Parse() ([]ai_define.Message, []ai_define.Tool, error) {
	op := h.Data[`op`].(string)
	switch op {
	case `model`:
		return h.ParseModel()
	default:
		return []ai_define.Message{}, []ai_define.Tool{}, errors.New(`暂不支持` + op)
	}
}

func (h *Parse) ParseModel() ([]ai_define.Message, []ai_define.Tool, error) {
	sql := h.Data[`sql`].(string)
	modelType := h.Data[`modelType`].(string)
	switch modelType {
	case `no`:
		return ai_model.ModelNo(sql)
	case `year`:
		return ai_model.ModelYear(sql)
	case `mod`:
		return ai_model.ModelMod(sql, h.Data[`mod`].(string))
	case `year_month`:
		return ai_model.ModelYearMonth(sql)
	case `year_mod`:
		return ai_model.ModelYearMod(sql, h.Data[`mod`].(string))
	case `year_month_mod`:
		return ai_model.ModelYearMonthMod(sql, h.Data[`mod`].(string))
	default:
		return []ai_define.Message{}, []ai_define.Tool{}, errors.New(`暂不支持` + modelType)
	}
}
