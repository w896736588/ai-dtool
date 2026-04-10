package memory

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dev_tool/internal/app/dtool/common"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
)

func TestMigrateWritesFragmentsAndTrashFiles(t *testing.T) {
	sourceDBPath := createLegacyMemoryDB(t)
	targetRoot := t.TempDir()

	report, err := MigrateLegacyDB(context.Background(), sourceDBPath, targetRoot)
	if err != nil {
		t.Fatalf("MigrateLegacyDB() error = %v", err)
	}
	if report.ActiveCount != 1 {
		t.Fatalf("ActiveCount = %d, want 1", report.ActiveCount)
	}
	if report.TrashCount != 1 {
		t.Fatalf("TrashCount = %d, want 1", report.TrashCount)
	}
	if len(report.Files) != 2 {
		t.Fatalf("len(Files) = %d, want 2", len(report.Files))
	}

	activeFiles, err := filepath.Glob(filepath.Join(targetRoot, `fragments`, `2026`, `2026-04`, `*.md`))
	if err != nil {
		t.Fatalf("Glob(active) error = %v", err)
	}
	if len(activeFiles) != 1 {
		t.Fatalf("len(activeFiles) = %d, want 1", len(activeFiles))
	}
	activeContent := readFile(t, activeFiles[0])
	assertContains(t, activeContent, "title: Redis 缓存")
	assertContains(t, activeContent, "# Redis 缓存")

	trashFiles, err := filepath.Glob(filepath.Join(targetRoot, `trash`, `2026`, `2026-04`, `*.md`))
	if err != nil {
		t.Fatalf("Glob(trash) error = %v", err)
	}
	if len(trashFiles) != 1 {
		t.Fatalf("len(trashFiles) = %d, want 1", len(trashFiles))
	}
	trashContent := readFile(t, trashFiles[0])
	assertContains(t, trashContent, "title: 已删除片段")
	assertContains(t, trashContent, "# 已删除片段")
}

func TestBuildFragmentPathUsesFragmentsAndTrashBuckets(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	activePath := BuildFragmentPath(`C:\memory`, createdAt, `abc`, false)
	trashPath := BuildFragmentPath(`C:\memory`, createdAt, `abc`, true)
	if !strings.Contains(activePath, filepath.Join(`fragments`, `2026`, `2026-01`, `abc.md`)) {
		t.Fatalf("activePath = %q, want fragments bucket", activePath)
	}
	if !strings.Contains(trashPath, filepath.Join(`trash`, `2026`, `2026-01`, `abc.md`)) {
		t.Fatalf("trashPath = %q, want trash bucket", trashPath)
	}
}

func createLegacyMemoryDB(t *testing.T) string {
	t.Helper()
	tempDir, err := os.MkdirTemp(``, `legacy-memory-db-*`)
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	dbPath := filepath.Join(tempDir, `memory.db`)
	sqliteClient, err := gsdb.NewSqlite(dbPath, false)
	if err != nil {
		t.Fatalf("NewSqlite() error = %v", err)
	}
	db := &common.CSqlite{Client: sqliteClient}
	schemaPath := filepath.Join(`..`, `database_memory`, `2026`, `03`, `20260306.记忆片段.sql`)
	sqlBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("ReadFile(schema) error = %v", err)
	}
	if _, err = db.Client.ExecBySql(string(sqlBytes)).Exec(); err != nil {
		t.Fatalf("ExecBySql(schema) error = %v", err)
	}
	activeCreatedAt := time.Date(2026, 4, 10, 8, 30, 0, 0, time.UTC).Unix()
	activeUpdatedAt := time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC).Unix()
	if _, err = db.Client.QuickCreate(`tbl_memory_fragment`, map[string]any{
		`title`:         ``,
		`content`:       "# Redis 缓存\n\n正文内容",
		`content_text`:  `Redis 缓存 正文内容`,
		`is_deleted`:    0,
		`index_status`:  `success`,
		`index_version`: 1,
		`create_time`:   activeCreatedAt,
		`update_time`:   activeUpdatedAt,
	}).Exec(); err != nil {
		t.Fatalf("QuickCreate(active) error = %v", err)
	}
	trashCreatedAt := time.Date(2026, 4, 11, 8, 30, 0, 0, time.UTC).Unix()
	trashUpdatedAt := time.Date(2026, 4, 11, 9, 0, 0, 0, time.UTC).Unix()
	if _, err = db.Client.QuickCreate(`tbl_memory_fragment`, map[string]any{
		`title`:         `已删除片段`,
		`content`:       "# 已删除片段\n\n历史正文",
		`content_text`:  `已删除片段 历史正文`,
		`is_deleted`:    1,
		`index_status`:  `success`,
		`index_version`: 1,
		`create_time`:   trashCreatedAt,
		`update_time`:   trashUpdatedAt,
	}).Exec(); err != nil {
		t.Fatalf("QuickCreate(trash) error = %v", err)
	}
	return dbPath
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return string(content)
}
