package ai_model

import (
	"dev_tool/internal/pkg/ai/ai_define"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

func ModelYearMonthMod(sql string, mod string) ([]ai_define.Message, []ai_define.Tool, error) {
	modelUse := `按年月取模模分表`
	table := "CREATE TABLE `lottery_detail_record_2025_12` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `avatar_logo` varchar(255) NOT NULL DEFAULT '' COMMENT '头像',\n  `create_time` bigint(20) unsigned NOT NULL DEFAULT '0',\n  `update_time` bigint(20) unsigned NOT NULL DEFAULT '0'\n  PRIMARY KEY (`id`) USING BTREE\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='抽奖签到明细表';"
	class := "<?php \n/**\n * 抽奖签到明细表\n * @User: frog\n * @Date: 2025/02/21 18:08\n */\nclass LotteryDetailModel extends BaseModel {\n\n    public function __construct($db = null) {\n        parent::__construct($db);\n        $this->table = 'lottery_detail';\n        $this->cols  = [\n           'id',                                  //id\n           'avatar_logo',                         //头像\n           'create_time',                         //create_time\n           'update_time',                         //update_time\n        ];\n    }\n    /**\n     * 按年按月按管理员分表\n     */\n    public function setTableName($year , $month , $admin_user_id): string {\n        $this->table = 'lottery_detail_' . $year . '_' . $month . '_' . ($admin_user_id%10);\n        return $this->table;\n    }\n}"
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
