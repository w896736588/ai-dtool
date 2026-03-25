package business

import (
	"bytes"
	"dev_tool/internal/app/dtool/common"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
)

func TestDataBaseUpRunOnlyLogsExecutedSQLFiles(t *testing.T) {
	tempDir := t.TempDir()
	sqliteClient, err := gsdb.NewSqlite(":memory:", false)
	if err != nil {
		t.Fatalf("NewSqlite() error = %v", err)
	}
	db := &common.CSqlite{Client: sqliteClient}

	migrationDir := filepath.Join(tempDir, "database")
	doneFile := filepath.Join(migrationDir, "2025", "10", "20251001.done.sql")
	todoFile := filepath.Join(migrationDir, "2025", "11", "20251101.todo.sql")
	writeSQLFile(t, doneFile, "create table if not exists tbl_done (id integer);")
	writeSQLFile(t, todoFile, "create table if not exists tbl_todo (id integer);")

	handler := newDatabaseUp(db, migrationDir, "tbl_database_up")
	handler.CheckDataBaseUp()
	if _, err = db.Client.QuickCreate("tbl_database_up", map[string]interface{}{
		"filename": filepath.Base(doneFile),
	}).Exec(); err != nil {
		t.Fatalf("seed upgrade record error = %v", err)
	}

	output := captureStdout(t, func() {
		handler.Run()
	})

	if strings.Contains(output, "开始检查数据库升级表") {
		t.Fatalf("output contains check table log: %s", output)
	}
	if strings.Contains(output, "当前已执行sql文件") {
		t.Fatalf("output contains executed count log: %s", output)
	}
	if strings.Contains(output, "开始扫描升级目录") {
		t.Fatalf("output contains scan dir log: %s", output)
	}
	if strings.Contains(output, filepath.Base(doneFile)) {
		t.Fatalf("output contains already executed file log: %s", output)
	}
	if !strings.Contains(output, filepath.Base(todoFile)) {
		t.Fatalf("output does not contain upgraded file log: %s", output)
	}

	rows, err := db.Client.QuickQuery("tbl_database_up", "filename", nil).Order("id asc").All()
	if err != nil {
		t.Fatalf("query upgrade rows error = %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("upgrade rows count = %d, want 2", len(rows))
	}
}

func writeSQLFile(t *testing.T, filePath, sql string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filePath, []byte(sql), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = writer

	outputCh := make(chan string, 1)
	go func() {
		var buffer bytes.Buffer
		_, _ = buffer.ReadFrom(reader)
		outputCh <- buffer.String()
	}()

	fn()

	_ = writer.Close()
	os.Stdout = oldStdout
	return <-outputCh
}
