package base

import (
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"gitee.com/Sxiaobai/gs/v2/gsapi"
	"github.com/spf13/cast"
	"regexp"
	"strings"
	"time"
)

type TGitlab struct {
	GitLab  gsapi.GsGitLab
	Author  string
	LogFunc func(string)
}

type Combine struct {
	Message string
	Status  string
}

func (h *TGitlab) AssignDayLogs(start, end string) ([]Combine, error) {
	startDay, _ := gstool.TimeStringToUnix(start, `Y-m-d H:i:s`)
	endDay, _ := gstool.TimeStringToUnix(end, `Y-m-d H:i:s`)
	return h.getLogs(startDay, endDay)
}

func (h *TGitlab) GetTodayLogs() ([]Combine, error) {
	now := time.Now()
	startDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endDay := startDay.AddDate(0, 0, 1).Add(-time.Nanosecond)
	return h.getLogs(startDay, endDay)
}

func (h *TGitlab) getLogs(startDay, endDay time.Time) ([]Combine, error) {
	projectMap := make(map[string]string)
	perPage := 20
	startTimestamp := startDay.Unix()
	endTimestamp := endDay.Unix()
	combineList := make([]Combine, 0)
	//所有有权限项目
	for page := 1; page < 10; page++ {
		projectParam := gsapi.GsGitLabParam{
			State:   "",
			Sort:    "",
			Page:    page,
			PerPage: perPage,
			RefName: "",
		}
		projectList, resErr := h.GitLab.GetProjects(projectParam)
		if resErr != nil {
			return combineList, resErr
		}
		for _, project := range projectList {
			projectMap[cast.ToString(project[`id`])] = cast.ToString(project[`name`])
		}
		if len(projectList) < perPage {
			break
		}
	}
	h.pushLog(`获取完项目列表 共：` + cast.ToString(len(projectMap)) + `个`)
	masterCommits := make([]string, 0)
	for projectId, projectName := range projectMap {
		err := h.checkCommits(projectId, projectName, h.Author, perPage, startTimestamp, endTimestamp, &combineList, &masterCommits)
		if err != nil {
			return combineList, err
		}
		err = h.checkMerges(projectId, projectName, h.Author, perPage, startTimestamp, endTimestamp, &combineList, &masterCommits)
		if err != nil {
			return combineList, err
		}
	}
	return combineList, nil
}

func (h *TGitlab) checkMerges(projectId, projectName, author string, perPage int, startTimestamp, endTimestamp int64, combineList *[]Combine, masterCommits *[]string) error {
	for page := 1; page < 100; page++ {
		gitLabParam := gsapi.GsGitLabParam{
			State:   "opened",
			Sort:    "desc",
			Page:    page,
			PerPage: perPage,
			RefName: "",
		}
		mergeList, resErr := h.GitLab.GetMerges(projectId, gitLabParam)
		if resErr != nil {
			return resErr
		}
		for _, merge := range mergeList {
			sourceBranch := cast.ToString(merge[`source_branch`])
			title := cast.ToString(merge[`title`])
			authorJoin, authorCommit, otherJoin, selfTest, err := h.checkMergeUserOp(projectId, sourceBranch, author, startTimestamp, endTimestamp, masterCommits)
			if err != nil {
				return err
			}
			status := h.getStatus(authorJoin, authorCommit, otherJoin, selfTest)
			if status != `` {
				combine := Combine{
					Message: title,
					Status:  status,
				}
				*combineList = append(*combineList, combine)
				h.LogFunc(gstool.JsonEncode(combine))
			}
		}
		if len(mergeList) < perPage {
			break
		}
	}
	return nil
}

func (h *TGitlab) getStatus(authorJoin, authorCommit, otherJoin, selfTest bool) string {
	if !authorJoin {
		return ``
	}
	if otherJoin { //其他人参与
		if selfTest { //自测完
			if authorCommit { //作者参与改动
				return `开发对接自测完`
			} else {
				return `对接自测完`
			}
		} else {
			if authorCommit { //作者参与改动
				return `开发对接`
			} else {
				return `对接`
			}
		}
	} else { //其他人不参与
		if selfTest { //自测完
			if authorCommit { //作者参与改动
				return `自测完`
			} else {
				return `自测完`
			}
		} else {
			if authorCommit { //作者参与改动
				return `开发`
			} else {
				return `` //其他人不参与 没有自测 没有开发
			}
		}
	}
}

func (h *TGitlab) checkCommits(projectId, projectName, author string,
	perPage int, startTimestamp, endTimestamp int64, combineList *[]Combine, masterCommits *[]string) error {
	for page := 1; page < 100; page++ {
		gitLabParam := gsapi.GsGitLabParam{
			State:   "",
			Sort:    "desc",
			Page:    page,
			PerPage: perPage,
			RefName: "",
		}
		commitList, resErr := h.GitLab.GetProjectCommits(projectId, gitLabParam)
		if resErr != nil {
			return resErr
		}
		h.pushLog(fmt.Sprintf(`获取%scommit 共：%d条`, projectName, len(commitList)))
		boolBroken := false
		for _, commit := range commitList {
			id := cast.ToString(commit[`id`])
			*masterCommits = append(*masterCommits, id)
			createdAt := cast.ToString(commit[`created_at`])
			beijingTime, beijingTimeErr := h.gBeijingTime(createdAt)
			if beijingTimeErr != nil {
				return errors.New(`解析时间报错 ` + beijingTimeErr.Error())
			}
			if beijingTime.Unix() < startTimestamp { //小于最小时间 那就直接退出
				boolBroken = true
				break
			}
			if beijingTime.Unix() > endTimestamp { //大于结束时间 继续循环
				continue
			}
			message := cast.ToString(commit[`message`])
			title := cast.ToString(commit[`title`])
			if strings.Contains(title, `into 'master'`) {
				if strings.Contains(message, author) {
					combine := Combine{
						Message: message,
						Status:  `已上线`,
					}
					*combineList = append(*combineList, combine) //收集合并
					h.LogFunc(gstool.JsonEncode(combine))
				} else {
					authorJoin, _, _, _, err := h.checkMergeUserOp(projectId, h.getBranchName(title), author, startTimestamp, endTimestamp, masterCommits)
					if err != nil {
						return err
					}
					if authorJoin {
						combine := Combine{
							Message: message,
							Status:  `已上线`,
						}
						*combineList = append(*combineList, combine)
						h.LogFunc(gstool.JsonEncode(combine))
					}
				}
			}
		}
		if boolBroken {
			break
		}
		if len(commitList) < perPage {
			break
		}
	}
	return nil
}

func (h *TGitlab) getBranchName(title string) string {
	re := regexp.MustCompile(`Merge branch '([^']+)' into`)
	matches := re.FindStringSubmatch(title)
	if len(matches) > 1 {
		return matches[1]
	} else {
		return ``
	}
}

// 检查某个分支 在某个范围内是否有某个用户的提交
func (h *TGitlab) checkMergeUserOp(projectId, branchName, author string, startTimestamp,
	endTimestamp int64, masterCommits *[]string) (bool, bool, bool, bool, error) {
	authorJoin := false   //author 是否参与了
	authorCommit := false //author 是否提交commit了，不算merge
	otherJoin := false    //其他人是否参与了
	selfTest := false     //是否自测了
	gitLabParam := gsapi.GsGitLabParam{
		State:   "",
		Sort:    "desc",
		Page:    1,
		PerPage: 20,
		RefName: branchName,
	}
	commitList, resErr := h.GitLab.GetProjectCommits(projectId, gitLabParam)
	if resErr != nil {
		return false, false, false, false, resErr
	}
	h.pushLog(fmt.Sprintf(`获取%scommit 共：%d条`, branchName, len(commitList)))
	for _, commit := range commitList {
		id := cast.ToString(commit[`id`])
		if gstool.ArrayFindValueIndex(masterCommits, id) >= 0 {
			continue
		}
		authorName := cast.ToString(commit[`author_name`])
		createdAt := cast.ToString(commit[`created_at`])
		message := cast.ToString(commit[`message`])
		beijingTime, beijingTimeErr := h.gBeijingTime(createdAt)
		if beijingTimeErr != nil {
			return false, false, false, false, beijingTimeErr
		}
		if beijingTime.Unix() < startTimestamp { //小于最小时间 那就直接退出
			break
		}
		if beijingTime.Unix() > endTimestamp { //大于结束时间 继续循环
			continue
		}
		if h.isTest(message) {
			selfTest = true
		}
		if strings.Contains(authorName, author) {
			authorJoin = true
			if !h.isMergeBranch(message) {
				authorCommit = true
			}
		} else {
			otherJoin = true
		}
	}
	return authorJoin, authorCommit, otherJoin, selfTest, nil
}

func (h *TGitlab) isMergeBranch(message string) bool {
	if strings.Contains(message, `Merge branch`) {
		return true
	}
	return false
}

func (h *TGitlab) isTest(message string) bool {
	if strings.Contains(message, `自测`) || strings.Contains(message, `测完`) ||
		strings.Contains(message, `测试`) {
		return true
	}
	return false
}

func (h *TGitlab) gBeijingTime(utcTimeStr string) (time.Time, error) {
	utcTime, err := time.Parse(time.RFC3339, utcTimeStr)
	if err != nil {
		return time.Now(), errors.New(err.Error())
	}

	loc, locErr := time.LoadLocation("Asia/Shanghai")
	if locErr != nil {
		return time.Now(), locErr
	}
	return utcTime.In(loc), nil
}

func (h *TGitlab) pushLog(msg string) {
	if h.LogFunc != nil {
		h.LogFunc(msg + "  ")
	}
}
