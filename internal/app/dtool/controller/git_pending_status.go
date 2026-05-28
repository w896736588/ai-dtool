package controller

import (
	"bytes"
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_define"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type gitPendingStatusChecker interface {
	IsGitRepo(dir string) (bool, error)
	RootDir(dir string) (string, error)
	ListChangedFiles(dir, fileName string) ([]string, error)
	AddFile(dir, fileName string) error
	Commit(dir, fileName, message string) error
	Push(dir string) error
}

func buildGitPendingStatusPayload() map[string]any {
	gitSyncer := business.NewMemoryGit()
	mainConfig := business.ReadMainDBConfig()
	memoryConfig := business.ReadMemoryConfigFromINI()

	type repoAggregate struct {
		labelSet map[string]bool
		dirSet   map[string]bool
		files    []string
		fileSet  map[string]bool
		rootDir  string
	}
	repoMap := make(map[string]*repoAggregate)

	appendRepo := func(label string, dir string, target string) {
		if strings.TrimSpace(dir) == `` {
			return
		}
		isGitRepo, err := gitSyncer.IsGitRepo(dir)
		if err != nil || !isGitRepo {
			return
		}
		rootDir, err := gitSyncer.RootDir(dir)
		if err != nil {
			return
		}
		files, err := gitSyncer.ListChangedFiles(rootDir, `.`)
		if err != nil {
			return
		}
		rootDir = filepath.Clean(strings.TrimSpace(rootDir))
		repo := repoMap[rootDir]
		if repo == nil {
			repo = &repoAggregate{
				labelSet: make(map[string]bool),
				dirSet:   make(map[string]bool),
				fileSet:  make(map[string]bool),
				rootDir:  rootDir,
			}
			repoMap[rootDir] = repo
		}
		repo.labelSet[label] = true
		repo.dirSet[filepath.Clean(strings.TrimSpace(dir))] = true
		for _, file := range files {
			file = strings.TrimSpace(file)
			if file == `` || repo.fileSet[file] {
				continue
			}
			repo.fileSet[file] = true
			repo.files = append(repo.files, file)
		}
	}

	appendRepo(`main_db`, mainConfig.Dir, `.`)
	appendRepo(`memory`, memoryConfig.Dir, `.`)

	items := make([]map[string]any, 0, len(repoMap))
	totalCount := 0
	repoDirs := make([]string, 0, len(repoMap))
	for dir := range repoMap {
		repoDirs = append(repoDirs, dir)
	}
	sort.Strings(repoDirs)
	for _, dir := range repoDirs {
		repo := repoMap[dir]
		sort.Strings(repo.files)
		labelList := make([]string, 0, len(repo.labelSet))
		for label := range repo.labelSet {
			labelList = append(labelList, label)
		}
		sort.Strings(labelList)
		sourceDirs := make([]string, 0, len(repo.dirSet))
		for sourceDir := range repo.dirSet {
			sourceDirs = append(sourceDirs, sourceDir)
		}
		sort.Strings(sourceDirs)
		totalCount += len(repo.files)
		items = append(items, map[string]any{
			`label`:       strings.Join(labelList, ` + `),
			`labels`:      labelList,
			`dir`:         repo.rootDir,
			`source_dirs`: sourceDirs,
			`is_git_repo`: true,
			`count`:       len(repo.files),
			`files`:       repo.files,
		})
	}

	return map[string]any{
		`total_count`: totalCount,
		`repos`:       items,
	}
}

// GitPendingStatus 返回主库和知识片段目录的未提交文件状态。
func GitPendingStatus(c *gin.Context) {
	gsgin.GinResponseSuccess(c, ``, buildGitPendingStatusPayload())
}

// GitPendingCommitPush 对指定仓库执行 add + commit + push。
func GitPendingCommitPush(c *gin.Context) {
	reqMap := gsgin.GinGetParams(c)
	dir := strings.TrimSpace(cast.ToString(reqMap[`dir`]))
	message := strings.TrimSpace(cast.ToString(reqMap[`message`]))
	if dir == `` || message == `` {
		var req struct {
			Dir     string `json:"dir"`
			Message string `json:"message"`
		}
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 {
				_ = json.Unmarshal(bodyBytes, &req)
			}
		}
		if dir == `` {
			dir = strings.TrimSpace(req.Dir)
		}
		if message == `` {
			message = strings.TrimSpace(req.Message)
		}
	}
	if dir == `` {
		gsgin.GinResponseError(c, `仓库目录不能为空`, nil)
		return
	}
	dir = filepath.Clean(dir)
	if message == `` {
		message = fmt.Sprintf(`chore: sync pending changes %s`, time.Now().Format(`2006-01-02 15:04:05`))
	}
	gitSyncer := business.NewMemoryGit()
	isGitRepo, err := gitSyncer.IsGitRepo(dir)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if !isGitRepo {
		gsgin.GinResponseError(c, `指定目录不是 git 仓库`, nil)
		return
	}
	rootDir, err := gitSyncer.RootDir(dir)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if err = gitSyncer.AddFile(rootDir, `.`); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if err = gitSyncer.Commit(rootDir, `.`, message); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if err = gitSyncer.Push(rootDir); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`dir`:     rootDir,
		`message`: message,
	})
}

func sendGitPendingStatusSnapshot(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	err := sse.SendToChan(gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseGitPendingStatus,
		Data:            buildGitPendingStatusPayload(),
		Type:            p_define.SseContentTypeMsg,
	}))
	if err != nil {
		gstool.FmtPrintlnLogTime(`GitPendingStatus 广播错误 %s`, err.Error())
	}
}

func BindGitPendingStatusSSE(sse *gsgin.Sse, stopC chan int, interval time.Duration) {
	if sse == nil {
		return
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}
	sendGitPendingStatusSnapshot(sse)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sendGitPendingStatusSnapshot(sse)
			case <-stopC:
				return
			}
		}
	}()
}
