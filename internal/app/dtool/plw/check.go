package plw

import (
	"fmt"
	"github.com/w896736588/go-tool/gstool"
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

func (h *Check) OutKeyBoolResult() string {
	h.log.Debugf(`OutKey %s`, h.OutKey)
	if h.OutKey == `` {
		return ``
	}
	if strings.Contains(h.Checks, `!=`) { //不等于
		checkList := strings.Split(h.Checks, `!=`)
		if len(checkList) != 2 {
			return fmt.Sprintf(`判断条件格式错误: %s`, h.Checks)
		}
		leftCheck := strings.TrimSpace(checkList[0])
		rightCheck := strings.TrimSpace(checkList[1])
		result := leftCheck != rightCheck
		h.BoolResultMap[h.OutKey] = result
		return fmt.Sprintf(`判断条件: "%s" != "%s"`, leftCheck, rightCheck)
	} else if strings.Contains(h.Checks, `==`) { //等于
		checkList := strings.Split(h.Checks, `==`)
		leftCheck := strings.TrimSpace(checkList[0])
		rightCheck := strings.TrimSpace(checkList[1])
		result := leftCheck == rightCheck
		h.BoolResultMap[h.OutKey] = result
		return fmt.Sprintf(`判断条件: "%s" == "%s"`, leftCheck, rightCheck)
	}
	return fmt.Sprintf(`无匹配操作符: %s`, h.Checks)
}

func (h *Check) AllowCheckKey() bool {
	if h.Checks == `` {
		return true
	}
	//"{is_login}": false,
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
