package ai_model_tpl

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"strings"
)

func ModelNo(sql string) ([]_struct.Message, []_struct.Tool, error) {
	modelUse := `不分表`
	table := "CREATE TABLE `tbl_customer` (\n  `_id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `kefu_user_id` int(11) DEFAULT NULL COMMENT '客服用户id',\n  `create_time` int(11) DEFAULT NULL,\n  `update_time` int(11) DEFAULT NULL,\n  PRIMARY KEY (`_id`),\n  UNIQUE KEY `openid_wechatid_kefu_id` (`openid`,`wechatapp_id`,`kefu_user_id`)\n) ENGINE=InnoDB AUTO_INCREMENT=46337208 DEFAULT CHARSET=utf8 COMMENT='客户表_20210705';"
	class := `<?php 
/**
 * 客户表_20210705
 * @User: frog
 * @Date: 2025/02/21 17:51
 */
class CustomerModel extends BaseModel {

    public function __construct($db = null) {
        parent::__construct($db);
        $this->table = 'tbl_customer';
        $this->cols  = [
           '_id',                                 //_id
           'kefu_user_id',                        //客服用户id
           'create_time',                         //create_time
           'update_time',                         //update_time
        ];
    }
}`
	descList := []string{
		`你是一个php开发者，会生成class model，下面是示例`,
		`假如有一个table：` + table,
		`生成了一个php类:` + class,
		`这是` + modelUse + `分表的示例`,
	}
	needList := []string{
		`现在我给你一个sql：` + sql,
		`帮我生成一个` + modelUse + `的model php 类，注意这个类的创建时间要是最新的时间`,
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
