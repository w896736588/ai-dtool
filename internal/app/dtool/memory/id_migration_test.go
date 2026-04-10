package memory

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"dev_tool/internal/app/dtool/common"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"github.com/spf13/cast"
)

func TestMigrateNumericFragmentIDsRenamesFilesAndReturnsMapping(t *testing.T) {
	root := t.TempDir()
	activeCreatedAt := time.Date(2026, 4, 10, 8, 30, 0, 0, time.UTC)
	activeOldID := "1775811447303500600"
	activeOldPath := BuildFragmentPath(root, activeCreatedAt, activeOldID, false)
	writeFragmentFixture(t, activeOldPath, Fragment{
		ID:        activeOldID,
		Title:     "数字片段",
		Content:   "# 数字片段\n\n正文",
		CreatedAt: activeCreatedAt,
		UpdatedAt: activeCreatedAt.Add(5 * time.Minute),
	})

	trashCreatedAt := time.Date(2026, 4, 11, 8, 30, 0, 0, time.UTC)
	trashOldID := "1775802937480157700"
	trashOldPath := BuildFragmentPath(root, trashCreatedAt, trashOldID, true)
	writeFragmentFixture(t, trashOldPath, Fragment{
		ID:        trashOldID,
		Title:     "回收站数字片段",
		Content:   "# 回收站数字片段\n\n正文",
		CreatedAt: trashCreatedAt,
		UpdatedAt: trashCreatedAt.Add(5 * time.Minute),
		IsDeleted: true,
	})

	report, err := MigrateNumericFragmentIDs(root)
	if err != nil {
		t.Fatalf("MigrateNumericFragmentIDs() error = %v", err)
	}
	if len(report.Renamed) != 2 {
		t.Fatalf("len(report.Renamed) = %d, want 2", len(report.Renamed))
	}

	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	assertRenamed := func(oldID string, wantDeleted bool) {
		t.Helper()
		newID, ok := report.IDMap[oldID]
		if !ok {
			t.Fatalf("missing IDMap for %q", oldID)
		}
		if !uuidPattern.MatchString(newID) {
			t.Fatalf("newID = %q, want uuid", newID)
		}
		if _, err := os.Stat(BuildFragmentPath(root, activeCreatedAt, oldID, wantDeleted)); !os.IsNotExist(err) {
			t.Fatalf("old file %q should be removed, err = %v", oldID, err)
		}
	}

	assertRenamed(activeOldID, false)
	assertRenamed(trashOldID, true)

	activeNewID := report.IDMap[activeOldID]
	activeNewPath := BuildFragmentPath(root, activeCreatedAt, activeNewID, false)
	activeFragment, err := ParseFragmentFile(activeNewPath, false, true)
	if err != nil {
		t.Fatalf("ParseFragmentFile(active) error = %v", err)
	}
	if activeFragment.Title != "数字片段" {
		t.Fatalf("active title = %q, want %q", activeFragment.Title, "数字片段")
	}

	trashNewID := report.IDMap[trashOldID]
	trashNewPath := BuildFragmentPath(root, trashCreatedAt, trashNewID, true)
	trashFragment, err := ParseFragmentFile(trashNewPath, true, true)
	if err != nil {
		t.Fatalf("ParseFragmentFile(trash) error = %v", err)
	}
	if trashFragment.Title != "回收站数字片段" {
		t.Fatalf("trash title = %q, want %q", trashFragment.Title, "回收站数字片段")
	}
}

func TestReplaceHomeTaskMemoryFragmentIDsUpdatesReferences(t *testing.T) {
	db := createHomeTaskMemoryFragmentDB(t)
	oldID := "1775811447303500600"
	newID := "6da2b5cd-6f93-442d-80ce-d28dce02dfb1"

	if err := db.ReplaceHomeTaskMemoryFragmentIDs(map[string]string{
		oldID: newID,
	}); err != nil {
		t.Fatalf("ReplaceHomeTaskMemoryFragmentIDs() error = %v", err)
	}

	row, err := db.Client.QueryBySql(`select memory_fragment_id from tbl_home_task where id = 1`).One()
	if err != nil {
		t.Fatalf("QueryBySql() error = %v", err)
	}
	if got := cast.ToString(row[`memory_fragment_id`]); got != newID {
		t.Fatalf("memory_fragment_id = %q, want %q", got, newID)
	}
}

func writeFragmentFixture(t *testing.T, path string, fragment Fragment) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	fragment.FilePath = path
	content, err := RenderFragmentMarkdown(fragment)
	if err != nil {
		t.Fatalf("RenderFragmentMarkdown() error = %v", err)
	}
	if err = os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func createHomeTaskMemoryFragmentDB(t *testing.T) *common.CSqlite {
	t.Helper()
	tempDir, err := os.MkdirTemp(``, `home-task-memory-db-*`)
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	dbPath := filepath.Join(tempDir, `frog.db`)
	sqliteClient, err := gsdb.NewSqlite(dbPath, false)
	if err != nil {
		t.Fatalf("NewSqlite() error = %v", err)
	}
	db := &common.CSqlite{Client: sqliteClient}
	if _, err = db.Client.ExecBySql(`
CREATE TABLE "tbl_home_task" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "name" TEXT NOT NULL DEFAULT '',
  "task_status" TEXT NOT NULL DEFAULT '',
  "memory_fragment_id" TEXT NOT NULL DEFAULT '',
  "is_archived" INTEGER NOT NULL DEFAULT 0,
  "start_time" INTEGER NOT NULL DEFAULT 0,
  "last_operated_at" INTEGER NOT NULL DEFAULT 0,
  "create_time" INTEGER NOT NULL DEFAULT 0,
  "update_time" INTEGER NOT NULL DEFAULT 0
);`).Exec(); err != nil {
		t.Fatalf("ExecBySql(schema) error = %v", err)
	}
	if _, err = db.Client.QuickCreate(`tbl_home_task`, map[string]any{
		`name`:               `任务A`,
		`task_status`:        `todo`,
		`memory_fragment_id`: `1775811447303500600`,
	}).Exec(); err != nil {
		t.Fatalf("QuickCreate(home_task) error = %v", err)
	}
	return db
}
