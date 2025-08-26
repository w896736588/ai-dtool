package test

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/pion/stun"
	"net"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestBailian 百炼 qwen2.5-coder-3b-instruct 模型
func TestBailian(t *testing.T) {
	str := "    Z{\"op\":\"set\",\"eventOffset\":1,\"chat\":{\"id\":\"d2mgcfq5hvlm18gjmv0g\",\"name\":\"未命名会话\"}}     {\"eventOffset\":1,\"heartbeat\":{}}   \v�{\"op\":\"set\",\"mask\":\"message\",\"eventOffset\":2,\"message\":{\"id\":\"d2mgcfuf2kq0qd3aafig\",\"role\":\"user\",\"status\":\"MESSAGE_STATUS_COMPLETED\",\"blocks\":[{\"id\":\"text_0_0\",\"text\":{\"content\":\"你是一个php开发者，会生成class model，下面是示例。假如有一个table：\\n ```sql CREATE TABLE `tbl_customer` (  `_id` int(11) unsigned NOT NULL AUTO_INCREMENT,  `kefu_user_id` int(11) DEFAULT NULL COMMENT '客服用户id',  `create_time` int(11) DEFAULT NULL,  `update_time` int(11) DEFAULT NULL,  PRIMARY KEY (`_id`),  UNIQUE KEY `openid_wechatid_kefu_id` (`openid`,`wechatapp_id`,`kefu_user_id`) ) ENGINE=InnoDB AUTO_INCREMENT=46337208 DEFAULT CHARSET=utf8 COMMENT='客户表_20210705'; ``` \\n生成了一个php类:\\n ```php <?php  /**  * 客户表_20210705  * @User: frog  * @Date: 2025/02/21 17:51  */ class CustomerModel extends BaseModel {   public function __construct($db = null) {  parent::__construct($db);  $this->table = 'tbl_customer';  $this->cols = [  '_id', //_id  'kefu_user_id', //客服用户id  'create_time', //create_time  'update_time', //update_time  ];  } } ```\\n这是不分表分表的示例 现在我给你一个sql。帮我生成一个不分表的model php 类，注意这个类的创建时间要是最新的时间。@Date后面的时间帮我换为 2025-04-21 15:41:02。不需要告诉我过程,请用Markdown格式输出代码，确保格式要保留缩进和换行。注意，忽略分表数。\\n\\nsql如下：\\n```sql\\nCREATE TABLE `clock_in_detail_record_2024_0` (\\n  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,\\n  `admin_user_id` int(11) unsigned NOT NULL DEFAULT '0',\\n  `wechatapp_id` int(11) unsigned NOT NULL DEFAULT '0',\\n  `clock_in_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '打卡签到活动ID',\\n  `openid` varchar(128) NOT NULL,\\n  `nick_name` varchar(128) NOT NULL COMMENT '昵称',\\n  `avatar_logo` varchar(128) NOT NULL COMMENT '头像',\\n  `incr_int` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '增加积分',\\n  `decr_int` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '扣除积分',\\n  `extra_incr_int` int(11) unsigned NOT NULL COMMENT '连续签到赠送积分',\\n  `create_time` int(11) unsigned NOT NULL DEFAULT '0',\\n  `update_time` int(11) unsigned NOT NULL DEFAULT '0',\\n  `point_type` tinyint(1) NOT NULL DEFAULT '0' COMMENT '类型0：签到 1：后台添加',\\n  `auto_decr_int` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '系统自动扣除积分',\\n  PRIMARY KEY (`id`) USING BTREE,\\n  UNIQUE KEY `uni_clock_openid_time` (`clock_in_id`,`openid`,`create_time`) USING BTREE,\\n  KEY `admin_user_id` (`admin_user_id`,`wechatapp_id`) USING BTREE,\\n  KEY `idx_clockinid_pointtype_openid` (`point_type`,`clock_in_id`,`openid`) USING BTREE\\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='打卡签到明细表';\\n```\\n分表数为：\\n{table_mod}\\n\\n注意：\\n1.col中每一行//的注释需要对齐，每个缩进为4个空格\\n2.没有COMMENT的字段，默认为//[字段]\"}}]}}    �{\"op\":\"set\",\"mask\":\"message\",\"eventOffset\":3,\"message\":{\"id\":\"d2mgcfuf2kq0qd3aafj0\",\"parentId\":\"d2mgcfuf2kq0qd3aafig\",\"role\":\"assistant\",\"status\":\"MESSAGE_STATUS_GENERATING\",\"scenario\":\"SCENARIO_K2\"}}    �{\"op\":\"set\",\"mask\":\"chat.name\",\"eventOffset\":4,\"chat\":{\"name\":\"你是一个php开发者，会生成class model，下面是示\"}}     {\"eventOffset\":4,\"heartbeat\":{}}     {\"eventOffset\":4,\"heartbeat\":{}}     {\"eventOffset\":4,\"heartbeat\":{}}     {\"eventOffset\":4,\"heartbeat\":{}}    i{\"op\":\"append\",\"mask\":\"block.text.content\",\"eventOffset\":5,\"block\":{\"id\":\"0_0\",\"text\":{\"content\":\"```\"}}}"
	re := regexp.MustCompile(`\s{4}.\{`)
	parts := re.Split(str, -1)
	resultList := make([]string, 0)
	for _, part := range parts {
		if strings.Trim(part, ` `) == `` {
			continue
		}
		//在按照
		secondList := regexp.MustCompile(`\s{3}.{2}\{`)
		for _, secondPart := range secondList.Split(`{`+part, -1) {
			resultList = append(resultList, secondPart)
		}
	}
	fmt.Println(gstool.JsonFormat(resultList)) // 输出: [ b d f h j]
}

func getPublicIPWithSTUN() (string, error) {
	// 1. 创建UDP连接
	conn, err := net.Dial("udp", "stun.l.google.com:19302") // Google公共STUN服务器
	if err != nil {
		return "", fmt.Errorf("创建UDP连接失败: %v", err)
	}
	defer conn.Close()

	// 2. 设置超时
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return "", fmt.Errorf("设置超时失败: %v", err)
	}

	// 3. 创建STUN客户端
	client, err := stun.NewClient(conn)
	if err != nil {
		return "", fmt.Errorf("创建STUN客户端失败: %v", err)
	}
	defer client.Close()

	// 4. 构建STUN请求
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// 5. 处理响应
	var publicIP string
	err = client.Do(message, func(res stun.Event) {
		if res.Error != nil {
			return
		}

		// 解析XOR-MAPPED-ADDRESS属性
		var xorAddr stun.XORMappedAddress
		if err := xorAddr.GetFrom(res.Message); err != nil {
			return
		}
		publicIP = xorAddr.IP.String()
	})

	if err != nil {
		return "", fmt.Errorf("STUN请求失败: %v", err)
	}

	if publicIP == "" {
		return "", fmt.Errorf("未从STUN响应中获取到IP地址")
	}

	return publicIP, nil
}
