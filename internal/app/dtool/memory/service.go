package memory

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	yaml "gopkg.in/yaml.v3"
)

var errFragmentNotFound = errors.New(`片段不存在`)
var errInvalidFragmentID = errors.New(`片段ID不合法`)

// Service 管理文件型知识片段及其内存索引。
type Service struct {
	root string

	mu      sync.RWMutex
	byID    map[string]Fragment
	byPath  map[string]string
	folders map[string]Folder
	ready   bool

	watcherMu sync.Mutex
	watcher   *fsnotify.Watcher
}

// NewService 创建知识片段服务。
func NewService(root string) *Service {
	return &Service{
		root:    root,
		byID:    make(map[string]Fragment),
		byPath:  make(map[string]string),
		folders: make(map[string]Folder),
	}
}

// Load 扫描目录并重建索引。
func (h *Service) Load() error {
	h.mu.Lock()
	h.byID = make(map[string]Fragment)
	h.byPath = make(map[string]string)
	h.folders = make(map[string]Folder)
	h.ready = false
	h.mu.Unlock()

	folders, err := loadFolders(h.root)
	if err != nil {
		return err
	}
	if err = ensureFolderDirectories(h.root, folders); err != nil {
		return err
	}
	h.mu.Lock()
	for _, folder := range folders {
		h.folders[folder.FolderName] = folder
	}
	h.mu.Unlock()

	for _, folder := range folders {
		if folder.FolderName == TrashFolderName {
			continue
		}
		if err = h.scanDir(filepath.Join(h.root, folder.FolderName), false); err != nil {
			return err
		}
	}
	if err = h.scanDir(filepath.Join(h.root, TrashFolderName), true); err != nil {
		return err
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
	return len(h.listByDeleted(false, 0, 0, ``))
}

// TrashCount 返回回收站片段数量。
func (h *Service) TrashCount() int {
	return len(h.listByDeleted(true, 0, 0, ``))
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
	watchDirs := []string{h.root, filepath.Join(h.root, TrashFolderName)}
	for _, folder := range h.listFolders() {
		if folder.FolderName == TrashFolderName {
			continue
		}
		watchDirs = append(watchDirs, filepath.Join(h.root, folder.FolderName))
	}
	for _, dir := range watchDirs {
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
func (h *Service) MemoryFragmentList(limit, offset int, folderName string) ([]map[string]any, error) {
	return h.listByDeleted(false, limit, offset, folderName), nil
}

// MemoryFragmentTrashList 查询回收站列表。
func (h *Service) MemoryFragmentTrashList(limit int) ([]map[string]any, error) {
	return h.listByDeleted(true, limit, 0, ``), nil
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
	return h.fragmentToMap(*loaded), nil
}

// MemoryFragmentSave 保存片段。
func (h *Service) MemoryFragmentSave(id any, title, content string, _ []string, folderName string) (map[string]any, error) {
	idText := normalizeID(id)
	now := time.Now()
	if idText == `` || idText == `0` {
		idText = h.generateFragmentID()
	} else if !IsValidFragmentID(idText) {
		return nil, errInvalidFragmentID
	}
	folderName = NormalizeFolderName(folderName)

	h.mu.RLock()
	oldFragment, exists := h.byID[idText]
	h.mu.RUnlock()
	if exists && strings.TrimSpace(folderName) == `` {
		folderName = NormalizeFolderName(oldFragment.FolderName)
	}
	if err := h.ensureFolderExists(folderName); err != nil {
		return nil, err
	}

	fragment := Fragment{
		ID:         idText,
		Title:      title,
		Content:    normalizeLineBreaks(content),
		FolderName: folderName,
		CreatedAt:  now,
		UpdatedAt:  now,
		IsDeleted:  false,
	}
	if exists {
		if strings.TrimSpace(content) == `` && strings.TrimSpace(oldFragment.Content) != `` {
			return nil, errors.New(`不允许将已有内容的片段更新为空`)
		}
		if strings.TrimSpace(title) == `` {
			title = oldFragment.Title
			fragment.Title = title
		}
		fragment.CreatedAt = oldFragment.CreatedAt
		fragment.IsDeleted = oldFragment.IsDeleted
		fragment.FilePath = oldFragment.FilePath
		fragment.FolderName = folderName
	}
	if fragment.CreatedAt.IsZero() {
		fragment.CreatedAt = now
	}
	if fragment.IsDeleted {
		fragment.IsDeleted = false
	}
	filePath := BuildFragmentPath(h.root, fragment.CreatedAt, fragment.ID, false, fragment.FolderName)
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
	return h.fragmentToMap(*parsed), nil
}

// MemoryFragmentSoftDelete 软删除片段。
func (h *Service) MemoryFragmentSoftDelete(id any) (int64, error) {
	fragment, ok := h.getFragment(normalizeID(id))
	if !ok || fragment.IsDeleted {
		return 0, errFragmentNotFound
	}
	targetPath := BuildFragmentPath(h.root, fragment.CreatedAt, fragment.ID, true, fragment.FolderName)
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
	restoreFolderName := NormalizeFolderName(fragment.FolderName)
	if !h.folderExists(restoreFolderName) || restoreFolderName == TrashFolderName {
		restoreFolderName = DefaultFolderName
	}
	targetPath := BuildFragmentPath(h.root, fragment.CreatedAt, fragment.ID, false, restoreFolderName)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return 0, err
	}
	if err := os.Rename(fragment.FilePath, targetPath); err != nil {
		return 0, err
	}
	fragment.FilePath = targetPath
	fragment.FolderName = restoreFolderName
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
	fragment, ok := h.getFragment(normalizeID(id))
	if !ok {
		return nil, errFragmentNotFound
	}
	return h.gitHistoryList(fragment)
}

// MemoryFragmentTagList 已移除标签功能。
func (h *Service) MemoryFragmentTagList() ([]map[string]any, error) {
	return []map[string]any{}, nil
}

// MemoryFragmentSearch 关键词搜索知识片段。
func (h *Service) MemoryFragmentSearch(_ string, query string, _ []string, folderName string, limit int) ([]map[string]any, error) {
	query = normalizeSearchQuery(query)
	if query == `` {
		return h.listByDeleted(false, limit, 0, folderName), nil
	}
	tokenList := strings.Fields(query)
	if len(tokenList) == 0 {
		return h.listByDeleted(false, limit, 0, folderName), nil
	}
	if _, err := exec.LookPath(`rg`); err != nil {
		return h.searchByTitleOnly(tokenList, folderName, limit), nil
	}
	return h.searchWithRipgrep(tokenList, folderName, limit)
}

func (h *Service) listByDeleted(isDeleted bool, limit, offset int, folderName string) []map[string]any {
	folderName = strings.TrimSpace(folderName)
	h.mu.RLock()
	rowList := make([]Fragment, 0, len(h.byID))
	for _, fragment := range h.byID {
		if fragment.IsDeleted != isDeleted {
			continue
		}
		if folderName != `` && NormalizeFolderName(fragment.FolderName) != NormalizeFolderName(folderName) {
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
	if offset > 0 && offset < len(rowList) {
		rowList = rowList[offset:]
	} else if offset >= len(rowList) {
		rowList = nil
	}
	if limit > 0 && len(rowList) > limit {
		rowList = rowList[:limit]
	}
	result := make([]map[string]any, 0, len(rowList))
	for _, fragment := range rowList {
		result = append(result, h.fragmentToMap(fragment))
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
		id := uuid.NewString()
		if _, exists := h.getFragment(id); exists {
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
	if !IsValidFragmentID(fragmentID) {
		return nil, fmt.Errorf(`%w: %s`, errInvalidFragmentID, fragmentID)
	}
	createdAt := parseTimeOrZero(meta.CreatedAt)
	updatedAt := parseTimeOrZero(meta.UpdatedAt)
	if createdAt.IsZero() {
		createdAt = info.ModTime()
	}
	if updatedAt.IsZero() {
		updatedAt = info.ModTime()
	}
	fragment := &Fragment{
		ID:         fragmentID,
		Title:      NormalizeFragmentTitle(meta.Title, markdownBody),
		FolderName: detectFragmentFolder(path, isDeleted, meta.FolderName),
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		IsDeleted:  isDeleted,
		FilePath:   path,
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
		// 文件型片段 ID 属于不透明标识，即便全数字也必须按字符串返回，避免前端出现 JS 安全整数精度丢失。
		// File-backed fragment IDs are opaque identifiers and must stay strings to avoid JS safe-integer precision loss.
		`id`:               fragment.ID,
		`file_id`:          fragment.ID,
		`file_path`:        fragment.FilePath,
		`title`:            fragment.Title,
		`content`:          fragment.Content,
		`folder_name`:      NormalizeFolderName(fragment.FolderName),
		`tags`:             []string{},
		`create_time`:      fragment.CreatedAt.Unix(),
		`update_time`:      fragment.UpdatedAt.Unix(),
		`create_time_desc`: fragment.CreatedAt.Format(`2006-01-02 15:04:05`),
		`update_time_desc`: fragment.UpdatedAt.Format(`2006-01-02 15:04:05`),
		`is_deleted`:       boolToInt(fragment.IsDeleted),
	}
	return result
}

func (h *Service) fragmentToMap(fragment Fragment) map[string]any {
	result := fragmentToMap(fragment)
	folder := h.getFolder(fragment.FolderName)
	result[`folder_label`] = folder.Name
	result[`is_system_folder`] = boolToInt(folder.System)
	return result
}

func normalizeID(id any) string {
	text := strings.TrimSpace(fmt.Sprintf("%v", id))
	if text == `<nil>` {
		return ``
	}
	// 如果传入的是文件路径（含路径分隔符或 .md），提取文件名的 UUID 部分
	text = filepath.Base(filepath.ToSlash(text))
	text = strings.TrimSuffix(text, `.md`)
	return text
}

func IsValidFragmentID(id string) bool {
	id = strings.TrimSpace(id)
	if id == `` || id == `.` || id == `..` {
		return false
	}
	if strings.ContainsAny(id, `/\`) {
		return false
	}
	return true
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

func (h *Service) searchByTitleOnly(tokenList []string, folderName string, limit int) []map[string]any {
	h.mu.RLock()
	rowList := make([]Fragment, 0, len(h.byID))
	for _, fragment := range h.byID {
		if fragment.IsDeleted {
			continue
		}
		if strings.TrimSpace(folderName) != `` && NormalizeFolderName(fragment.FolderName) != NormalizeFolderName(folderName) {
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
		result = append(result, h.fragmentToMap(fragment))
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

func (h *Service) searchWithRipgrep(tokenList []string, folderName string, limit int) ([]map[string]any, error) {
	searchRoot := h.searchRootForFolder(folderName)
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
		if strings.TrimSpace(folderName) != `` && NormalizeFolderName(fragment.FolderName) != NormalizeFolderName(folderName) {
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
		row := h.fragmentToMap(item.fragment)
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

func (h *Service) searchRootForFolder(folderName string) string {
	folderName = strings.TrimSpace(folderName)
	if folderName == `` {
		return h.root
	}
	return filepath.Join(h.root, NormalizeFolderName(folderName))
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

func detectFragmentFolder(path string, isDeleted bool, metaFolderName string) string {
	metaFolderName = strings.TrimSpace(metaFolderName)
	if metaFolderName != `` {
		return NormalizeFolderName(metaFolderName)
	}
	normalized := filepath.ToSlash(filepath.Clean(path))
	parts := strings.Split(normalized, `/`)
	for idx, part := range parts {
		if isDeleted && part == TrashFolderName {
			if idx+1 < len(parts) {
				return NormalizeFolderName(parts[idx+1])
			}
			return DefaultFolderName
		}
	}
	if len(parts) >= 4 {
		return NormalizeFolderName(parts[len(parts)-4])
	}
	return DefaultFolderName
}

// MemoryFragmentBatchInfoByPaths 批量按文件路径查询片段摘要（id + title）。
func (h *Service) MemoryFragmentBatchInfoByPaths(paths []string) []map[string]any {
	results := make([]map[string]any, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == `` {
			continue
		}
		fragment, ok := h.getFragmentByPath(p)
		if !ok || fragment.IsDeleted {
			continue
		}
		results = append(results, map[string]any{
			`file_path`: fragment.FilePath,
			`id`:        fragment.ID,
			`title`:     fragment.Title,
		})
	}
	return results
}

func (h *Service) listFolders() []Folder {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]Folder, 0, len(h.folders))
	for _, folder := range h.folders {
		result = append(result, folder)
	}
	sortFolders(result)
	return result
}

func (h *Service) getFolder(folderName string) Folder {
	h.mu.RLock()
	defer h.mu.RUnlock()
	folderName = NormalizeFolderName(folderName)
	if folder, ok := h.folders[folderName]; ok {
		return folder
	}
	return Folder{
		FolderName: folderName,
		Name:       folderName,
		System:     false,
		Editable:   true,
	}
}

func (h *Service) folderExists(folderName string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.folders[NormalizeFolderName(folderName)]
	return ok
}

func (h *Service) ensureFolderExists(folderName string) error {
	folderName = NormalizeFolderName(folderName)
	if folderName == TrashFolderName {
		return errors.New(`不能将片段保存到回收站`)
	}
	if !h.folderExists(folderName) {
		return fmt.Errorf(`文件夹不存在: %s`, folderName)
	}
	return nil
}

func (h *Service) MemoryFragmentFolderList() ([]map[string]any, error) {
	folders := h.listFolders()
	countMap := make(map[string]int)
	h.mu.RLock()
	for _, fragment := range h.byID {
		if fragment.IsDeleted {
			continue
		}
		countMap[NormalizeFolderName(fragment.FolderName)]++
	}
	h.mu.RUnlock()
	result := make([]map[string]any, 0, len(folders))
	for _, folder := range folders {
		result = append(result, map[string]any{
			`folder_name`:    folder.FolderName,
			`name`:           folder.Name,
			`system`:         boolToInt(folder.System),
			`editable`:       boolToInt(folder.Editable),
			`fragment_count`: countMap[folder.FolderName],
		})
	}
	return result, nil
}

func (h *Service) MemoryFragmentFolderCreate(name, folderName string) (map[string]any, error) {
	name = strings.TrimSpace(name)
	if name == `` {
		return nil, errors.New(`名称不能为空`)
	}
	folderName = strings.TrimSpace(folderName)
	if err := validateNewFolderName(folderName); err != nil {
		return nil, err
	}
	folders := h.listFolders()
	for _, folder := range folders {
		if folder.FolderName == folderName {
			return nil, errors.New(`文件夹名称不能重复`)
		}
		if strings.EqualFold(strings.TrimSpace(folder.Name), name) {
			return nil, errors.New(`名称不能重复`)
		}
	}
	item := Folder{
		FolderName: folderName,
		Name:       name,
		System:     false,
		Editable:   true,
	}
	folders = append(folders, item)
	if err := saveFolders(h.root, folders); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(h.root, folderName), 0o755); err != nil {
		return nil, err
	}
	h.mu.Lock()
	h.folders[folderName] = item
	h.mu.Unlock()
	return map[string]any{
		`folder_name`:    item.FolderName,
		`name`:           item.Name,
		`system`:         0,
		`editable`:       1,
		`fragment_count`: 0,
	}, nil
}

func (h *Service) MemoryFragmentFolderUpdate(folderName, name string) (map[string]any, error) {
	folderName = NormalizeFolderName(folderName)
	name = strings.TrimSpace(name)
	if name == `` {
		return nil, errors.New(`名称不能为空`)
	}
	folders := h.listFolders()
	found := false
	var updated Folder
	for idx, folder := range folders {
		if folder.FolderName == folderName {
			if !folder.Editable {
				return nil, errors.New(`该文件夹不允许编辑`)
			}
			folder.Name = name
			folders[idx] = folder
			updated = folder
			found = true
			continue
		}
		if strings.EqualFold(strings.TrimSpace(folder.Name), name) {
			return nil, errors.New(`名称不能重复`)
		}
	}
	if !found {
		return nil, errors.New(`文件夹不存在`)
	}
	if err := saveFolders(h.root, folders); err != nil {
		return nil, err
	}
	h.mu.Lock()
	h.folders[folderName] = updated
	h.mu.Unlock()
	return map[string]any{
		`folder_name`:    updated.FolderName,
		`name`:           updated.Name,
		`system`:         boolToInt(updated.System),
		`editable`:       boolToInt(updated.Editable),
		`fragment_count`: 0,
	}, nil
}

func (h *Service) MemoryFragmentChangeFolder(id any, folderName string) (map[string]any, error) {
	folderName = NormalizeFolderName(folderName)
	if err := h.ensureFolderExists(folderName); err != nil {
		return nil, err
	}
	fragment, ok := h.getFragment(normalizeID(id))
	if !ok || fragment.IsDeleted {
		return nil, errFragmentNotFound
	}
	loaded, err := ParseFragmentFile(fragment.FilePath, false, true)
	if err != nil {
		return nil, err
	}
	if NormalizeFolderName(loaded.FolderName) == folderName {
		return h.fragmentToMap(*loaded), nil
	}
	loaded.FolderName = folderName
	loaded.UpdatedAt = time.Now()
	targetPath := BuildFragmentPath(h.root, loaded.CreatedAt, loaded.ID, false, folderName)
	if err = os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return nil, err
	}
	rendered, err := RenderFragmentMarkdown(*loaded)
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(targetPath, []byte(rendered), 0o644); err != nil {
		return nil, err
	}
	if fragment.FilePath != targetPath {
		_ = os.Remove(fragment.FilePath)
	}
	parsed, err := ParseFragmentFile(targetPath, false, true)
	if err != nil {
		return nil, err
	}
	h.upsert(fragmentMetadata(*parsed))
	return h.fragmentToMap(*parsed), nil
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

// SearchFragmentsOr 使用 OR 逻辑搜索片段（任一关键词匹配即返回），用于 AI 智能搜索。
func (h *Service) SearchFragmentsOr(keywords []string, limit int) ([]map[string]any, error) {
	searchRoot := filepath.Join(h.root, DefaultFolderName)
	// 逐关键词搜索，用并集合并结果
	matchMap := make(map[string][]string)
	for _, token := range keywords {
		tokenMatches, err := h.runRipgrepToken(searchRoot, token)
		if err != nil {
			// rg 不可用时回退到标题搜索
			return h.searchByTitleOr(keywords, limit), nil
		}
		for path, snippets := range tokenMatches {
			matchMap[path] = append(matchMap[path], snippets...)
		}
	}

	type searchRow struct {
		fragment Fragment
		score    int
	}
	rowList := make([]searchRow, 0, len(matchMap))
	for path, snippets := range matchMap {
		fragment, ok := h.getFragmentByPath(path)
		if !ok || fragment.IsDeleted {
			continue
		}
		score := len(uniqueStrings(snippets))
		titleText := strings.ToLower(fragment.Title)
		for _, token := range keywords {
			if strings.Contains(titleText, token) {
				score += 3
			}
		}
		rowList = append(rowList, searchRow{fragment: fragment, score: score})
	}
	sort.SliceStable(rowList, func(i, j int) bool {
		if rowList[i].score != rowList[j].score {
			return rowList[i].score > rowList[j].score
		}
		return rowList[i].fragment.UpdatedAt.After(rowList[j].fragment.UpdatedAt)
	})
	if limit <= 0 {
		limit = 200
	}
	result := make([]map[string]any, 0, min(limit, len(rowList)))
	for idx, item := range rowList {
		if idx >= limit {
			break
		}
		result = append(result, h.fragmentToMap(item.fragment))
	}
	return result, nil
}

// searchByTitleOr 使用 OR 逻辑按标题搜索（rg 不可用时的回退方案）。
func (h *Service) searchByTitleOr(keywords []string, limit int) []map[string]any {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]map[string]any, 0)
	for _, fragment := range h.byID {
		if fragment.IsDeleted {
			continue
		}
		titleText := strings.ToLower(fragment.Title)
		matched := false
		for _, token := range keywords {
			if strings.Contains(titleText, token) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		result = append(result, h.fragmentToMap(fragment))
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

// ReadFragmentContent 根据文件路径读取片段正文内容（不含 frontmatter）。
func (h *Service) ReadFragmentContent(filePath string) (string, error) {
	body, err := os.ReadFile(filePath)
	if err != nil {
		return ``, err
	}
	content := normalizeLineBreaks(string(body))
	_, markdownBody, err := parseFrontMatter(content)
	if err != nil {
		return ``, err
	}
	return strings.TrimSpace(markdownBody), nil
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func (h *Service) gitHistoryList(fragment Fragment) ([]map[string]any, error) {
	if strings.TrimSpace(fragment.FilePath) == `` {
		return []map[string]any{}, nil
	}
	repoRelativePath, err := h.gitRepoRelativePath(fragment.FilePath)
	if err != nil {
		return nil, err
	}
	logOutput, err := h.runGitCommand(
		`log`,
		`--follow`,
		`--date=iso-strict`,
		`--format=%H%x1f%P%x1f%ad%x1f%s`,
		`--`,
		repoRelativePath,
	)
	if err != nil {
		if isGitMissingPathError(err) {
			return []map[string]any{}, nil
		}
		return nil, err
	}
	lines := strings.Split(normalizeLineBreaks(logOutput), "\n")
	result := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == `` {
			continue
		}
		item, buildErr := h.buildGitHistoryEntry(line, repoRelativePath)
		if buildErr != nil {
			return nil, buildErr
		}
		if item != nil {
			result = append(result, item)
		}
	}
	return result, nil
}

func (h *Service) gitRepoRelativePath(filePath string) (string, error) {
	repoRoot, err := h.gitRepoRoot()
	if err != nil {
		return ``, err
	}
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == `` {
		return ``, fmt.Errorf(`未检测到 git 仓库根目录`)
	}
	absoluteFilePath := filePath
	if !filepath.IsAbs(absoluteFilePath) {
		absoluteFilePath, err = filepath.Abs(absoluteFilePath)
		if err != nil {
			return ``, err
		}
	}
	absoluteRepoRoot := repoRoot
	if !filepath.IsAbs(absoluteRepoRoot) {
		absoluteRepoRoot, err = filepath.Abs(absoluteRepoRoot)
		if err != nil {
			return ``, err
		}
	}
	relativePath, err := filepath.Rel(absoluteRepoRoot, absoluteFilePath)
	if err != nil {
		return ``, err
	}
	relativePath = filepath.ToSlash(relativePath)
	if relativePath == `.` || strings.HasPrefix(relativePath, `../`) {
		return ``, fmt.Errorf(`片段文件不在 git 仓库目录内`)
	}
	return relativePath, nil
}

func (h *Service) buildGitHistoryEntry(line, relativePath string) (map[string]any, error) {
	partList := strings.Split(line, "\x1f")
	if len(partList) < 4 {
		return nil, nil
	}
	commitHash := strings.TrimSpace(partList[0])
	parentHashes := strings.Fields(strings.TrimSpace(partList[1]))
	commitTimeRaw := strings.TrimSpace(partList[2])
	commitMessage := strings.TrimSpace(partList[3])
	newContent, err := h.readGitFileContent(commitHash, relativePath)
	if err != nil {
		return nil, err
	}
	oldContent := ``
	if len(parentHashes) > 0 {
		oldContent, err = h.readGitFileContent(parentHashes[0], relativePath)
		if err != nil && !isGitMissingPathError(err) {
			return nil, err
		}
	}
	oldTitle := extractFragmentTitleFromContent(oldContent)
	newTitle := extractFragmentTitleFromContent(newContent)
	createTimeDesc := commitTimeRaw
	if parsedTime, err := time.Parse(time.RFC3339, commitTimeRaw); err == nil {
		createTimeDesc = parsedTime.Format(`2006-01-02 15:04:05`)
	}
	return map[string]any{
		`id`:               commitHash,
		`commit_hash`:      commitHash,
		`create_time_desc`: createTimeDesc,
		`change_desc`:      commitMessage,
		`title_old`:        oldTitle,
		`title_new`:        newTitle,
		`content_old`:      strings.TrimSpace(oldContent),
		`content_new`:      strings.TrimSpace(newContent),
		`tags_old_list`:    []string{},
		`tags_new_list`:    []string{},
	}, nil
}

func (h *Service) readGitFileContent(revision, relativePath string) (string, error) {
	if strings.TrimSpace(revision) == `` {
		return ``, nil
	}
	return h.runGitCommand(`show`, revision+`:`+relativePath)
}

func (h *Service) runGitCommand(args ...string) (string, error) {
	repoRoot, err := h.gitRepoRoot()
	if err != nil {
		return ``, err
	}
	return runGitCommandInDir(repoRoot, args...)
}

func (h *Service) gitRepoRoot() (string, error) {
	return runGitCommandInDir(h.root, `rev-parse`, `--show-toplevel`)
}

func runGitCommandInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command(`git`, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	text := normalizeLineBreaks(string(output))
	if err != nil {
		return ``, fmt.Errorf(`git %s 失败: %s`, strings.Join(args, ` `), strings.TrimSpace(text))
	}
	return strings.TrimSpace(text), nil
}

func isGitMissingPathError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, `does not have that path`) ||
		strings.Contains(message, `exists on disk, but not in`) ||
		strings.Contains(message, `unknown revision or path not in the working tree`) ||
		strings.Contains(message, `fatal: ambiguous argument`)
}

func extractFragmentTitleFromContent(content string) string {
	content = strings.TrimSpace(content)
	if content == `` {
		return ``
	}
	meta, markdownBody, err := parseFrontMatter(content)
	if err == nil {
		title := NormalizeFragmentTitle(meta.Title, markdownBody)
		if strings.TrimSpace(title) != `` {
			return title
		}
	}
	return NormalizeFragmentTitle(``, content)
}
