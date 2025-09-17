package test

import (
	"sync"
	"testing"
	"time"

	"gitee.com/Sxiaobai/gs/gstool"
)

var wg sync.WaitGroup

// TestFpm 测试fpm无session的情况
func TestFpmNoSession(t *testing.T) {
	str := `[{"label":"传统机器人","value":"ai_server","build":"D:\GnuWin32\bin\make.exe -v && cd D:\\go\\go-micro-server1.14\\aiServer && D:\GnuWin32\bin\make.exe SHELL=cmd.exe aiServer","docker_cmd":"sudo docker exec xkf_common sh -c 'supervisorctl restart aiServer'","local_file":"D:\\go\\go-micro-server1.14\\aiServer\\aiServer","chmod_x_cmd":"sudo chmod +x /var/www/docker_apps/common/go-micro-server1.14/aiServer/aiServer","target_dir":"/var/www/docker_apps/common/go-micro-server1.14/aiServer"},{"label":"表单","value":"form","build":"D:\GnuWin32\bin\make.exe -v && cd D:\\go\\fansdiscover && D:\GnuWin32\bin\make.exe SHELL=cmd.exe linux_fansDiscover","docker_cmd":"sudo docker exec xkf_common sh -c 'supervisorctl restart fansDiscover'","local_file":"D:\\go\\fansdiscover\\fansDiscover","chmod_x_cmd":"sudo chmod +x /var/www/docker_apps/common/fansdiscover/fansDiscover","target_dir":"/var/www/docker_apps/common/fansdiscover"}]`
	list := make([]map[string]any, 0)
	err := gstool.JsonDecode(str, &list)
	if err != nil {
		gstool.FmtPrintlnLogTime(`%s`, err.Error())
	} else {
		gstool.FmtPrintlnLogTime(`%s`, gstool.JsonFormat(list))
	}
}

func HttpNoCookie() {
	gslog := gstool.NewSlog3(`./`, `test`)
	headers := make(map[string]string)
	get, err := gstool.NewHttp(time.Second*60, gslog).
		HttpPostJson(`https://bug.xiaokefu.com.cn/test/t`, ``, &headers)
	if err != nil {
		gstool.FmtPrintlnLogTime(`err:%v`, err)
		return
	} else {
		gstool.FmtPrintlnLogTime(`结果 %s`, get)
	}
	wg.Done()
}

func HttpHasCookie() {
	gstool.FmtPrintlnLogTime(`开始`)
	gslog := gstool.NewSlog3(`./`, `test`)
	headers := map[string]string{
		`cookie`: `Hm_lvt_a4964b2514693874bb3c7104e129d76e=1730252646; Hm_lvt_c0af941dc5b11c6ccf7dfa77fc3e3c24=1733465577; Hm_lvt_2911e7fbbc2af45ce5bee6f3e22033c6=1740456213; 75a3df0a6c405c8ce2563b2c882753cb=15768cc1acddf62c1fb26e5885b5d8bbcf383ab0a%3A4%3A%7Bi%3A0%3Bs%3A9%3A%22800078833%22%3Bi%3A1%3Bs%3A9%3A%22800078833%22%3Bi%3A2%3Bi%3A604800%3Bi%3A3%3Ba%3A2%3A%7Bs%3A7%3A%22user_id%22%3Bs%3A9%3A%22800078833%22%3Bs%3A8%3A%22username%22%3Bs%3A11%3A%2213317179268%22%3B%7D%7D; xkf_login_token=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE3NDA0NTYyMTcsImV4cCI6MTc0MTA2MTAxNywiZGF0YSI6eyJ1c2VyX2lkIjoiODAwMDc4ODMzIiwidXNlcl9uYW1lIjoiMTMzMTcxNzkyNjgiLCJzaWduIjoiNWQ1ZjZlMDk4NDM0NmViY2QwYzQxMDBkNWFmMTRiZTciLCJ0eXBlIjoxfX0.LGDHOshj0eQXxlWHDtxGV2Tyu_wVECjaguNJjfMznCs; Hm_lvt_2911e7fbbc2af45ce5bee6f3e22033c6=1739324762,1739518087,1740382135,1740471488; sendingTextPeriod=true; last_flag=yun; yun_wechatapp_id=359251; system:kefu:last:use:app:info:800078833=800078833%7C359251; Hm_lvt_e57b8b134e41424995fb7e19768f061e=1740535065,1740620877,1740706948,1740969010; Hm_lpvt_e57b8b134e41424995fb7e19768f061e=1740969010; HMACCOUNT=C670E581006CCD0C; Hm_lvt_26b5094a3b36a601595d7a7521f2a840=1740535065,1740620877,1740706949,1740969010; xkf:last:use:system:800078833=kefu; yii_zhima_session=57f1skntcp5daqlk8pupqiaja5; Hm_lpvt_26b5094a3b36a601595d7a7521f2a840=1740982216; xkf_userid=800078833`,
	}
	get, err := gstool.NewHttp(time.Second*60, gslog).
		HttpPostJson(`https://bug.xiaokefu.com.cn/test/t`, ``, &headers)
	if err != nil {
		gstool.FmtPrintlnLogTime(`err:%v`, err)
		return
	} else {
		gstool.FmtPrintlnLogTime(`结果 %s`, get)
	}
	wg.Done()
}
