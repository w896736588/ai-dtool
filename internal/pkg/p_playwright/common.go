package p_playwright

import (
	"dev_tool/base"
	"github.com/playwright-community/playwright-go"
)

func AddTipMsg(page *playwright.Page, tip string) {
	go base.Component.TPlaywright.AddTipMsg(page, tip)
}
