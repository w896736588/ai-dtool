package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"dev_tool/internal/pkg/p_playwright"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gs"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"time"
)

// SmartLinkUpWebkit 更新核心
func SmartLinkUpWebkit(c *gin.Context) {
	pw, _ := playwright.NewDriver()
	go base.Component.TPlaywright.Install(pw.Version)
	gsgin.GinResponseSuccess(c, `更新浏览器核心中`, ``)
	return
}

func SmartLinkRecycle(c *gin.Context) {
	p := p_playwright.NewPlaywright(nil, base.Component.TPlaywright.Log)
	err := p.Recycle()
	if err != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`释放失败 %s`, err.Error()), nil)
		return
	}
	gsgin.GinResponseSuccess(c, `释放成功`, ``)
	return
}

func SmartLinkDownloadPath(c *gin.Context) {
	err := base.Component.TPlaywright.SmartLinkDownloadPath()
	if err != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`释放失败 %s`, err.Error()), nil)
		return
	}
	gsgin.GinResponseSuccess(c, `释放成功`, ``)
	return
}

// SmartLinkList 获取列表
func SmartLinkList(c *gin.Context) {
	variableGroupList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeSmartLink,
	}).All()
	smartLinkList, _ := base.Component.TSqlite.Client.QueryBySql(`select * from tbl_smart_link where status = ? order by weight asc`, define.SmartLinkStatusNormal).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`group_list`:      variableGroupList,
		`smart_link_list`: smartLinkList,
	})
}

// SmartLinkInfo 获取单个详情
func SmartLinkInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	smartLinkId := dataMap[`id`]
	if cast.ToInt(smartLinkId) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	smartLinkInfo, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_smart_link`, `*`, map[string]any{
		`id`:     smartLinkId,
		`status`: define.SmartLinkStatusNormal,
	}).One()
	smartLinkProcessList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_smart_link_process`, `*`, map[string]any{
		`smart_link_id`: smartLinkId,
		`status`:        define.SmartLinkStatusNormal,
	}).Order(`weight asc`).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`smart_link_info`:         smartLinkInfo,
		`smart_link_process_list`: smartLinkProcessList,
	})
}

// SmartLinkAdd 新增
func SmartLinkAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	validateErr := validateProcess(cast.ToString(dataMap[`process`]))
	if validateErr != nil {
		gsgin.GinResponseError(c, validateErr.Error(), nil)
		return
	}
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `smart_link_group_id`, `links`, `open_num`, `open_type`, `process`, `weight`, `combine_type`, `download_finds`, `auto_close_second`, `channel`, `show_cookies`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_smart_link`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	variable, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_smart_link`, `*`, map[string]any{
		`id`:     id,
		`status`: define.SmartLinkStatusNormal,
	}).One()
	gsgin.GinResponseSuccess(c, ``, variable)
}

func validateProcess(process string) error {
	if process == `` {
		return nil
	}
	processList := make([]map[string]any, 0)
	decodeErr := gstool.JsonDecode(process, &processList)
	if decodeErr != nil {
		return errors.New(`解析process失败`)
	}
	for _, processVal := range processList {
		//类型
		processType := cast.ToString(processVal[`type`])
		if processType == `` {
			return errors.New(`type不能为空`)
		}
		//元素选择
		Locator := cast.ToString(processVal[`Locator`])
		switch processType {
		case `click`: //点击
			if Locator == `` {
				return errors.New(`type为click时Locator不能为空`)
			}
		case `input`: //输入
			if cast.ToString(processVal[`value`]) == `` {
				return errors.New(`type为input时value不能为空`)
			}
		case `redirect_uri`: //跳转 保持当前域名
			if cast.ToString(processVal[`value`]) == `` {
				return errors.New(`type为redirect_uri时，value不能为空`)
			}
		}
	}
	return nil
}

// SmartLinkDelete 删除
func SmartLinkDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	dataGMap := gs.NewTransMap(&dataMap)
	if dataGMap.G(`id`).IsZero() {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link`, map[string]any{
			`id`: dataGMap.G(`id`).ToStr(),
		}, map[string]interface{}{
			`status`: define.SmartLinkStatusDelete,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkRunPlaywright 执行 playwright
func SmartLinkRunPlaywright(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	id := cast.ToInt(dataMap[`id`])
	label := cast.ToString(dataMap[`label`])
	if id == 0 || label == `` {
		gsgin.GinResponseError(c, `id和label不能为空`, nil)
		return
	}
	userName := cast.ToString(dataMap[`user_name`])
	password := cast.ToString(dataMap[`password`])
	openNum := max(1, cast.ToInt(dataMap[`open_num`]))
	replaceList := make([]map[string]string, 0)
	gstool.FmtPrintlnLogTime(`开始运行 %d`, openNum)
	for i := 0; i < openNum; i++ {
		go func() {
			runParams, runParamsErr := base.Component.TPlaywright.GetRunParams(id, label, userName, password, openNum, &replaceList)
			gstool.FmtPrintlnLogTime(`初始化结束1`)
			if runParamsErr != nil {
				gstool.FmtPrintlnLogTime(`打开错误 %s`, runParamsErr.Error())
				return
			}
			gstool.FmtPrintlnLogTime(`初始化结束`)
			p := p_playwright.NewPlaywright(runParams, base.Component.TPlaywright.Log)
			openErr := p.Open()
			if openErr != nil {
				gstool.FmtPrintlnLogTime(`错误 %s`, openErr.Error())
			}
		}()
		time.Sleep(time.Second * 2)
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkRunPlaywrightList 获取运行的列表
func SmartLinkRunPlaywrightList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	contextPageList := p_playwright.NewContextList(base.Component.TPlaywright.Log)
	runList := contextPageList.GetPlaywrightRunList()
	gsgin.GinResponseSuccess(c, ``, runList)
}

func SmartLinkPlaywrightVersion(c *gin.Context) {
	pw, pwErr := base.Component.TPlaywright.SmartLinkPlaywrightVersion()
	if pwErr != nil {
		gsgin.GinResponseError(c, `查询失败`+pwErr.Error(), nil)
		return
	}
	//是否在安装中
	isInstall := 0
	if !gstool.FileIsExisted(base.Component.TPlaywright.LockFileFullPath) {
		isInstall = 1
	} else {
		content, _ := gstool.FileGetContent(base.Component.TPlaywright.LockFileFullPath)
		if content == `` {
			isInstall = 1
		}
	}
	gsgin.GinResponseSuccess(c, pw.Version, map[string]any{
		`is_install`: isInstall,
		`version`:    pw.Version,
	})
}
