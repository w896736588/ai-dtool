package p_gitlab

import (
	"errors"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsapi"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

// MergeMainBranchs 已上线的记录，只要目标分支包含以下关键字即可判定。
var MergeMainBranchs = []string{
	"master", "main",
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

type mergeUserOpSummary struct {
	authorJoin                 bool
	authorCommitToday          bool
	otherCommitToday           bool
	authorMergeLikeCommitToday bool
}

func (s mergeUserOpSummary) branchActiveToday() bool {
	return s.authorCommitToday || s.otherCommitToday || s.authorMergeLikeCommitToday
}

func (h *TGitlab) AssignDayLogs(start, end string) ([]Combine, error) {
	startDay, _ := gstool.TimeStringToUnix(start, "Y-m-d H:i:s")
	endDay, _ := gstool.TimeStringToUnix(end, "Y-m-d H:i:s")
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

	projectIDs := make([]string, 0, perPage*2)
	projectNameByID := make(map[string]string)
	for page := 1; page < 20; page++ {
		projectParam := gsapi.GsGitLabParam{Page: page, PerPage: perPage}
		projectList, resErr := h.GitLab.GetProjects(projectParam)
		if resErr != nil {
			return combineList, resErr
		}
		for _, project := range projectList {
			projectID := cast.ToString(project["id"])
			if _, exist := projectNameByID[projectID]; exist {
				continue
			}
			projectIDs = append(projectIDs, projectID)
			projectNameByID[projectID] = cast.ToString(project["name"])
		}
		if len(projectList) < perPage {
			break
		}
	}
	h.pushLog("获取项目列表，总计: " + cast.ToString(len(projectIDs)))

	sort.Strings(projectIDs)
	for _, projectID := range projectIDs {
		masterCommitSet := make(map[string]struct{}, perPage*2)
		err := h.collectMainBranchCommits(projectID, perPage, startTimestamp, masterCommitSet)
		if err != nil {
			return combineList, err
		}
		err = h.checkMerges(projectID, authorMatch(h.Author), perPage, startTimestamp, endTimestamp, &combineList, combineDedup, masterCommitSet)
		if err != nil {
			return combineList, err
		}
	}

	sort.SliceStable(combineList, func(i, j int) bool {
		iw := statusWeight(combineList[i].Status)
		jw := statusWeight(combineList[j].Status)
		if iw == jw {
			return combineList[i].Message < combineList[j].Message
		}
		return iw < jw
	})
	return combineList, nil
}

func (h *TGitlab) checkMerges(projectID, author string, perPage int, startTimestamp, endTimestamp int64, combineList *[]Combine, combineDedup map[string]struct{}, masterCommitSet map[string]struct{}) error {
	for page := 1; page < 100; page++ {
		gitLabParam := gsapi.GsGitLabParam{State: "all", Sort: "desc", Page: page, PerPage: perPage}
		mergeList, resErr := h.GitLab.GetMerges(projectID, gitLabParam)
		if resErr != nil {
			return resErr
		}
		for _, merge := range mergeList {
			sourceBranch := cast.ToString(merge["source_branch"])
			targetBranch := cast.ToString(merge["target_branch"])
			title := cast.ToString(merge["title"])

			createdAtUnix, createdAtOk := h.getUnixTime(merge["created_at"])
			updatedAtUnix, updatedAtOk := h.getUnixTime(merge["updated_at"])
			mergedAtUnix, mergedAtOk := h.getUnixTime(merge["merged_at"])

			userCreated := strings.Contains(strings.ToLower(h.getMergeAuthor(merge)), author)
			createdToday := userCreated && inRange(createdAtOk, createdAtUnix, startTimestamp, endTimestamp)
			updatedToday := inRange(updatedAtOk, updatedAtUnix, startTimestamp, endTimestamp)
			mergedToday := inRange(mergedAtOk, mergedAtUnix, startTimestamp, endTimestamp)
			relevantByTime := createdToday || updatedToday || mergedToday

			opSummary, err := h.checkMergeUserOp(projectID, sourceBranch, author, startTimestamp, endTimestamp, masterCommitSet)
			if err != nil {
				return err
			}
			if !relevantByTime && !opSummary.branchActiveToday() {
				continue
			}
			if !opSummary.authorJoin && !userCreated {
				continue
			}

			mergedToMainToday := mergedToday && h.isMainBranch(targetBranch)
			status := h.getStatus(mergedToMainToday, opSummary.authorCommitToday || createdToday || (userCreated && updatedToday), opSummary.otherCommitToday)
			if status == "" {
				continue
			}
			h.addCombine(combineList, combineDedup, Combine{Message: title, Status: status})
		}
		if len(mergeList) < perPage {
			break
		}
	}
	return nil
}

func (h *TGitlab) getStatus(mergedToMain, hasCommitToday, hasOtherCommitToday bool) string {
	if mergedToMain {
		return "已上线"
	}
	if hasOtherCommitToday {
		return "对接中"
	}
	if hasCommitToday {
		return "开发中"
	}
	return ""
}

func (h *TGitlab) isMainBranch(branch string) bool {
	branchLower := strings.ToLower(branch)
	for _, mainBranch := range MergeMainBranchs {
		if branchLower == mainBranch || strings.Contains(branchLower, mainBranch) {
			return true
		}
	}
	return false
}

func (h *TGitlab) collectMainBranchCommits(projectID string, perPage int, startTimestamp int64, masterCommitSet map[string]struct{}) error {
	for _, mainBranch := range MergeMainBranchs {
		for page := 1; page < 100; page++ {
			gitLabParam := gsapi.GsGitLabParam{Sort: "desc", Page: page, PerPage: perPage, RefName: mainBranch}
			commitList, resErr := h.GitLab.GetProjectCommits(projectID, gitLabParam)
			if resErr != nil {
				return resErr
			}
			boolBroken := false
			for _, commit := range commitList {
				id := cast.ToString(commit["id"])
				masterCommitSet[id] = struct{}{}
				createdAt := cast.ToString(commit["created_at"])
				beijingTime, beijingTimeErr := h.gBeijingTime(createdAt)
				if beijingTimeErr != nil {
					return errors.New("解析时间报错 " + beijingTimeErr.Error())
				}
				if beijingTime.Unix() < startTimestamp {
					boolBroken = true
					break
				}
			}
			if boolBroken || len(commitList) < perPage {
				break
			}
		}
	}
	return nil
}

func (h *TGitlab) checkMergeUserOp(projectID, branchName, author string, startTimestamp, endTimestamp int64, masterCommitSet map[string]struct{}) (mergeUserOpSummary, error) {
	summary := mergeUserOpSummary{}
	if branchName == "" {
		return summary, nil
	}
	for page := 1; page < 100; page++ {
		gitLabParam := gsapi.GsGitLabParam{Sort: "desc", Page: page, PerPage: 50, RefName: branchName}
		commitList, resErr := h.GitLab.GetProjectCommits(projectID, gitLabParam)
		if resErr != nil {
			return summary, resErr
		}
		boolBroken := false
		for _, commit := range commitList {
			id := cast.ToString(commit["id"])
			if _, exist := masterCommitSet[id]; exist {
				continue
			}

			authorName := strings.ToLower(cast.ToString(commit["author_name"]))
			committerName := strings.ToLower(cast.ToString(commit["committer_name"]))
			createdAt := cast.ToString(commit["created_at"])
			message := cast.ToString(commit["message"])

			beijingTime, beijingTimeErr := h.gBeijingTime(createdAt)
			if beijingTimeErr != nil {
				return summary, beijingTimeErr
			}
			isAuthorCommit := strings.Contains(authorName, author) || strings.Contains(committerName, author)
			isMergeLikeCommit := h.isMergeLikeCommit(message)
			commitUnix := beijingTime.Unix()

			if commitUnix > endTimestamp {
				if isAuthorCommit {
					summary.authorJoin = true
				}
				continue
			}
			if commitUnix < startTimestamp {
				if isAuthorCommit {
					summary.authorJoin = true
				}
				if summary.authorJoin {
					boolBroken = true
					break
				}
				continue
			}

			if isAuthorCommit {
				summary.authorJoin = true
				if isMergeLikeCommit {
					summary.authorMergeLikeCommitToday = true
				} else {
					summary.authorCommitToday = true
				}
			} else if !isMergeLikeCommit {
				summary.otherCommitToday = true
			}
		}
		if boolBroken || len(commitList) < 50 {
			break
		}
	}
	return summary, nil
}

func (h *TGitlab) isMergeLikeCommit(message string) bool {
	message = strings.ToLower(strings.TrimSpace(message))
	if message == "" {
		return false
	}

	mergeLikeKeywords := []string{
		"merge branch",
		"merge remote-tracking branch",
		"merge pull request",
		"merged in ",
		"see merge request",
	}
	for _, keyword := range mergeLikeKeywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return strings.HasPrefix(message, "merge ")
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
	key := combine.Status + "|" + combine.Message
	if _, exist := combineDedup[key]; exist {
		return
	}
	combineDedup[key] = struct{}{}
	*combineList = append(*combineList, combine)
	if h.LogFunc != nil {
		h.LogFunc(gstool.JsonEncode(combine))
	}
}

func (h *TGitlab) getUnixTime(v any) (int64, bool) {
	s := cast.ToString(v)
	if s == "" || s == "<nil>" {
		return 0, false
	}
	t, err := h.gBeijingTime(s)
	if err != nil {
		return 0, false
	}
	return t.Unix(), true
}

func (h *TGitlab) getMergeAuthor(merge map[string]any) string {
	if v, ok := merge["author_name"]; ok {
		if s := cast.ToString(v); s != "" && s != "<nil>" {
			return s
		}
	}
	if v, ok := merge["author"]; ok {
		if m, ok := v.(map[string]any); ok {
			if s := cast.ToString(m["name"]); s != "" && s != "<nil>" {
				return s
			}
			if s := cast.ToString(m["username"]); s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func statusWeight(status string) int {
	switch status {
	case "已上线":
		return 1
	case "对接中":
		return 2
	case "开发中":
		return 3
	default:
		return 4
	}
}

func authorMatch(author string) string {
	return strings.ToLower(author)
}

func inRange(ok bool, unixVal, startTimestamp, endTimestamp int64) bool {
	return ok && unixVal >= startTimestamp && unixVal <= endTimestamp
}
