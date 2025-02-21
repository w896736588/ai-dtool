package test

import (
	"dev_tool/internal/pkg/ai/ai_bailian"
	"dev_tool/internal/pkg/ai/ai_define"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
	"testing"
)

// BailianTest 百炼 qwen2.5-coder-3b-instruct 模型
func BailianTest(t *testing.T) {
	ai := ai_bailian.NewBailian(`qwen2.5-coder-3b-instruct`, `sk-938dc32c6e394fe089e64aac7ee6443f`, true)
	gslog := gstool.NewSlogDefault(`./`, `ai_test`)
	table := "CREATE TABLE `tbl_kf_response_stat_detail_2022` (\n  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,\n  `wechatapp_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '应用ID',\n  `response_type` int(10) unsigned NOT NULL DEFAULT '1' COMMENT '1 未回复（老数据都是未回复）  2 未应答',\n  PRIMARY KEY (`id`),\n  KEY `create_date_time` (`create_date_time`),\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='客服应答统计 未回复明细以及未应答明细';"
	class := "<?php \n/**\n * 客服应答统计 未回复明细以及未应答明细\n * @User: frog\n * @Date: 2025/02/21 15:16\n */\nclass KfResponseStatDetailModel extends BaseModel {\n\n    public function __construct($db = null) {\n        parent::__construct($db);\n        $this->table = 'tbl_kf_response_stat_detail';\n        $this->cols  = [\n           'id',                                  //id\n           'wechatapp_id',                        //应用ID\n           'response_type',                       //1 未回复（老数据都是未回复）  2 未应答\n        ];\n    }\n\n    /**\n     * 按年分表\n     */\n    public function setTableName($year): string {\n        $this->table = 'tbl_kf_response_stat_detail_' . $year;\n        return $this->table;\n    }\n}"
	descList := []string{
		`你是一个php开发者，会生成class model，下面是示例`,
		`假如有一个table：` + table,
		`生成了一个php类:` + class,
		`这是按年分表的示例`,
	}
	sql := "CREATE TABLE `tbl_kf_response_chat_detail_2022_08` (\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\n  `admin_user_id` int(11) unsigned NOT NULL DEFAULT '0',\n  `wechatapp_id` int(11) unsigned NOT NULL DEFAULT '0',\n  `channel_id` int(11) unsigned NOT NULL DEFAULT '0',\n  `staff_user_id` int(11) unsigned NOT NULL DEFAULT '0',\n  `date` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '日期天：20220303',\n  `openid` varchar(190) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'openid',\n  `chat_session_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '会话ID',\n  `first_response_second` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '首次应答秒数',\n  `first_response_unexceed` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '首次应答是否在设置的分钟数内，1未超出，0超出',\n  `turn_num_total` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '对话轮次',\n  `turn_num_unexceed_total` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '未超过设置的分钟数的对话轮次',\n  `create_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'create_time',\n  `update_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'update_time',\n  `need_count_first_response` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '是否需要计算 首次应答率 1需要：0不需要',\n  `staff_first_response_time` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '客服首次回复时间',\n  `turn_num_qualified_unexceed_total` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '未超过设置的分钟数的对话轮次 合格值',\n  `first_response_qualified_unexceed` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '首次应答是否在设置的分钟数内，1未超出，0超出 合格值',\n  `response_second_max` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '最长响应时长',\n  `response_second_sum` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '响应时长累计，秒',\n  PRIMARY KEY (`id`),\n  KEY `idx_staff_wechat_date_channel` (`staff_user_id`,`wechatapp_id`,`date`,`channel_id`),\n  KEY `idx_wechat_date_channel` (`wechatapp_id`,`date`,`channel_id`),\n  KEY `idx_admin_sessionid` (`admin_user_id`,`chat_session_id`)\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客服应答明细 一个会话一条记录';"
	needList := []string{
		`现在我给你一个sql：` + sql,
		`帮我生成一个按年分表的model php 类，注意这个类的创建时间要是最新的时间`,
		`@Date后面的时间帮我换为 ` + gstool.DateCurrent(),
		`不需要告诉我过程,请用Markdown格式输出代码，确保格式要保留缩进和换行。`,
	}
	ret, retErr := ai.Api([]ai_define.Message{
		{
			Role:    ai_define.RoleSystem,
			Content: strings.Join(descList, `。`),
		},
		{
			Role:    ai_define.RoleUser,
			Content: strings.Join(needList, `。`),
		},
	}, []ai_define.Tool{})
	if retErr != nil {
		gstool.FmtPrintlnLogTime(`生成失败 %s`, retErr.Error())
	} else {
		gslog.Debugf(ret)
	}
}
