package ai_action_tpl

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"strings"
)

func Action(actionTypeList []any) ([]_struct.Message, []_struct.Tool, error) {
	classList := make([]string, 0)
	classList = append(classList, ActionClass())
	for _, actionType := range actionTypeList {
		switch actionType {
		case `list`:
			classList = append(classList, ActionList())
			break
		case `create`:
			classList = append(classList, ActionCreate())
			break
		case `detail`:
			classList = append(classList, ActionDetail())
			break
		case `delete`:
			classList = append(classList, ActionDelete())
			break
		default:
			return []_struct.Message{}, []_struct.Tool{}, errors.New(`不支持的操作`)
		}
	}
	classList = append(classList, `}`)
	descList := []string{
		`你是一个php开发者，根据模板生成action，下面是示例，注意示例中的[]包起来的是提示,类名取值和注释应该基于生成的model类名`,
		`示例php controller:` + strings.Join(classList, "\n"),
	}
	return []_struct.Message{
		{
			Role:    define.RoleUser,
			Content: strings.Join(descList, `。`),
		},
	}, []_struct.Tool{}, nil
}

func ActionClass() string {
	return `<?php

/**
 * [使用数据表的备注替换]
 * Created by PhpStorm.
 * User: frog
 * Date: 2024/5/16 2:32
 */
class TemplateController extends BaseController
{
    private $userId;
    private $adminUserId;

    public function __construct($id,$module=null)
    {
        parent::__construct($id,$module);
        $this->userId = getUserId();
        $this->adminUserId = StaffServices::getAdminUserIdByStaffId(getUserId());
    }
`
}
