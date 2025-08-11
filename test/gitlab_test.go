package test

import (
	"errors"
	"gitee.com/Sxiaobai/gs/gsapi"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
	"regexp"
	"strings"
	"testing"
	"time"
)

type Combine struct {
	Message string
	Status  string
}

func TestGitLab(t *testing.T) {
	author := `frog`
	projectMap := make(map[string]string)
	gitLab := gsapi.GsGitLab{
		BaseUrl:     "https://gitlab.zmwk.cn/api/v4",
		AccessToken: "Q75FMHQxrQhiPS4VPyZU",
	}
	//now := time.Now()
	//startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	//endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Nanosecond)
	perPage := 20
	startOfDay, _ := gstool.TimeStringToUnix(`2025-03-24 00:00:00`, `Y-m-d H:i:s`)
	endOfDay, _ := gstool.TimeStringToUnix(`2025-03-24 23:59:59`, `Y-m-d H:i:s`)
	startTimestamp := startOfDay.Unix()
	endTimestamp := endOfDay.Unix()
	gstool.FmtPrintlnLogTime(`开始时间 %s`, gstool.TimeUnixToString(startOfDay, `Y-m-d H:i:s`))
	gstool.FmtPrintlnLogTime(`结束时间 %s`, gstool.TimeUnixToString(endOfDay, `Y-m-d H:i:s`))
	//所有有权限项目
	for page := 1; page < 10; page++ {
		projectParam := gsapi.GsGitLabParam{
			State:   "",
			Sort:    "",
			Page:    page,
			PerPage: perPage,
			RefName: "",
		}
		projectList, resErr := gitLab.GetProjects(projectParam)
		if resErr != nil {
			panic(`查询错误 ` + resErr.Error())
		}
		for _, project := range projectList {
			projectMap[cast.ToString(project[`id`])] = cast.ToString(project[`name`])
		}
		if len(projectList) < perPage {
			break
		}
	}
	combineList := make([]Combine, 0)
	//所有今日已合并上线的分支名
	masterCommits := make([]string, 0)
	for projectId, projectName := range projectMap {
		//TODO 临时
		if projectId != `20` {
			continue
		}
		//检查所有已合并的提交
		checkCommits(&gitLab, projectId, projectName, author, perPage, startTimestamp, endTimestamp, &combineList, &masterCommits)
		//所有合并请求
		checkMerges(&gitLab, projectId, projectName, author, perPage, startTimestamp, endTimestamp, &combineList, &masterCommits)
	}
	gstool.FmtPrintlnLogTime(`%s`, gstool.JsonEncode(combineList))
}

func checkMerges(gitLab *gsapi.GsGitLab, projectId, projectName, author string,
	perPage int, startTimestamp, endTimestamp int64, combineList *[]Combine, masterCommits *[]string) {
	gstool.FmtPrintlnLogTime(`开始检查已合并`)
	for page := 1; page < 100; page++ {
		gitLabParam := gsapi.GsGitLabParam{
			State:   "opened",
			Sort:    "desc",
			Page:    page,
			PerPage: perPage,
			RefName: "",
		}
		mergeList, resErr := gitLab.GetMerges(projectId, gitLabParam)
		if resErr != nil {
			panic(`查询错误 ` + resErr.Error())
		}
		for _, merge := range mergeList {
			sourceBranch := cast.ToString(merge[`source_branch`])
			title := cast.ToString(merge[`title`])
			authorJoin, otherJoin, selfTest := checkMergeUserOp(gitLab, projectId, sourceBranch, author, startTimestamp, endTimestamp, masterCommits)
			status := ``
			if authorJoin {
				if otherJoin {
					if selfTest {
						status = `对接自测`
					} else {
						status = `对接`
					}
				} else {
					status = `开发`
				}
			} else {

			}
			if status != `` {
				*combineList = append(*combineList, Combine{
					Message: title,
					Status:  status,
				})
			}
		}
		if len(mergeList) < perPage {
			break
		}
	}
}

func checkCommits(gitLab *gsapi.GsGitLab, projectId, projectName, author string,
	perPage int, startTimestamp, endTimestamp int64, combineList *[]Combine, masterCommits *[]string) {
	gstool.FmtPrintlnLogTime(`开始检查commits`)
	//拿到所有的已合并到master的提交
	for page := 1; page < 100; page++ {
		gitLabParam := gsapi.GsGitLabParam{
			State:   "",
			Sort:    "desc",
			Page:    page,
			PerPage: perPage,
			RefName: "",
		}
		commitList, resErr := gitLab.GetProjectCommits(projectId, gitLabParam)
		if resErr != nil {
			panic(`查询错误 ` + resErr.Error())
		}
		boolBroken := false
		for _, commit := range commitList {
			id := cast.ToString(commit[`id`])
			*masterCommits = append(*masterCommits, id)
			createdAt := cast.ToString(commit[`created_at`])
			beijingTime, beijingTimeErr := gBeijingTime(createdAt)
			if beijingTimeErr != nil {
				panic(`解析时间报错 ` + beijingTimeErr.Error())
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
					*combineList = append(*combineList, Combine{
						Message: message,
						Status:  `已上线`,
					}) //收集合并
				} else {
					authorJoin, _, _ := checkMergeUserOp(gitLab, projectId, getBranchName(title), author, startTimestamp, endTimestamp, masterCommits)
					if authorJoin {
						*combineList = append(*combineList, Combine{
							Message: message,
							Status:  `已上线`,
						})
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
}

func getBranchName(title string) string {
	re := regexp.MustCompile(`Merge branch '([^']+)' into`)
	matches := re.FindStringSubmatch(title)
	if len(matches) > 1 {
		return matches[1]
	} else {
		return ``
	}
}

// 检查某个分支 在某个范围内是否有某个用户的提交
func checkMergeUserOp(gitLab *gsapi.GsGitLab, projectId, branchName, author string, startTimestamp, endTimestamp int64, masterCommits *[]string) (bool, bool, bool) {
	authorJoin := false //author 是否参与了
	otherJoin := false  //其他人是否参与了
	selfTest := false   //是否自测了
	gitLabParam := gsapi.GsGitLabParam{
		State:   "",
		Sort:    "desc",
		Page:    1,
		PerPage: 20,
		RefName: branchName,
	}
	commitList, resErr := gitLab.GetProjectCommits(projectId, gitLabParam)
	if resErr != nil {
		panic(`查询错误 ` + resErr.Error())
	}
	for _, commit := range commitList {
		id := cast.ToString(commit[`id`])
		if gstool.ArrayFindValueIndex(masterCommits, id) >= 0 {
			continue
		}
		authorName := cast.ToString(commit[`author_name`])
		createdAt := cast.ToString(commit[`created_at`])
		message := cast.ToString(commit[`message`])
		beijingTime, beijingTimeErr := gBeijingTime(createdAt)
		if beijingTimeErr != nil {
			panic(`解析时间报错 ` + beijingTimeErr.Error())
		}
		if beijingTime.Unix() < startTimestamp { //小于最小时间 那就直接退出
			break
		}
		if beijingTime.Unix() > endTimestamp { //大于结束时间 继续循环
			continue
		}
		if strings.Contains(message, `自测`) || strings.Contains(message, `测完`) {
			selfTest = true
		}
		if strings.Contains(authorName, author) {
			authorJoin = true
		} else {
			otherJoin = true
		}
	}
	return authorJoin, otherJoin, selfTest
}

func gBeijingTime(utcTimeStr string) (time.Time, error) {
	utcTime, err := time.Parse(time.RFC3339, utcTimeStr)
	if err != nil {
		return time.Now(), errors.New(err.Error())
	}

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Now(), errors.New(err.Error())
	}
	return utcTime.In(loc), nil
}
