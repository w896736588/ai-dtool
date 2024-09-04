package redis_manager

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"testing"
)

func TestTest(t *testing.T) {
	bash := `#!/bin/bash  
app_info='{app_info}'  
wechatAppId=$(echo "$app_info" | jq -r '._id')  
appId=$(echo "$app_info" | jq -r '.app_id')  
dockerList=$(sudo docker ps --format "{{.Names}}" | grep xkf)  
if [ -z "$dockerList" ]; then  
    echo "未查找到任何带有 'xkf' 的运行的docker容器."  
    exit 0  
fi  
for dockerName in $dockerList; do  
    pidList=$(sudo docker exec "$dockerName" sh -c "ps aux | grep -i -E \"$wechatAppId|$appId\" | grep -v grep | awk '{print \$2}'")  
    if [ -z "$pidList" ]; then  
        echo "$dockerName 未查找到任何运行的微信客服进程."  
        continue  
    fi  
    for pid in $pidList; do  
        sudo docker exec "$dockerName" kill -9 "$pid"  
         echo "$dockerName kill -9 $pid."  
    done  
done`
	replaceLists := ` [{"{choose_env_info}.admin_user_id":"799533513"},{"{choose_env_info}.label":"xkf_common3"},{"{choose_env_info}.value":"xkf_common3"},{"{choose_env_info}.code_dir":"yii_customer_service"},{"{choose_env_info}":"{\"admin_user_id\":\"799533513\",\"code_dir\":\"yii_customer_service\",\"label\":\"xkf_common3\",\"value\":\"xkf_common3\"}"},{"{select_wechat_list}":"[{\"_id\":\"210866\",\"app_id\":\"wpX2IKEAAAwS7tM_udiV9JL4FVibhXpw\",\"label\":\"上海芝麻小事网络科技有限公司\",\"value\":\"210866\"},{\"_id\":\"210867\",\"app_id\":\"ww6b983da349fd945e\",\"label\":\"上海芝麻小事密码接入\",\"value\":\"210867\"},{\"_id\":\"210886\",\"app_id\":\"ww002ef006aa9eb31e\",\"label\":\"全新创建企业\",\"value\":\"210886\"},{\"_id\":\"211284\",\"app_id\":\"wpX2IKEAAA9PH4WJVe2nQgEfOh7MXD-A\",\"label\":\"芝麻微客V2\",\"value\":\"211284\"}]"},{"{app_info}._id":"211284"},{"{app_info}.app_id":"wpX2IKEAAA9PH4WJVe2nQgEfOh7MXD-A"},{"{app_info}.label":"芝麻微客V2"},{"{app_info}.value":"211284"},{"{app_info}":"{\"_id\":\"211284\",\"app_id\":\"wpX2IKEAAA9PH4WJVe2nQgEfOh7MXD-A\",\"label\":\"芝麻微客V2\",\"value\":\"211284\"}"}]`
	replaceList := make([]map[string]string, 0)
	gstool.JsonDecode(replaceLists, &replaceList)
	for _, replace := range replaceList {
		bash = gstool.StringReplaces(bash, replace)
	}
	gstool.FmtPrintlnLogTime(`%s`, bash)
}
