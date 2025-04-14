package ai_other_set

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"strings"
)

func OtherSet(otherSetList []any, actionPrefix string) ([]_struct.Message, []_struct.Tool, error) {
	descList := make([]string, 0)
	for _, otherSet := range otherSetList {
		switch otherSet {
		case `memo`:
			descList = append(descList, `根据action帮我生生接口文档,只需要生成请求参数，并对参数进行描述，包括返回参数解释`)
			descList = append(descList, `注意，因为编辑和新增是一个接口，所以要说明编辑的时候才需要传递id字段`)
			descList = append(descList, `注意，任何时候admin_user_id不是新增编辑的传递参数`)
			if actionPrefix != `` {
				descList = append(descList, `action前缀为`+actionPrefix+`,比如前缀为/kf,TemplateController,actionList的接口地址的组成为/kf/Template/List`)
			}
			break
		}
	}
	return []_struct.Message{
		{
			Role:    define.RoleUser,
			Content: strings.Join(descList, `。`),
		},
	}, []_struct.Tool{}, nil
}
