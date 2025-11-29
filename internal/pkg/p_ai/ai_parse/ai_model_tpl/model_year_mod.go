package ai_model_tpl

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"strings"
)

func ModelYearMod(sql string, mod string) ([]_struct.Message, []_struct.Tool, error) {
	modelUse := `按年取模模分表`
	table := "CREATE TABLE `clock_in_detail_record_2025_20` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `admin_user_id` int(11) unsigned NOT NULL DEFAULT '0',\n  `wechatapp_id` int(11) unsigned NOT NULL DEFAULT '0',\n  `clock_in_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '打卡签到活动ID',\n  PRIMARY KEY (`id`) USING BTREE\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='打卡签到明细表';"
	class := `<?php 
/**
 * 打卡签到明细表
 * @User: frog
 * @Date: 2025/02/21 18:05
 */
class ClockInDetailRecordModel extends BaseModel {

    public function __construct($db = null) {
        parent::__construct($db);
        $this->table = 'clock_in_detail_record';
        $this->cols  = [
           'id',                                  //id
           'admin_user_id',                       //admin_user_id
           'wechatapp_id',                        //wechatapp_id
           'clock_in_id',                         //打卡签到活动ID
        ];
    }

    /**
     * 按年按管理员分表
     */
    public function setTableName($year , $admin_user_id): string {
        $this->table = 'clock_in_detail_record_' . $year . '_' . ($admin_user_id%10);
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
