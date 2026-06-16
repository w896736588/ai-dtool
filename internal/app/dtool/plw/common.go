package plw

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_curl"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

func buildAccountKey(userName string) string {
	userName = strings.TrimSpace(userName)
	if userName == `` {
		return ``
	}
	return `account_user_` + userName
}

func buildDirectoryMappingKey(smartLinkID int, label, channel, accountKey string) string {
	keyParts := []string{
		fmt.Sprintf(`smart_link_%d`, smartLinkID),
		`label_` + strings.TrimSpace(label),
		`channel_` + channel,
	}
	if accountKey != `` {
		keyParts = append(keyParts, accountKey)
	}
	return strings.Join(keyParts, `_`)
}

func GetRunParams(id int, label, userName, password string, openType int, openNum int, replaceList map[string]string) (*PlaywrightRunParams, error) {
	runParams := &PlaywrightRunParams{}
	if id == 0 {
		return runParams, errors.New(`链接ID不能为空`)
	}
	if label == `` {
		return runParams, errors.New(`链接label不能为空`)
	}
	runParams.Id = id
	runParams.Label = label

	// 优先查询新表 smart_link
	smartLinkItem, itemErr := common.DbMain.Client.QueryBySql(`select * from smart_link where id = ? and status = ?`, id, define.SmartLinkStatusNormal).One()
	if itemErr == nil && len(smartLinkItem) > 0 && cast.ToString(smartLinkItem[`link`]) != `` {
		// 新表格式：直接使用行内字段
		return getRunParamsFromNewTable(id, label, userName, password, openType, openNum, replaceList, smartLinkItem, runParams)
	}

	// 兼容老表 tbl_smart_link（links JSON 格式）
	smartLink, smartLinkErr := common.DbMain.Client.QueryBySql(`select * from tbl_smart_link where id = ? `, id).One()
	if smartLinkErr != nil {
		return runParams, errors.New(smartLinkErr.Error())
	}
	if len(smartLink) == 0 {
		return runParams, errors.New(`不存在的链接`)
	}
	return getRunParamsFromOldTable(id, label, userName, password, openType, openNum, replaceList, smartLink, runParams)
}

// getRunParamsFromNewTable 从新表 smart_link 构建运行参数
func getRunParamsFromNewTable(id int, label, userName, password string, openType int, openNum int, replaceList map[string]string, item map[string]any, runParams *PlaywrightRunParams) (*PlaywrightRunParams, error) {
	runParams.Link = cast.ToString(item[`link`])
	runParams.LinkIdLabel = `link_id_` + cast.ToString(id) + `_label_` + label
	runParams.OpenNum = 0
	runParams.Cookie = cast.ToString(item[`cookie`])
	headerMap := make(map[string]string)
	_ = gstool.JsonDecode(cast.ToString(item[`headers`]), &headerMap)
	runParams.Headers = headerMap
	runParams.BrowserAuthUsername = cast.ToString(item[`browser_auth_username`])
	runParams.BrowserAuthPassword = cast.ToString(item[`browser_auth_password`])
	runParams.DownloadFinds = strings.Split(cast.ToString(item[`download_finds`]), `,`)
	runParams.AutoCloseSecond = cast.ToInt(item[`auto_close_second`])
	runParams.Channel = cast.ToString(item[`channel`])
	runParams.LinkId = GetLinkId(runParams)
	runParams.ShowCookies = make([]ShowCookie, 0)
	_ = gstool.JsonDecode(cast.ToString(item[`show_cookies`]), &runParams.ShowCookies)
	if runParams.Channel == `` {
		runParams.Channel = `chromium`
	}
	runParams.Link = gstool.SReplaces(runParams.Link, map[string]string{
		`{rand}`: p_common.TBaseClient.GetUnique(`link_rand_`),
	})
	runParams.Link = gstool.SReplaces(runParams.Link, replaceList)
	runParams.CombineType = define.CombineTypeFix
	runParams.OpenNum = cast.ToInt(math.Max(1, cast.ToFloat64(openNum)))
	runParams.AccountKey = buildAccountKey(userName)
	runParams.DirectoryMappingKey = buildDirectoryMappingKey(id, label, runParams.Channel, runParams.AccountKey)
	if openType != 0 {
		runParams.OpenType = define.OpenType(openType)
	} else {
		runParams.OpenType = define.OpenType(cast.ToInt(item[`open_type`]))
	}

	processList := make([]map[string]any, 0)
	processId := cast.ToInt(item[`process_id`])
	if processId > 0 {
		processList, _ = common.DbMain.Client.QueryBySql("select * from tbl_smart_link_process_item where status = 1 and smart_link_process_id = ? order by weight asc", processId).All()
	}
	parsedURL, err := url.Parse(runParams.Link)
	if err != nil {
		return runParams, gstool.Error(`解析地址%s失败 %s`, runParams.Link, err.Error())
	}
	runParams.Domain = parsedURL.Host
	runParams.Scheme = parsedURL.Scheme
	replaceList[`{user_name}`] = userName
	replaceList[`{password}`] = password
	if userName != `` {
		runParams.LastIndexLabel = userName
	} else {
		runParams.LastIndexLabel = label
	}
	runParams.ProcessList = processList
	runParams.ReplaceList = replaceList
	runParams.LocatorTimeout = 1000
	runParams.GetPageTimeout = 3000
	runParams.ListenCurls = make(map[string]*p_curl.CurlRun)
	runParams.FilterUris = strings.Split(cast.ToString(item[`filter_uris`]), "\n")
	if runParams.StreamFunc == nil {
		runParams.StreamFunc = func(_, _ string) {}
	}
	return runParams, nil
}

// getRunParamsFromOldTable 从老表 tbl_smart_link（含 links JSON）构建运行参数，保留兼容
func getRunParamsFromOldTable(id int, label, userName, password string, openType int, openNum int, replaceList map[string]string, smartLink map[string]any, runParams *PlaywrightRunParams) (*PlaywrightRunParams, error) {
	linkList := make([]map[string]any, 0)
	linkProcessId := 0
	runParams.DownloadFinds = strings.Split(cast.ToString(smartLink[`download_finds`]), `,`)
	runParams.AutoCloseSecond = cast.ToInt(smartLink[`auto_close_second`])
	runParams.Channel = cast.ToString(smartLink[`channel`])
	runParams.LinkId = GetLinkId(runParams)
	runParams.ShowCookies = make([]ShowCookie, 0)
	_ = gstool.JsonDecode(cast.ToString(smartLink[`show_cookies`]), &runParams.ShowCookies)
	if runParams.Channel == `` {
		runParams.Channel = `chromium`
	}
	decodeErr := gstool.JsonDecode(cast.ToString(smartLink[`links`]), &linkList)
	if decodeErr != nil {
		return runParams, errors.New(decodeErr.Error())
	}
	for _, link := range linkList {
		if cast.ToString(link[`label`]) == label {
			runParams.Link = cast.ToString(link[`link`])
			runParams.LinkIdLabel = `link_id_` + cast.ToString(runParams.Id) + `_label_` + label
			runParams.OpenNum = 0
			runParams.Cookie = cast.ToString(link[`cookie`])
			headerMap := make(map[string]string)
			_ = gstool.JsonDecode(cast.ToString(link[`headers`]), &headerMap)
			runParams.Headers = headerMap
			runParams.BrowserAuthUsername = cast.ToString(link[`browser_auth_username`])
			runParams.BrowserAuthPassword = cast.ToString(link[`browser_auth_password`])
			linkProcessId = cast.ToInt(link[`process_id`])
			break
		}
	}
	if runParams.Link == `` {
		return runParams, errors.New(`链接不存在，检查是否json格式错误`)
	}
	runParams.Link = gstool.SReplaces(runParams.Link, map[string]string{
		`{rand}`: p_common.TBaseClient.GetUnique(`link_rand_`),
	})
	runParams.Link = gstool.SReplaces(runParams.Link, replaceList)
	runParams.CombineType = define.CombineTypeFix
	runParams.OpenNum = cast.ToInt(math.Max(1, cast.ToFloat64(openNum)))
	runParams.AccountKey = buildAccountKey(userName)
	runParams.DirectoryMappingKey = buildDirectoryMappingKey(runParams.Id, label, runParams.Channel, runParams.AccountKey)
	if openType != 0 {
		runParams.OpenType = define.OpenType(openType)
	} else {
		runParams.OpenType = define.OpenType(cast.ToInt(smartLink[`open_type`]))
	}

	processList := make([]map[string]any, 0)
	processId := linkProcessId
	if processId == 0 {
		processId = cast.ToInt(smartLink[`process_id`])
	}
	if processId > 0 {
		processList, _ = common.DbMain.Client.QueryBySql("select * from tbl_smart_link_process_item where status = 1 and smart_link_process_id = ? order by weight asc", processId).All()
	}
	parsedURL, err := url.Parse(runParams.Link)
	if err != nil {
		return runParams, gstool.Error(`解析地址%s失败 %s`, runParams.Link, err.Error())
	}
	runParams.Domain = parsedURL.Host
	runParams.Scheme = parsedURL.Scheme
	replaceList[`{user_name}`] = userName
	replaceList[`{password}`] = password
	if userName != `` {
		runParams.LastIndexLabel = userName
	} else {
		runParams.LastIndexLabel = label
	}
	runParams.ProcessList = processList
	runParams.ReplaceList = replaceList
	runParams.LocatorTimeout = 1000
	runParams.GetPageTimeout = 3000
	runParams.ListenCurls = make(map[string]*p_curl.CurlRun)
	runParams.FilterUris = strings.Split(cast.ToString(smartLink[`filter_uris`]), "\n")
	if runParams.StreamFunc == nil {
		runParams.StreamFunc = func(_, _ string) {}
	}
	return runParams, nil
}

func GetLinkId(runParams *PlaywrightRunParams) string {
	return fmt.Sprintf(`link_id_%d`, runParams.Id)
}

// ShowCookieTip 展示cookie中的某个值
func ShowCookieTip(page *playwright.Page, runParams *PlaywrightRunParams) {
	if len(runParams.ShowCookies) == 0 {
		return
	}
	replaceList := make([]ShowCookie, 0)
	for _, config := range runParams.ShowCookies {
		if gstool.SContains(strings.ToLower((*page).URL()), config.DomainList) {
			replaceList = append(replaceList, config)
		}
	}
	if len(replaceList) == 0 {
		return
	}
	config := gstool.JsonEncode(replaceList)
	content := p_common.TJasClient.Get(`p_js`, `info.js`)
	content = gstool.SReplaces(content, map[string]string{
		`{config}`: config,
	})
	_, _ = (*page).Evaluate(content)
}

func ValueFormat(name, value string, runParams *PlaywrightRunParams) string {
	if value == `` {
		return value
	}
	replaceValue := gstool.SReplaces(value, map[string]string{
		`{rand}`:   p_common.TBaseClient.GetUnique(`input_rand_`),
		`{domain}`: runParams.Domain,
	})

	//针对输入进行替换
	replaceValue = gstool.SReplaces(replaceValue, runParams.ReplaceList)
	return replaceValue
}
