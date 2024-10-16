package redis_manager

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"log"
	"testing"
)

func TestTest(t *testing.T) {
	if !gstool.FileIsExisted("./playwright.lock") {
		installErr := playwright.Install()
		if installErr != nil {
			gstool.FmtPrintlnLogTime(`安装浏览器核心失败 %s`, installErr.Error())
			return
		}
		createLockErr := gstool.FileCreate(`./`, `playwright.lock`, ``)
		if createLockErr != nil {
			gstool.FmtPrintlnLogTime(`创建lock失败 %s`, createLockErr.Error())
			return
		}
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err.Error())
	}
	//静默打开
	//browser, err := pw.Chromium.Launch()
	//if err != nil {
	//	log.Fatalf("could not launch browser: %v", err)
	//}
	//如果非静默打开
	// 启动 Chromium 浏览器，并设置 Headless 为 false 以显示浏览器窗口
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // 设置为非 Headless 模式
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto("http://common3.xiaokefu.com.cn/api/mobileweb/home?wechatapp_id=211418&channel_id=15399&channel_key=153993h83&kefu_uid=799533513&key=43435hjs2"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	//找到键盘
	if err = page.Click(`.switch-input`); err != nil {
		log.Fatalf("could not fill password input: %v", err)
	}
	//找到输入框
	if err = page.Click(`.input.message-input`); err != nil {
		log.Fatalf("could not fill password input: %v", err)
	}
	if err = page.Fill(`.input.message-input`, "哈哈哈自动发"); err != nil {
		log.Fatalf("could not fill password input: %v", err)
	}
	//找到发送按钮
	buttonSelector := ".fasong_btn" // 替换为实际的按钮选择器
	if err = page.Click(buttonSelector); err != nil {
		log.Fatalf("could not click button: %v", err)
	}
	//if err = browser.Close(); err != nil {
	//	log.Fatalf("could not close browser: %v", err)
	//}
	//if err = pw.Stop(); err != nil {
	//	log.Fatalf("could not stop Playwright: %v", err)
	//}
	gstool.SignalDefault()
}
