package p_playwright

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

type Check struct {
	OutKey        string
	Checks        string
	BoolResultMap map[string]bool //map是传递的指针值 当作引用传值使用
	log           *gstool.GsSlog
}

func NewCheck(outKey, checks string, boolResultMap map[string]bool, log *gstool.GsSlog) *Check {
	return &Check{
		OutKey:        outKey,
		Checks:        checks,
		BoolResultMap: boolResultMap,
		log:           log,
	}
}

func (h *Check) OutKeyBoolResult() {
	if h.OutKey == `` {
		return
	}
	if strings.Contains(h.Checks, `!=`) { //不等于
		checkList := strings.Split(h.Checks, `!=`)
		if len(checkList) != 2 {
			return
		}
		leftCheck := strings.TrimSpace(checkList[0])
		rightCheck := strings.TrimSpace(checkList[1])
		if leftCheck != rightCheck {
			h.BoolResultMap[h.OutKey] = true
		} else {
			h.BoolResultMap[h.OutKey] = false
		}
	} else if strings.Contains(h.Checks, `==`) { //等于
		checkList := strings.Split(h.Checks, `==`)
		leftCheck := strings.TrimSpace(checkList[0])
		rightCheck := strings.TrimSpace(checkList[1])
		if leftCheck != rightCheck {
			h.BoolResultMap[h.OutKey] = false
		} else {
			h.BoolResultMap[h.OutKey] = true
		}
	}
}

func (h *Check) AllowCheckKey() bool {
	if h.Checks == `` {
		return true
	}
	checkList := strings.Split(h.Checks, `&&`)
	for _, checkKeyVal := range checkList {

		if strings.HasPrefix(checkKeyVal, `!`) { //不等于时 等于了 那么跳过
			if h.BoolResultMap[checkKeyVal[1:]] == true {
				return false
			}
		} else if !h.BoolResultMap[checkKeyVal] { //等于时  不等于了 那么跳过
			return false
		}
	}
	return true
}
