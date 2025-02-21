package ai_model

import (
	"dev_tool/internal/pkg/ai/ai_define"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

func ModelNo(sql string) ([]ai_define.Message, []ai_define.Tool, error) {
	modelUse := `不分表`
	table := "CREATE TABLE `tbl_customer` (\n  `_id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `kefu_user_id` int(11) DEFAULT NULL COMMENT '客服用户id',\n  `create_time` int(11) DEFAULT NULL,\n  `update_time` int(11) DEFAULT NULL,\n  PRIMARY KEY (`_id`),\n  UNIQUE KEY `openid_wechatid_kefu_id` (`openid`,`wechatapp_id`,`kefu_user_id`)\n) ENGINE=InnoDB AUTO_INCREMENT=46337208 DEFAULT CHARSET=utf8 COMMENT='客户表_20210705';"
	class := "<?php \n/**\n * 客户表_20210705\n * @User: frog\n * @Date: 2025/02/21 17:51\n */\nclass CustomerModel extends BaseModel {\n\n    public function __construct($db = null) {\n        parent::__construct($db);\n        $this->table = 'tbl_customer';\n        $this->cols  = [\n           '_id',                                 //_id\n           'kefu_user_id',                        //客服用户id\n           'create_time',                         //create_time\n           'update_time',                         //update_time\n        ];\n    }\n}"
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
	return []ai_define.Message{
		{
			Role:    ai_define.RoleSystem,
			Content: strings.Join(descList, `。`),
		},
		{
			Role:    ai_define.RoleUser,
			Content: strings.Join(needList, `。`),
		},
	}, []ai_define.Tool{}, nil
}
