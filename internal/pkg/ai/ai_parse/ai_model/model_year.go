package ai_model

import (
	"dev_tool/internal/pkg/ai/ai_define"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

func ModelYear(sql string) ([]ai_define.Message, []ai_define.Tool, error) {
	modelUse := `按年分表`
	table := "CREATE TABLE `tbl_kf_response_stat_detail_2022` (\n  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,\n  `wechatapp_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '应用ID',\n  `response_type` int(10) unsigned NOT NULL DEFAULT '1' COMMENT '1 未回复（老数据都是未回复）  2 未应答',\n  PRIMARY KEY (`id`),\n  KEY `create_date_time` (`create_date_time`),\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客服应答统计 未回复明细以及未应答明细';"
	class := "<?php \n/**\n * 客服应答统计 未回复明细以及未应答明细\n * @User: frog\n * @Date: 2025/02/21 15:16\n */\nclass KfResponseStatDetailModel extends BaseModel {\n\n    public function __construct($db = null) {\n        parent::__construct($db);\n        $this->table = 'tbl_kf_response_stat_detail';\n        $this->cols  = [\n           'id',                                  //id\n           'wechatapp_id',                        //应用ID\n           'response_type',                       //1 未回复（老数据都是未回复）  2 未应答\n        ];\n    }\n\n    /**\n     * 按年分表\n     */\n    public function setTableName($year): string {\n        $this->table = 'tbl_kf_response_stat_detail_' . $year;\n        return $this->table;\n    }\n}"
	descList := []string{
		`你是一个php开发者，会生成class model，下面是示例`,
		`假如有一个table：` + table,
		`生成了一个php类:` + class,
		`这是` + modelUse + `的示例`,
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
