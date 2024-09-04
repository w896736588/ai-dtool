go>	以执行微信客服启动为例  
	涉及到mysql查询，docker查询，kill命令，执行命令。
```
use ssh id=1
{app_id}=wait input
{app_info}=[mysql id=8]=select _id,app_id from tbl_wechatapp where _id = {app_id} or app_id = '{app_id}' limit 1
{xkf_docker}=sudo docker ps --format "{{.Names}}" | grep xkf
sudo docker exec [xkf_docker.each] sh -c "ps aux | grep [variable:1724052744563.each.any] | grep -v grep  | awk '{print "\n" \$2}'"
```