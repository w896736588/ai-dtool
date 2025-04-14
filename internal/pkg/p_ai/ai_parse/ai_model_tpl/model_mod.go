package ai_model_tpl

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

func ModelMod(sql string, mod string) ([]_struct.Message, []_struct.Tool, error) {
	modelUse := `按模分表`
	table := "CREATE TABLE `tbl_chat_label_customer_6` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n  `admin_user_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '管理员用户ID',\n  PRIMARY KEY (`id`),\n  UNIQUE KEY `uni_wechatapp_openid_label` (`wechatapp_id`,`openid`,`label_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客户打上的聊天标签6';"
	class := `<?php 
/**
 * 客户打上的聊天标签6
 * @User: frog
 * @Date: 2025/02/21 17:52
 */
class ChatLabelCustomerModel extends BaseModel {

    public function __construct($db = null) {
        parent::__construct($db);
        $this->table = 'tbl_chat_label_customer';
        $this->cols  = [
           'id',                                  //ID
           'admin_user_id',                       //管理员用户ID
        ];
    }

    /**
     * 按管理员分表
     */
    public function setTableName($admin_user_id): string {
        $this->table = 'tbl_chat_label_customer_' . ($admin_user_id%10);
        return $this->table;
    }
}`
	descList := []string{
		`你是一个php开发者，会生成class model，下面是示例`,
		`假如有一个table：` + table,
		`生成了一个php类:` + class,
		`这是` + modelUse + `的示例,其中，%10中的10就是分表数`,
	}
	needList := []string{
		`现在我给你一个sql：` + sql,
		`帮我生成一个` + modelUse + `的model php 类，注意分表数为` + mod,
		`@Date后面的时间帮我换为 ` + gstool.DateCurrent(),
		`不需要告诉我过程,请用Markdown格式输出代码，确保格式要保留缩进和换行。`,
	}
	return []_struct.Message{
		{
			Role:    define.RoleSystem,
			Content: strings.Join(descList, `。`),
		},
		{
			Role:    define.RoleUser,
			Content: strings.Join(needList, `。`),
		},
	}, []_struct.Tool{}, nil
}
