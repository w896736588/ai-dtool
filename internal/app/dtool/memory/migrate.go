package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsdb"
)

const timeLayout = time.RFC3339

// BuildFragmentPath 生成片段文件路径。
func BuildFragmentPath(root string, createdAt time.Time, id string, isDeleted bool, folderName string) string {
	folderName = NormalizeFolderName(folderName)
	bucket := folderName
	if isDeleted {
		bucket = filepath.Join(TrashFolderName, folderName)
	}
	year := createdAt.Format(`2006`)
	month := createdAt.Format(`2006-01`)
	return filepath.Join(root, bucket, year, month, id+`.md`)
}

// MigrateLegacyDB 将旧 sqlite 中的知识片段迁移到文件目录。
func MigrateLegacyDB(ctx context.Context, sourceDBPath, targetRoot string) (*MigrationReport, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	sourceDBPath = strings.TrimSpace(sourceDBPath)
	targetRoot = strings.TrimSpace(targetRoot)
	if sourceDBPath == `` || targetRoot == `` {
		return nil, fmt.Errorf(`sourceDBPath and targetRoot are required`)
	}
	sqliteClient, err := gsdb.NewSqlite(sourceDBPath, false)
	if err != nil {
		return nil, err
	}
	rowList, err := sqliteClient.QueryBySql(`
select
	id,
	title,
	content,
	is_deleted,
	create_time,
	update_time
from tbl_memory_fragment
order by id asc`).All()
	if err != nil {
		return nil, err
	}
	report := &MigrationReport{
		Files: make([]MigrationFile, 0, len(rowList)),
	}
	for _, row := range rowList {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		item := LegacyFragment{
			ID:         cast.ToInt(row[`id`]),
			Title:      cast.ToString(row[`title`]),
			Content:    cast.ToString(row[`content`]),
			IsDeleted:  cast.ToInt(row[`is_deleted`]) > 0,
			CreateTime: cast.ToInt64(row[`create_time`]),
			UpdateTime: cast.ToInt64(row[`update_time`]),
		}
		if err = writeLegacyFragment(targetRoot, item, report); err != nil {
			return nil, err
		}
	}
	return report, nil
}

func writeLegacyFragment(targetRoot string, item LegacyFragment, report *MigrationReport) error {
	createdAt := time.Unix(item.CreateTime, 0)
	updatedAt := time.Unix(item.UpdateTime, 0)
	if item.CreateTime <= 0 {
		createdAt = updatedAt
	}
	if item.UpdateTime <= 0 {
		updatedAt = createdAt
	}
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	newID := legacyFragmentID(item.ID)
	filePath := BuildFragmentPath(targetRoot, createdAt, newID, item.IsDeleted, DefaultFolderName)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	content, err := RenderFragmentMarkdown(Fragment{
		ID:        newID,
		Title:     item.Title,
		Content:   item.Content,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		IsDeleted: item.IsDeleted,
		FilePath:  filePath,
	})
	if err != nil {
		return err
	}
	if err = os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return err
	}
	if item.IsDeleted {
		report.TrashCount += 1
	} else {
		report.ActiveCount += 1
	}
	report.Files = append(report.Files, MigrationFile{
		OldID:    item.ID,
		NewID:    newID,
		FilePath: filePath,
		Deleted:  item.IsDeleted,
	})
	return nil
}

func legacyFragmentID(oldID int) string {
	return fmt.Sprintf("legacy-%d", oldID)
}
