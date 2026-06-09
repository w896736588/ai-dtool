package memory

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var numericFragmentIDPattern = regexp.MustCompile(`^\d+$`)

// FragmentIDMigrationItem 描述单个片段 ID 迁移结果。 // Describes one fragment ID migration result.
type FragmentIDMigrationItem struct {
	OldID   string
	NewID   string
	OldPath string
	NewPath string
	Deleted bool
}

// FragmentIDMigrationReport 汇总数字片段 ID 迁移结果。 // Summarizes numeric fragment ID migration results.
type FragmentIDMigrationReport struct {
	IDMap   map[string]string
	Renamed []FragmentIDMigrationItem
}

// MigrateNumericFragmentIDs 把旧的数字型片段 ID 迁移为 UUID 文件名。 // Rename legacy numeric fragment IDs to UUID file names.
func MigrateNumericFragmentIDs(root string) (*FragmentIDMigrationReport, error) {
	report := &FragmentIDMigrationReport{
		IDMap:   map[string]string{},
		Renamed: make([]FragmentIDMigrationItem, 0),
	}
	for _, bucket := range []struct {
		dir       string
		isDeleted bool
	}{
		{dir: filepath.Join(root, `fragments`), isDeleted: false},
		{dir: filepath.Join(root, `trash`), isDeleted: true},
	} {
		if err := migrateNumericFragmentBucket(root, bucket.dir, bucket.isDeleted, report); err != nil {
			return nil, err
		}
	}
	return report, nil
}

// RepairInvalidFragmentIDs 修复明显非法的片段文件名，例如 "."、".."、空名等。
func RepairInvalidFragmentIDs(root string) (*FragmentIDMigrationReport, error) {
	report := &FragmentIDMigrationReport{
		IDMap:   map[string]string{},
		Renamed: make([]FragmentIDMigrationItem, 0),
	}
	for _, bucket := range []struct {
		dir       string
		isDeleted bool
	}{
		{dir: filepath.Join(root, `fragments`), isDeleted: false},
		{dir: filepath.Join(root, `trash`), isDeleted: true},
	} {
		if err := repairInvalidFragmentBucket(root, bucket.dir, bucket.isDeleted, report); err != nil {
			return nil, err
		}
	}
	return report, nil
}

func migrateNumericFragmentBucket(root, bucketPath string, isDeleted bool, report *FragmentIDMigrationReport) error {
	if err := os.MkdirAll(bucketPath, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(bucketPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || strings.ToLower(filepath.Ext(path)) != `.md` {
			return nil
		}
		oldID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		// 仅迁移纯数字旧 ID，避免误改现有 UUID 与 legacy-* 兼容命名。 // Only migrate pure numeric legacy IDs; keep UUID and legacy-* names untouched.
		if !isNumericFragmentID(oldID) {
			return nil
		}
		fragment, parseErr := ParseFragmentFile(path, isDeleted, false)
		if parseErr != nil {
			return parseErr
		}
		newID, newPath, buildErr := nextAvailableFragmentUUIDPath(root, fragment.CreatedAt, isDeleted)
		if buildErr != nil {
			return buildErr
		}
		if err = os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
			return err
		}
		if err = os.Rename(path, newPath); err != nil {
			return err
		}
		report.IDMap[oldID] = newID
		report.Renamed = append(report.Renamed, FragmentIDMigrationItem{
			OldID:   oldID,
			NewID:   newID,
			OldPath: path,
			NewPath: newPath,
			Deleted: isDeleted,
		})
		return nil
	})
}

func repairInvalidFragmentBucket(root, bucketPath string, isDeleted bool, report *FragmentIDMigrationReport) error {
	if err := os.MkdirAll(bucketPath, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(bucketPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || strings.ToLower(filepath.Ext(path)) != `.md` {
			return nil
		}
		oldID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if IsValidFragmentID(oldID) {
			return nil
		}
		fragment, parseErr := parseFragmentFileAllowInvalidID(path, isDeleted, false)
		if parseErr != nil {
			return parseErr
		}
		newID, newPath, buildErr := nextAvailableFragmentUUIDPath(root, fragment.CreatedAt, isDeleted)
		if buildErr != nil {
			return buildErr
		}
		if err = os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
			return err
		}
		if err = os.Rename(path, newPath); err != nil {
			return err
		}
		report.IDMap[oldID] = newID
		report.Renamed = append(report.Renamed, FragmentIDMigrationItem{
			OldID:   oldID,
			NewID:   newID,
			OldPath: path,
			NewPath: newPath,
			Deleted: isDeleted,
		})
		return nil
	})
}

func parseFragmentFileAllowInvalidID(path string, isDeleted bool, loadContent bool) (*Fragment, error) {
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
	createdAt := parseTimeOrZero(meta.CreatedAt)
	updatedAt := parseTimeOrZero(meta.UpdatedAt)
	if createdAt.IsZero() {
		createdAt = info.ModTime()
	}
	if updatedAt.IsZero() {
		updatedAt = info.ModTime()
	}
	fragment := &Fragment{
		ID:         strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
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

func nextAvailableFragmentUUIDPath(root string, createdAt time.Time, isDeleted bool) (string, string, error) {
	for {
		id := uuid.NewString()
		path := BuildFragmentPath(root, createdAt, id, isDeleted, DefaultFolderName)
		if _, err := os.Stat(path); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return ``, ``, err
		}
		return id, path, nil
	}
}

// isNumericFragmentID 判断片段 ID 是否为旧的纯数字时间戳格式。 // Check whether the fragment ID is a legacy numeric timestamp format.
func isNumericFragmentID(id string) bool {
	return numericFragmentIDPattern.MatchString(strings.TrimSpace(id))
}
