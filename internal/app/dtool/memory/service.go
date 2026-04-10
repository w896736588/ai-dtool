package memory

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	yaml "gopkg.in/yaml.v3"
)

var errFragmentNotFound = errors.New(`片段不存在`)

// Service 管理文件型知识片段及其内存索引。
type Service struct {
	root string

	mu     sync.RWMutex
	byID   map[string]Fragment
	byPath map[string]string
	ready  bool

	watcherMu sync.Mutex
	watcher   *fsnotify.Watcher
}

// NewService 创建知识片段服务。
func NewService(root string) *Service {
	return &Service{
		root:   root,
		byID:   make(map[string]Fragment),
		byPath: make(map[string]string),
	}
}

// Load 扫描目录并重建索引。
func (h *Service) Load() error {
	h.mu.Lock()
	h.byID = make(map[string]Fragment)
	h.byPath = make(map[string]string)
	h.ready = false
	h.mu.Unlock()

	for _, item := range []struct {
		dir       string
		isDeleted bool
	}{
		{dir: filepath.Join(h.root, `fragments`), isDeleted: false},
		{dir: filepath.Join(h.root, `trash`), isDeleted: true},
	} {
		if err := h.scanDir(item.dir, item.isDeleted); err != nil {
			return err
		}
	}

	h.mu.Lock()
	h.ready = true
	h.mu.Unlock()
	return nil
}

// LoadAsync 异步重建索引。
func (h *Service) LoadAsync() {
	go func() {
		_ = h.Load()
	}()
}

// IndexReady 返回索引是否已建立完成。
func (h *Service) IndexReady() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ready
}

// FragmentCount 返回正常片段数量。
func (h *Service) FragmentCount() int {
	return len(h.listByDeleted(false, 0))
}

// TrashCount 返回回收站片段数量。
func (h *Service) TrashCount() int {
	return len(h.listByDeleted(true, 0))
}

// StartWatching 启动目录监听。
func (h *Service) StartWatching() error {
	h.watcherMu.Lock()
	defer h.watcherMu.Unlock()
	if h.watcher != nil {
		return nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	for _, dir := range []string{h.root, filepath.Join(h.root, `fragments`), filepath.Join(h.root, `trash`)} {
		if err = h.addWatchRecursive(watcher, dir); err != nil {
			_ = watcher.Close()
			return err
		}
	}
	h.watcher = watcher
	go h.watchLoop(watcher)
	return nil
}

// StopWatching 停止目录监听。
func (h *Service) StopWatching() error {
	h.watcherMu.Lock()
	defer h.watcherMu.Unlock()
	if h.watcher == nil {
		return nil
	}
	err := h.watcher.Close()
	h.watcher = nil
	return err
}

// MemoryFragmentList 查询正常片段列表。
func (h *Service) MemoryFragmentList(limit int) ([]map[string]any, error) {
	return h.listByDeleted(false, limit), nil
}

// MemoryFragmentTrashList 查询回收站列表。
func (h *Service) MemoryFragmentTrashList(limit int) ([]map[string]any, error) {
	return h.listByDeleted(true, limit), nil
}

// MemoryFragmentInfo 查询单个片段详情。
func (h *Service) MemoryFragmentInfo(id any) (map[string]any, error) {
	fragment, ok := h.getFragment(normalizeID(id))
	if !ok || fragment.IsDeleted {
		return nil, errFragmentNotFound
	}
	loaded, err := ParseFragmentFile(fragment.FilePath, false, true)
	if err != nil {
		return nil, err
	}
	loaded.FilePath = fragment.FilePath
	return fragmentToMap(*loaded), nil
}

// MemoryFragmentSave 保存片段。
func (h *Service) MemoryFragmentSave(id any, title, content string, _ []string) (map[string]any, error) {
	idText := normalizeID(id)
	now := time.Now()
	if idText == `` || idText == `0` {
		idText = h.generateFragmentID()
	}

	h.mu.RLock()
	oldFragment, exists := h.byID[idText]
	h.mu.RUnlock()

	fragment := Fragment{
		ID:        idText,
		Title:     title,
		Content:   normalizeLineBreaks(content),
		CreatedAt: now,
		UpdatedAt: now,
		IsDeleted: false,
	}
	if exists {
		fragment.CreatedAt = oldFragment.CreatedAt
		fragment.IsDeleted = oldFragment.IsDeleted
		fragment.FilePath = oldFragment.FilePath
	}
	if fragment.CreatedAt.IsZero() {
		fragment.CreatedAt = now
	}
	if fragment.IsDeleted {
		fragment.IsDeleted = false
	}
	filePath := BuildFragmentPath(h.root, fragment.CreatedAt, fragment.ID, false)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return nil, err
	}
	fragment.FilePath = filePath
	rendered, err := RenderFragmentMarkdown(fragment)
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(filePath, []byte(rendered), 0o644); err != nil {
		return nil, err
	}
	if exists && oldFragment.FilePath != `` && oldFragment.FilePath != filePath {
		_ = os.Remove(oldFragment.FilePath)
	}
	parsed, err := ParseFragmentFile(filePath, false, true)
	if err != nil {
		return nil, err
	}
	h.upsert(fragmentMetadata(*parsed))
	return fragmentToMap(*parsed), nil
}

// MemoryFragmentSoftDelete 软删除片段。
func (h *Service) MemoryFragmentSoftDelete(id any) (int64, error) {
	fragment, ok := h.getFragment(normalizeID(id))
	if !ok || fragment.IsDeleted {
		return 0, errFragmentNotFound
	}
	targetPath := BuildFragmentPath(h.root, fragment.CreatedAt, fragment.ID, true)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return 0, err
	}
	if err := os.Rename(fragment.FilePath, targetPath); err != nil {
		return 0, err
	}
	fragment.FilePath = targetPath
	fragment.IsDeleted = true
	fragment.UpdatedAt = time.Now()
	h.upsert(fragment)
	return 1, nil
}

// MemoryFragmentRestore 从回收站恢复片段。
func (h *Service) MemoryFragmentRestore(id any) (int64, error) {
	fragment, ok := h.getFragment(normalizeID(id))
	if !ok || !fragment.IsDeleted {
		return 0, errFragmentNotFound
	}
	targetPath := BuildFragmentPath(h.root, fragment.CreatedAt, fragment.ID, false)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return 0, err
	}
	if err := os.Rename(fragment.FilePath, targetPath); err != nil {
		return 0, err
	}
	fragment.FilePath = targetPath
	fragment.IsDeleted = false
	fragment.UpdatedAt = time.Now()
	h.upsert(fragment)
	return 1, nil
}

// MemoryFragmentHardDelete 彻底删除回收站中的片段。
func (h *Service) MemoryFragmentHardDelete(id any) error {
	fragment, ok := h.getFragment(normalizeID(id))
	if !ok || !fragment.IsDeleted {
		return errFragmentNotFound
	}
	if err := os.Remove(fragment.FilePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	h.remove(fragment.ID, fragment.FilePath)
	return nil
}

// MemoryFragmentHistoryList 首版不再维护独立历史，统一交给 Git。
func (h *Service) MemoryFragmentHistoryList(id any) ([]map[string]any, error) {
	if _, ok := h.getFragment(normalizeID(id)); !ok {
		return nil, errFragmentNotFound
	}
	return []map[string]any{}, nil
}

// MemoryFragmentTagList 已移除标签功能。
func (h *Service) MemoryFragmentTagList() ([]map[string]any, error) {
	return []map[string]any{}, nil
}

// MemoryFragmentSearch 关键词搜索知识片段。
func (h *Service) MemoryFragmentSearch(_ string, query string, _ []string, limit int) ([]map[string]any, error) {
	query = normalizeSearchQuery(query)
	if query == `` {
		return h.listByDeleted(false, limit), nil
	}
	tokenList := strings.Fields(query)
	if len(tokenList) == 0 {
		return h.listByDeleted(false, limit), nil
	}
	if _, err := exec.LookPath(`rg`); err != nil {
		return h.searchByTitleOnly(tokenList, limit), nil
	}
	return h.searchWithRipgrep(tokenList, limit)
}

func (h *Service) listByDeleted(isDeleted bool, limit int) []map[string]any {
	h.mu.RLock()
	rowList := make([]Fragment, 0, len(h.byID))
	for _, fragment := range h.byID {
		if fragment.IsDeleted != isDeleted {
			continue
		}
		rowList = append(rowList, fragment)
	}
	h.mu.RUnlock()
	sort.SliceStable(rowList, func(i, j int) bool {
		if rowList[i].UpdatedAt.Equal(rowList[j].UpdatedAt) {
			return rowList[i].ID > rowList[j].ID
		}
		return rowList[i].UpdatedAt.After(rowList[j].UpdatedAt)
	})
	if limit > 0 && len(rowList) > limit {
		rowList = rowList[:limit]
	}
	result := make([]map[string]any, 0, len(rowList))
	for _, fragment := range rowList {
		result = append(result, fragmentToMap(fragment))
	}
	return result
}

func (h *Service) scanDir(root string, isDeleted bool) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != `.md` {
			return nil
		}
		fragment, parseErr := ParseFragmentFile(path, isDeleted, false)
		if parseErr != nil {
			return parseErr
		}
		h.upsert(*fragment)
		return nil
	})
}

func (h *Service) upsert(fragment Fragment) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if oldID, ok := h.byPath[fragment.FilePath]; ok && oldID != fragment.ID {
		delete(h.byID, oldID)
	}
	h.byID[fragment.ID] = fragment
	h.byPath[fragment.FilePath] = fragment.ID
}

func (h *Service) remove(id, path string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.byID, id)
	delete(h.byPath, path)
}

func (h *Service) removeByPath(path string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if id, ok := h.byPath[path]; ok {
		delete(h.byID, id)
		delete(h.byPath, path)
	}
}

func (h *Service) getFragment(id string) (Fragment, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	fragment, ok := h.byID[id]
	return fragment, ok
}

func (h *Service) generateFragmentID() string {
	for {
		id := strconv.FormatInt(time.Now().UnixNano(), 10)
		if _, exists := h.getFragment(id); exists {
			time.Sleep(time.Nanosecond)
			continue
		}
		return id
	}
}

// ParseFragmentFile 解析单个 Markdown 片段文件。
func ParseFragmentFile(path string, isDeleted bool, loadContent bool) (*Fragment, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := normalizeLineBreaks(string(body))
	meta, markdownBody, err := parseFrontMatter(content)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	fragmentID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	createdAt := parseTimeOrZero(meta.CreatedAt)
	updatedAt := parseTimeOrZero(meta.UpdatedAt)
	if createdAt.IsZero() {
		createdAt = info.ModTime()
	}
	if updatedAt.IsZero() {
		updatedAt = info.ModTime()
	}
	fragment := &Fragment{
		ID:        fragmentID,
		Title:     NormalizeFragmentTitle(meta.Title, markdownBody),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		IsDeleted: isDeleted,
		FilePath:  path,
	}
	if loadContent {
		fragment.Content = strings.TrimSpace(markdownBody)
	}
	return fragment, nil
}

func parseFrontMatter(content string) (FrontMatter, string, error) {
	content = normalizeLineBreaks(content)
	if !strings.HasPrefix(content, "---\n") {
		return FrontMatter{}, strings.TrimSpace(content), nil
	}
	rest := strings.TrimPrefix(content, "---\n")
	endIndex := strings.Index(rest, "\n---\n")
	if endIndex < 0 {
		return FrontMatter{}, strings.TrimSpace(content), nil
	}
	metaText := rest[:endIndex]
	body := strings.TrimSpace(rest[endIndex+5:])
	meta := FrontMatter{}
	if err := yaml.Unmarshal([]byte(metaText), &meta); err != nil {
		return FrontMatter{}, ``, err
	}
	return meta, body, nil
}

func fragmentToMap(fragment Fragment) map[string]any {
	result := map[string]any{
		`id`:                fragmentIDForResponse(fragment.ID),
		`file_id`:           fragment.ID,
		`file_path`:         fragment.FilePath,
		`title`:             fragment.Title,
		`content`:           fragment.Content,
		`tags`:              []string{},
		`create_time`:       fragment.CreatedAt.Unix(),
		`update_time`:       fragment.UpdatedAt.Unix(),
		`create_time_desc`:  fragment.CreatedAt.Format(`2006-01-02 15:04:05`),
		`update_time_desc`:  fragment.UpdatedAt.Format(`2006-01-02 15:04:05`),
		`is_deleted`:        boolToInt(fragment.IsDeleted),
	}
	return result
}

func fragmentIDForResponse(id string) any {
	if value, err := strconv.ParseInt(id, 10, 64); err == nil {
		return value
	}
	return id
}

func normalizeID(id any) string {
	text := strings.TrimSpace(fmt.Sprintf("%v", id))
	if text == `<nil>` {
		return ``
	}
	return text
}

func fragmentMetadata(fragment Fragment) Fragment {
	fragment.Content = ``
	return fragment
}

func normalizeSearchQuery(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(query)), ` `)
}

func markdownToPlainText(content string) string {
	text := normalizeLineBreaks(content)
	replacer := strings.NewReplacer(
		"```", " ",
		"`", " ",
		"#", " ",
		">", " ",
		"*", " ",
		"_", " ",
		"-", " ",
		"|", " ",
		"[", " ",
		"]", " ",
		"(", " ",
		")", " ",
	)
	text = replacer.Replace(text)
	linkReg := regexp.MustCompile(`https?://[^\s]+`)
	text = linkReg.ReplaceAllString(text, ` `)
	spaceReg := regexp.MustCompile(`\s+`)
	text = spaceReg.ReplaceAllString(text, ` `)
	return strings.TrimSpace(text)
}

func parseTimeOrZero(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == `` {
		return time.Time{}
	}
	if value, err := time.Parse(timeLayout, raw); err == nil {
		return value
	}
	return time.Time{}
}

func (h *Service) addWatchRecursive(watcher *fsnotify.Watcher, root string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		return watcher.Add(path)
	})
}

func (h *Service) watchLoop(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			h.handleWatchEvent(watcher, event)
		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (h *Service) handleWatchEvent(watcher *fsnotify.Watcher, event fsnotify.Event) {
	info, err := os.Stat(event.Name)
	if err == nil && info.IsDir() && event.Has(fsnotify.Create) {
		_ = h.addWatchRecursive(watcher, event.Name)
		return
	}
	if strings.ToLower(filepath.Ext(event.Name)) != `.md` {
		return
	}
	if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		h.removeByPath(event.Name)
	}
	if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) || event.Has(fsnotify.Chmod) {
		isDeleted := strings.Contains(filepath.ToSlash(event.Name), `/trash/`)
		fragment, parseErr := ParseFragmentFile(event.Name, isDeleted, false)
		if parseErr != nil {
			return
		}
		h.upsert(*fragment)
	}
}

func (h *Service) searchByTitleOnly(tokenList []string, limit int) []map[string]any {
	h.mu.RLock()
	rowList := make([]Fragment, 0, len(h.byID))
	for _, fragment := range h.byID {
		if fragment.IsDeleted {
			continue
		}
		rowList = append(rowList, fragment)
	}
	h.mu.RUnlock()
	sort.SliceStable(rowList, func(i, j int) bool {
		return rowList[i].UpdatedAt.After(rowList[j].UpdatedAt)
	})
	result := make([]map[string]any, 0, len(rowList))
	for _, fragment := range rowList {
		titleText := strings.ToLower(fragment.Title)
		matched := true
		for _, token := range tokenList {
			if !strings.Contains(titleText, token) {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		result = append(result, fragmentToMap(fragment))
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

func (h *Service) searchWithRipgrep(tokenList []string, limit int) ([]map[string]any, error) {
	searchRoot := filepath.Join(h.root, `fragments`)
	matchMap := make(map[string][]string)
	firstToken := true
	for _, token := range tokenList {
		tokenMatches, err := h.runRipgrepToken(searchRoot, token)
		if err != nil {
			return nil, err
		}
		if firstToken {
			for path, snippets := range tokenMatches {
				matchMap[path] = append([]string{}, snippets...)
			}
			firstToken = false
			continue
		}
		for path, existing := range matchMap {
			newSnippets, ok := tokenMatches[path]
			if !ok {
				delete(matchMap, path)
				continue
			}
			matchMap[path] = append(existing, newSnippets...)
		}
	}

	type searchRow struct {
		fragment Fragment
		score    int
		snippets []string
	}
	rowList := make([]searchRow, 0, len(matchMap))
	for path, snippets := range matchMap {
		fragment, ok := h.getFragmentByPath(path)
		if !ok || fragment.IsDeleted {
			continue
		}
		score := len(uniqueStrings(snippets))
		titleText := strings.ToLower(fragment.Title)
		for _, token := range tokenList {
			if strings.Contains(titleText, token) {
				score += 3
			}
		}
		rowList = append(rowList, searchRow{
			fragment: fragment,
			score:    score,
			snippets: uniqueStrings(snippets),
		})
	}
	sort.SliceStable(rowList, func(i, j int) bool {
		if rowList[i].score != rowList[j].score {
			return rowList[i].score > rowList[j].score
		}
		return rowList[i].fragment.UpdatedAt.After(rowList[j].fragment.UpdatedAt)
	})
	if limit <= 0 {
		limit = 50
	}
	result := make([]map[string]any, 0, min(limit, len(rowList)))
	for idx, item := range rowList {
		if idx >= limit {
			break
		}
		row := fragmentToMap(item.fragment)
		row[`search_snippets`] = item.snippets
		result = append(result, row)
	}
	return result, nil
}

func (h *Service) runRipgrepToken(searchRoot, token string) (map[string][]string, error) {
	cmd := exec.Command(`rg`, `--line-number`, `--with-filename`, `--color`, `never`, `--glob`, `*.md`, `--ignore-case`, `--`, token, searchRoot)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return map[string][]string{}, nil
		}
		return nil, fmt.Errorf(`rg 搜索失败: %w`, err)
	}
	result := make(map[string][]string)
	for _, line := range strings.Split(normalizeLineBreaks(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line == `` {
			continue
		}
		path, snippet, ok := parseRipgrepLine(line)
		if !ok {
			continue
		}
		if snippet == `` {
			continue
		}
		result[path] = append(result[path], snippet)
	}
	return result, nil
}

func parseRipgrepLine(line string) (string, string, bool) {
	regList := []*regexp.Regexp{
		regexp.MustCompile(`^([A-Za-z]:.+?):\d+:(.*)$`),
		regexp.MustCompile(`^(.+?):\d+:(.*)$`),
	}
	for _, reg := range regList {
		matches := reg.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}
		return filepath.Clean(matches[1]), strings.TrimSpace(matches[2]), true
	}
	return ``, ``, false
}

func (h *Service) getFragmentByPath(path string) (Fragment, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	id, ok := h.byPath[filepath.Clean(path)]
	if !ok {
		return Fragment{}, false
	}
	fragment, exists := h.byID[id]
	return fragment, exists
}

func uniqueStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == `` {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}
