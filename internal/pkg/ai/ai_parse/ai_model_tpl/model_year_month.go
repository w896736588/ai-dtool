package ai_model_tpl

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

func ModelYearMonth(sql string) ([]_struct.Message, []_struct.Tool, error) {
	modelUse := `按年月分表`
	table := "CREATE TABLE `tbl_mp_unionid_record_2025_12` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `admin_user_id` int(11) NOT NULL DEFAULT '0' COMMENT '管理员id',\n  `create_time` int(11) NOT NULL DEFAULT '0',\n  `update_time` int(11) NOT NULL DEFAULT '0',\n  PRIMARY KEY (`id`) USING BTREE,\n  KEY `union_idx` (`wechatapp_id`,`rule_id`,`unionid`) USING BTREE\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='unionid预打标签存储表';"
	class := `<?php 
/**
 * unionid预打标签存储表
 * @User: frog
 * @Date: 2025/02/21 18:01
 */
class MpUnionidRecordModel extends BaseModel {

    public function __construct($db = null) {
        parent::__construct($db);
        $this->table = 'tbl_mp_unionid_record';
        $this->cols  = [
           'id',                                  //id
           'admin_user_id',                       //管理员id
           'create_time',                         //create_time
           'update_time',                         //update_time
        ];
    }

    /**
     * 按年按月分表
     */
    public function setTableName($year , $month): string {
        $this->table = 'tbl_mp_unionid_record_' . $year . '_' . $month;
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
