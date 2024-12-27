package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gs"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"net/url"
	"sync"
	"time"
)

// SmartLinkUpWebkit 更新核心
func SmartLinkUpWebkit(c *gin.Context) {
	installErr := playwright.Install()
	if installErr != nil {
		gsgin.GinResponseError(c, `安装浏览器核心失败 %s`, installErr.Error())
		return
	}
	gsgin.GinResponseSuccess(c, `更新浏览器核心成功`, ``)
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
	//if cast.ToInt(dataMap[`smart_link_group_id`]) == 0 {
	//	gsgin.GinResponseError(c, `组id不能为空 `, nil)
	//	return
	//}
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `smart_link_group_id`, `links`, `open_num`, `open_type`, `process`, `weight`})
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

// SmartLinkProcessAdd 新增子操作
func SmartLinkProcessAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `process_type`, `smart_link_id`, `selecter`, `weight`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_smart_link_process`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_process`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkProcessDel 删除子操作
func SmartLinkProcessDel(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	dataGMap := gs.NewTransMap(&dataMap)
	if dataGMap.G(`id`).IsZero() {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_process`, map[string]any{
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
	link := cast.ToString(dataMap[`link`])
	openNum := cast.ToInt(dataMap[`open_num`])
	openType := cast.ToInt(dataMap[`open_type`])
	process := cast.ToString(dataMap[`process`])
	if link == `` {
		gsgin.GinResponseError(c, `链接不存在，检查是否json格式错误`, nil)
		return
	}
	processList := make([]map[string]any, 0)
	if process != `` {
		decodeErr := gstool.JsonDecode(process, &processList)
		if decodeErr != nil {
			gsgin.GinResponseError(c, `配置失败`+decodeErr.Error(), nil)
			return
		}
	}

	for i := 0; i < openNum; i++ {
		go func() {
			openErr := openBrowserPlaywright(openType, link, processList, dataMap)
			if openErr != nil {
				gstool.FmtPrintlnLogTime(`错误 %s`, openErr.Error())
			}
		}()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkRunPlaywrightList 获取运行的列表
func SmartLinkRunPlaywrightList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	runList := make([]map[string]any, 0)
	for uniKey, runInfo := range base.Component.TSmartLink.PageList {
		runList = append(runList, map[string]any{
			`name`:   runInfo.Value,
			`unikey`: uniKey,
		})
	}
	gsgin.GinResponseSuccess(c, ``, runList)
}

// SmartLinkPlaywrightForward 唤醒 注意，如果点击了最小化，那么弹不出来
func SmartLinkPlaywrightForward(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToString(dataMap[`unikey`]) == `` {
		gsgin.GinResponseError(c, `unikey不能为空`, nil)
		return
	}
	page := base.Component.TSmartLink.PageList[cast.ToString(dataMap[`unikey`])]
	if page == nil {
		gsgin.GinResponseError(c, `窗口已不存在`, nil)
		return
	}
	gstool.FmtPrintlnLogTime(`活跃的Pid %v`, page.BrowserPid)
	err := base.Component.TSmartLink.SetForegroundWindowPid(page.BrowserPid)
	if err != nil {
		gsgin.GinResponseSuccess(c, `唤醒失败`+err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
	return
}

// SmartLinkPlaywrightVersion 获取浏览器核心版本
func SmartLinkPlaywrightVersion(c *gin.Context) {
	pw, _ := playwright.NewDriver()
	lockFileName := `playwright.lock`
	lockFileFullPath := base.Component.Env.RootPath + `/` + lockFileName
	if !gstool.FileIsExisted(lockFileFullPath) {
		go install(pw.Version, lockFileFullPath)
	} else {
		content, contentErr := gstool.FileGetContent(lockFileFullPath)
		if contentErr != nil {
			gstool.FmtPrintlnLogTime(`获取文件内容失败 %s`, contentErr.Error())
		} else if content != pw.Version {
			go install(pw.Version, lockFileFullPath)
		} else {
			gstool.FmtPrintlnLogTime(`浏览器核心最新版本为：%s ，当前安装版本为：%s,不需要进行更新`, pw.Version, content)
		}
	}
	gsgin.GinResponseSuccess(c, pw.Version, nil)
	return
}

func install(version, lockFileFullPath string) {
	if base.Component.TSmartLink.IsInstall {
		gstool.FmtPrintlnLogTime(`正在安装中..`)
		return
	}
	defer func() {
		base.Component.TSmartLink.IsInstall = false
	}()
	base.Component.TSmartLink.IsInstall = true
	_ = gstool.FilePutContentCover(lockFileFullPath, version)
	gstool.FmtPrintlnLogTime(`开始安装浏览器核心(只安装chrome),大约几分钟时间`)
	err := playwright.Install(&playwright.RunOptions{Browsers: []string{`chromium`}})
	if err != nil {
		gstool.FmtPrintlnLogTime(`安装浏览器核心失败 %s`, err.Error())
		_ = gstool.FileDelete(lockFileFullPath)
	} else {
		gstool.FmtPrintlnLogTime(`安装完成`)
	}
}

// 打开浏览器
func openBrowserPlaywright(openType int, link string, processList []map[string]any, dataMap map[string]any) error {
	browserAuthUsername := cast.ToString(dataMap[`browser_auth_username`])
	browserAuthPassword := cast.ToString(dataMap[`browser_auth_password`])
	value := cast.ToString(dataMap[`value`])
	page, pageErr := base.Component.TSmartLink.GetPage(openType, link, value, browserAuthUsername, browserAuthPassword)
	if pageErr != nil {
		return pageErr
	}
	for _, processVal := range processList {
		time.Sleep(time.Millisecond * 200)
		//类型
		processType := cast.ToString(processVal[`type`])
		//如果不存在
		notExistLocator := cast.ToString(processVal[`not_exist_Locator`])
		//元素选择
		Locator := cast.ToString(processVal[`Locator`])
		gstool.FmtPrintlnLogTime(`查找元素 %s`, Locator)
		//链接
		redirectUri := cast.ToString(processVal[`uri`])
		//操作描述
		tip := cast.ToString(processVal[`tip`])
		//等待时间
		waitSecond := cast.ToFloat64(processVal[`wait_second`])
		if waitSecond == 0 {
			waitSecond = cast.ToFloat64(1)
		}
		waitSecond = waitSecond * 1000
		// 等待导航完成
		waitErr := (*page.Page).WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateDomcontentloaded, //三种LoadStateNetworkidle 网络加载最低程度 LoadStateDomcontentloaded DOM加载完成
		})
		if waitErr != nil {
			gstool.FmtPrintlnLogTime("等待页面 DOM 内容加载完成失败: %s", waitErr.Error())
		}

		base.Component.TSmartLink.AddTipMsg(*page.Page, tip)
		switch processType {
		case `window_max`: //窗口最大化
			base.Component.TSmartLink.SetWindowMax(page.BrowserPid)
		case `wait_navigation`: //等待导航完成
			waitUrlErr := (*page.Page).WaitForURL((*page.Page).URL())
			if waitUrlErr != nil {
				return waitUrlErr
			}
			//nodePid := base.Component.TSmartLink.FindNewPidByName(`node.exe`)
			//browserPid := base.Component.TSmartLink.FindPidChildLastPid(nodePid)
			//打开新页卡时，进程ID不变，但是活跃的标签变了，也不会变为活跃，这可能跟浏览器内部处理有关
			//但是手动打开一个新页卡，好像又可以变为活跃 先不研究了
			//gstool.FmtPrintlnLogTime(`node pid %d 新打开的页卡进程 %d`, nodePid, browserPid)
			//base.Component.TSmartLink.PageList[page.CreateTimeDesc].BrowserPid = cast.ToInt(browserPid)
			//监听控制台
			//go listenDevTool(*page.Context, *page.Page, *page.Browser)
		case `click`: //点击
			click(Locator, notExistLocator, waitSecond, *page.Page)
			break
		case `input`: //输入
			inputValue := cast.ToString(processVal[`value`])
			inputValue = gstool.StringReplaces(inputValue, map[string]string{
				`{user_name}`: cast.ToString(dataMap[`user_name`]),
				`{password}`:  cast.ToString(dataMap[`password`]),
			})
			inputSelecter := (*page.Page).Locator(Locator)
			selectorLoaderWaitErr := inputSelecter.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: &waitSecond,
			})
			if selectorLoaderWaitErr == nil {
				inputErr := inputSelecter.Fill(inputValue)
				if inputErr != nil {
					gstool.FmtPrintlnLogTime("无法将元素转换为输入框: %v", inputErr.Error())
				}
			}
			break
		case `redirect_uri`: //跳转 保持当前域名
			currentURL := (*page.Page).URL()
			parsedURL, err := url.Parse(currentURL)
			if err != nil {
				gstool.FmtPrintlnLogTime("could not parse URL: %v", err)
			}
			domain := parsedURL.Scheme + `://` + parsedURL.Host
			gstool.FmtPrintlnLogTime(`跳转地址 %s`, domain+redirectUri)
			if _, goErr := (*page.Page).Goto(domain + redirectUri); goErr != nil {
				return goErr
			}
		}
	}
	//if err = page.Click(`.switch-input`); err != nil {
	//	log.Fatalf("could not fill password input: %v", err)
	//}
	//if err = browser.Close(); err != nil {
	//	log.Fatalf("could not close browser: %v", err)
	//}
	//if err = pw.Stop(); err != nil {
	//	log.Fatalf("could not stop Playwright: %v", err)
	//}
	return nil
}

// 点击
func click(Locator, notExistLocator string, waitSecond float64, page playwright.Page) {
	gstool.FmtPrintlnLogTime(`click %s`, notExistLocator)
	if notExistLocator == `` { //等待点击
		selectorLoader := page.Locator(Locator)
		selectorLoaderWaitErr := selectorLoader.WaitFor(playwright.LocatorWaitForOptions{
			Timeout: &waitSecond,
			State:   playwright.WaitForSelectorStateVisible,
		})
		if selectorLoaderWaitErr == nil {
			gstool.FmtPrintlnLogTime(`找到元素%s，点击`, Locator)
			clickErr := selectorLoader.Click()
			if clickErr != nil {
				gstool.FmtPrintlnLogTime(`等待元素后 点击失败 %s`, clickErr.Error())
			}
		}
	} else { //等待或者某个元素存在
		wait := sync.WaitGroup{}
		wait.Add(1)
		//等待某个元素存在时直接走
		go func() {
			existLoader := page.Locator(notExistLocator)
			existLoaderErr := existLoader.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: &waitSecond,
				State:   playwright.WaitForSelectorStateVisible,
			})
			if existLoaderErr == nil { //存在 那么直接返回往下走
				gstool.FmtPrintlnLogTime(`找到了not exit `)
				wait.Done()
			}
		}()
		//等待原本要找的元素出现并点击
		go func() {
			selectorLoader := page.Locator(Locator)
			selectorLoaderWaitErr := selectorLoader.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: &waitSecond,
				State:   playwright.WaitForSelectorStateVisible,
			})
			if selectorLoaderWaitErr == nil {
				gstool.FmtPrintlnLogTime(`找到元素%s，点击`, Locator)
				clickErr := selectorLoader.Click()
				if clickErr != nil {
					gstool.FmtPrintlnLogTime(`等待元素后 点击失败 %s`, clickErr.Error())
					wait.Done()
				}
			}
		}()
		wait.Wait()
	}
}

func registerJs(page playwright.Page) {
	keyEventPath := base.Component.Env.RootPath + `/internal/pkg/js_script/key_event.js`
	gstool.FmtPrintlnLogTime(`加载初始化js %s`, keyEventPath)
	if err := page.AddInitScript(playwright.Script{
		Path: &keyEventPath,
	}); err != nil {
		gstool.FmtPrintlnLogTime(`err %s`, err.Error())
	}
}

// SetTitle 设置title
func SetTitle(page playwright.Page, title string) {
	_, _ = page.Evaluate(`(function() {
			setTimeout(function() {
				document.title = "` + title + `";
			}, 100); 
		})();`)
}

func listenDevTool(context playwright.BrowserContext, page playwright.Page, browser playwright.Browser) {
	// 使用 CDP 监听控制台消息
	cdpSession, err := context.NewCDPSession(page)
	if err != nil {
		gstool.FmtPrintlnLogTime("could not create CDP session: %v", err)
		return
	}

	// 启用控制台消息监听
	if _, err := cdpSession.Send("Runtime.enable", nil); err != nil {
		gstool.FmtPrintlnLogTime("could not enable Runtime: %v", err)
		return
	}

	// 监听控制台消息
	cdpSession.On("Runtime.consoleAPICalled", func(event interface{}) {
		gstool.FmtPrintlnLogTime(`接受到控制台信息 %s`, gstool.JsonEncode(event))
		params := _struct.DevToolEvent{}
		_ = gstool.JsonDecode(gstool.JsonEncode(event), &params)
		for _, arg := range params.Args {
			if arg.Value == "F12 key pressed" {

			}
		}
	})

}
