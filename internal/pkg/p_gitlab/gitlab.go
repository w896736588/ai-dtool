package p_gitlab

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsapi"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// MergeMainBranchs 已上线的记录 只要包含以下就可以
var MergeMainBranchs = []string{
	`master`, `main`,
}

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
	perPage := 20
	startTimestamp := startDay.Unix()
	endTimestamp := endDay.Unix()
	combineList := make([]Combine, 0)
	combineDedup := make(map[string]struct{})
	//所有有权限项目
	projectIds := make([]string, 0, perPage*2)
	projectNameById := make(map[string]string)
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
			projectId := cast.ToString(project[`id`])
			if _, exist := projectNameById[projectId]; exist {
				continue
			}
			projectIds = append(projectIds, projectId)
			projectNameById[projectId] = cast.ToString(project[`name`])
		}
		if len(projectList) < perPage {
			break
		}
	}
	h.pushLog(`获取完项目列表 共：` + cast.ToString(len(projectIds)) + `个`)

	sort.Strings(projectIds)
	for _, projectId := range projectIds {
		projectName := projectNameById[projectId]
		if !strings.Contains(projectName, `chatwiki`) {
			continue
		}
		masterCommitSet := make(map[string]struct{}, perPage*2)
		err := h.checkCommits(projectId, projectName, h.Author, perPage, startTimestamp, endTimestamp, &combineList, combineDedup, masterCommitSet)
		if err != nil {
			return combineList, err
		}
		err = h.checkMerges(projectId, projectName, h.Author, perPage, startTimestamp, endTimestamp, &combineList, combineDedup, masterCommitSet)
		if err != nil {
			return combineList, err
		}
	}
	sort.SliceStable(combineList, func(i, j int) bool {
		if combineList[i].Status == combineList[j].Status {
			return combineList[i].Message < combineList[j].Message
		}
		return combineList[i].Status < combineList[j].Status
	})
	return combineList, nil
}

func (h *TGitlab) checkMerges(projectId, projectName, author string, perPage int, startTimestamp, endTimestamp int64, combineList *[]Combine, combineDedup map[string]struct{}, masterCommitSet map[string]struct{}) error {
	for page := 1; page < 100; page++ {
		gitLabParam := gsapi.GsGitLabParam{
			State:   "all",
			Sort:    "desc",
			Page:    page,
			PerPage: perPage,
			RefName: "",
		}
		mergeList, resErr := h.GitLab.GetMerges(projectId, gitLabParam)
		if resErr != nil {
			return resErr
		}
		boolBroken := false
		for _, merge := range mergeList {
			sourceBranch := cast.ToString(merge[`source_branch`])
			title := cast.ToString(merge[`title`])

			createdAtUnix, createdAtOk := h.getUnixTime(merge[`created_at`])
			updatedAtUnix, updatedAtOk := h.getUnixTime(merge[`updated_at`])
			mergedAtUnix, mergedAtOk := h.getUnixTime(merge[`merged_at`])
			userCreated := strings.Contains(h.getMergeAuthor(merge), author)

			relevantByTime := (createdAtOk && createdAtUnix >= startTimestamp && createdAtUnix <= endTimestamp) ||
				(updatedAtOk && updatedAtUnix >= startTimestamp && updatedAtUnix <= endTimestamp) ||
				(mergedAtOk && mergedAtUnix >= startTimestamp && mergedAtUnix <= endTimestamp)
			if !relevantByTime {
				if updatedAtOk && updatedAtUnix < startTimestamp {
					boolBroken = true
					break
				}
				continue
			}

			authorJoin, authorCommit, otherJoin, selfTest, err := h.checkMergeUserOp(projectId, sourceBranch, author, startTimestamp, endTimestamp, masterCommitSet)
			if err != nil {
				return err
			}
			status := h.getStatus(authorJoin, authorCommit, otherJoin, selfTest)
			if status == `` && userCreated && createdAtOk && createdAtUnix >= startTimestamp && createdAtUnix <= endTimestamp {
				status = `创建`
			}
			if status != `` {
				combine := Combine{
					Message: title,
					Status:  status,
				}
				h.addCombine(combineList, combineDedup, combine)
			}
		}
		if boolBroken {
			break
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
	perPage int, startTimestamp, endTimestamp int64, combineList *[]Combine, combineDedup map[string]struct{}, masterCommitSet map[string]struct{}) error {
	re := regexp.MustCompile(`into\s+['"]([^'"]+)['"]`)
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
		h.pushLog(fmt.Sprintf(`获取project:%s commit:%d条`, projectName, len(commitList)))
		boolBroken := false
		for _, commit := range commitList {
			id := cast.ToString(commit[`id`])
			masterCommitSet[id] = struct{}{}
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
			if h.isMergeIntoMain(title, re) {
				if strings.Contains(message, author) {
					combine := Combine{
						Message: message,
						Status:  `已上线`,
					}
					h.addCombine(combineList, combineDedup, combine)
				} else {
					branchName := h.getBranchName(title)
					if branchName == `` {
						continue
					}
					authorJoin, _, _, _, err := h.checkMergeUserOp(projectId, branchName, author, startTimestamp, endTimestamp, masterCommitSet)
					if err != nil {
						return err
					}
					if authorJoin {
						combine := Combine{
							Message: message,
							Status:  `已上线`,
						}
						h.addCombine(combineList, combineDedup, combine)
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
	endTimestamp int64, masterCommitSet map[string]struct{}) (bool, bool, bool, bool, error) {
	authorJoin := false   //author 是否参与了
	authorCommit := false //author 是否提交commit了，不算merge
	otherJoin := false    //其他人是否参与了
	selfTest := false     //是否自测了
	if branchName == `` {
		return false, false, false, false, nil
	}
	total := 0
	for page := 1; page < 100; page++ {
		gitLabParam := gsapi.GsGitLabParam{
			State:   "",
			Sort:    "desc",
			Page:    page,
			PerPage: 50,
			RefName: branchName,
		}
		commitList, resErr := h.GitLab.GetProjectCommits(projectId, gitLabParam)
		if resErr != nil {
			return false, false, false, false, resErr
		}
		total += len(commitList)
		boolBroken := false
		for _, commit := range commitList {
			id := cast.ToString(commit[`id`])
			if _, exist := masterCommitSet[id]; exist {
				continue
			}
			authorName := cast.ToString(commit[`author_name`])
			committerName := cast.ToString(commit[`committer_name`])
			createdAt := cast.ToString(commit[`created_at`])
			message := cast.ToString(commit[`message`])
			beijingTime, beijingTimeErr := h.gBeijingTime(createdAt)
			if beijingTimeErr != nil {
				return false, false, false, false, beijingTimeErr
			}
			if beijingTime.Unix() < startTimestamp {
				boolBroken = true
				break
			}
			if beijingTime.Unix() > endTimestamp {
				continue
			}
			if h.isTest(message) {
				selfTest = true
			}
			if strings.Contains(authorName, author) || strings.Contains(committerName, author) {
				authorJoin = true
				if !h.isMergeBranch(message) {
					authorCommit = true
				}
			} else {
				otherJoin = true
			}
		}
		if boolBroken || len(commitList) < 50 {
			break
		}
	}
	h.pushLog(fmt.Sprintf(`获取%scommit 共：%d条`, branchName, total))
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

func (h *TGitlab) addCombine(combineList *[]Combine, combineDedup map[string]struct{}, combine Combine) {
	key := combine.Status + `|` + combine.Message
	if _, exist := combineDedup[key]; exist {
		return
	}
	combineDedup[key] = struct{}{}
	*combineList = append(*combineList, combine)
	if h.LogFunc != nil {
		h.LogFunc(gstool.JsonEncode(combine))
	}
}

func (h *TGitlab) isMergeIntoMain(title string, re *regexp.Regexp) bool {
	matches := re.FindStringSubmatch(title)
	if len(matches) < 2 {
		return false
	}
	target := matches[1]
	for _, mainBranch := range MergeMainBranchs {
		if target == mainBranch || strings.Contains(target, mainBranch) {
			return true
		}
	}
	return false
}

func (h *TGitlab) getUnixTime(v any) (int64, bool) {
	s := cast.ToString(v)
	if s == `` || s == `<nil>` {
		return 0, false
	}
	t, err := h.gBeijingTime(s)
	if err != nil {
		return 0, false
	}
	return t.Unix(), true
}

func (h *TGitlab) getMergeAuthor(merge map[string]any) string {
	if v, ok := merge[`author_name`]; ok {
		if s := cast.ToString(v); s != `` && s != `<nil>` {
			return s
		}
	}
	if v, ok := merge[`author`]; ok {
		if m, ok := v.(map[string]any); ok {
			if s := cast.ToString(m[`name`]); s != `` && s != `<nil>` {
				return s
			}
			if s := cast.ToString(m[`username`]); s != `` && s != `<nil>` {
				return s
			}
		}
	}
	return ``
}
