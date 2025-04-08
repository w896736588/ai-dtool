package ai_model_tpl

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

func ModelYear(sql string) ([]_struct.Message, []_struct.Tool, error) {
	modelUse := `按年分表`
	table := "CREATE TABLE `tbl_kf_response_stat_detail_2022` (\n  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,\n  `wechatapp_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '应用ID',\n  `response_type` int(10) unsigned NOT NULL DEFAULT '1' COMMENT '1 未回复（老数据都是未回复）  2 未应答',\n  PRIMARY KEY (`id`),\n  KEY `create_date_time` (`create_date_time`),\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客服应答统计 未回复明细以及未应答明细';"
	class := `<?php 
/**
 * 客服应答统计 未回复明细以及未应答明细
 * @User: frog
 * @Date: 2025/02/21 15:16
 */
class KfResponseStatDetailModel extends BaseModel {

    public function __construct($db = null) {
        parent::__construct($db);
        $this->table = 'tbl_kf_response_stat_detail';
        $this->cols  = [
           'id',                                  //id
           'wechatapp_id',                        //应用ID
           'response_type',                       //1 未回复（老数据都是未回复）  2 未应答
        ];
    }

    /**
     * 按年分表
     */
    public function setTableName($year): string {
        $this->table = 'tbl_kf_response_stat_detail_' . $year;
        return $this->table;
    }
}`
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
