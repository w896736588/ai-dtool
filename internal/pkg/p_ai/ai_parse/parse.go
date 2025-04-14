package ai_parse

import (
	_struct "dev_tool/base/struct"
	"dev_tool/internal/pkg/p_ai/ai_parse/ai_action_tpl"
	"dev_tool/internal/pkg/p_ai/ai_parse/ai_model_tpl"
	"dev_tool/internal/pkg/p_ai/ai_parse/ai_other_set"
	"dev_tool/internal/pkg/p_ai/ai_parse/ai_service_tpl"
	"errors"
	"github.com/spf13/cast"
)

type Parse struct {
	Data map[string]any
}

func NewParse(data map[string]any) *Parse {
	return &Parse{
		Data: data,
	}
}

func (h *Parse) Parse() ([]_struct.Message, []_struct.Tool, error) {
	op := h.Data[`op`].(string)
	switch op {
	case `model`:
		return h.ParseModel()
	case `action`:
		return h.ParseAction()
	case `service`:
		return h.ParseService()
	default:
		return []_struct.Message{}, []_struct.Tool{}, errors.New(`暂不支持` + op)
	}
}

func (h *Parse) ParseModel() ([]_struct.Message, []_struct.Tool, error) {
	sql := h.Data[`sql`].(string)
	modelType := h.Data[`modelType`].(string)
	switch modelType {
	case `no`:
		return ai_model_tpl.ModelNo(sql)
	case `year`:
		return ai_model_tpl.ModelYear(sql)
	case `mod`:
		return ai_model_tpl.ModelMod(sql, h.Data[`mod`].(string))
	case `year_month`:
		return ai_model_tpl.ModelYearMonth(sql)
	case `year_mod`:
		return ai_model_tpl.ModelYearMod(sql, h.Data[`mod`].(string))
	case `year_month_mod`:
		return ai_model_tpl.ModelYearMonthMod(sql, h.Data[`mod`].(string))
	default:
		return []_struct.Message{}, []_struct.Tool{}, errors.New(`暂不支持` + modelType)
	}
}

func (h *Parse) ParseAction() ([]_struct.Message, []_struct.Tool, error) {
	//actionList := h.Data[`actionList`].([]any)
	actionList := []any{`list`, `detail`, `create`, `delete`}
	return ai_action_tpl.Action(actionList)
}

func (h *Parse) ParseService() ([]_struct.Message, []_struct.Tool, error) {
	cacheType := h.Data[`cacheType`].(string)
	mainTemplateField := cast.ToString(h.Data[`main_template_field`])
	childTemplateField := cast.ToString(h.Data[`childTemplateField`])
	return ai_service_tpl.Service(cacheType, mainTemplateField, childTemplateField)
}

func (h *Parse) ParseOtherSet() ([]_struct.Message, []_struct.Tool, error) {
	otherSetList := h.Data[`otherSetList`].([]any)
	actionPrefix := h.Data[`actionPrefix`].(string)
	return ai_other_set.OtherSet(otherSetList, actionPrefix)
}
