package common

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"github.com/spf13/cast"
)

// TestMemoryFragmentSearchTokens 验证多关键词搜索会按空格拆分并去重。
func TestMemoryFragmentSearchTokens(t *testing.T) {
	handler := &CSqlite{}
	result := handler.memoryFragmentSearchTokens(" Git   冲突  git ")
	expect := []string{"git", "冲突"}
	if !reflect.DeepEqual(result, expect) {
		t.Fatalf("memoryFragmentSearchTokens() = %v, want %v", result, expect)
	}
}

// TestMemoryFragmentSearchScoreMultiKeyword 验证多关键词需要同时命中才算匹配。
func TestMemoryFragmentSearchScoreMultiKeyword(t *testing.T) {
	handler := &CSqlite{}
	query := handler.memoryFragmentNormalizeSearchQuery(" git   冲突 ")
	tokens := handler.memoryFragmentSearchTokens(query)

	matched, score := handler.memoryFragmentSearchScore(
		"keyword",
		query,
		tokens,
		"这是处理冲突时常用的 git rebase 记录",
		[]string{"版本控制"},
		"git 操作备忘",
	)
	if !matched {
		t.Fatalf("memoryFragmentSearchScore() matched = false, want true")
	}
	if score <= 0 {
		t.Fatalf("memoryFragmentSearchScore() score = %d, want > 0", score)
	}

	matched, _ = handler.memoryFragmentSearchScore(
		"keyword",
		query,
		tokens,
		"这里只有 git，没有另一个关键词",
		[]string{"版本控制"},
		"git 操作备忘",
	)
	if matched {
		t.Fatalf("memoryFragmentSearchScore() matched = true, want false")
	}
}

// TestMemoryFragmentTrashLifecycle 验证片段删除、恢复与彻底删除的完整生命周期。
func TestMemoryFragmentTrashLifecycle(t *testing.T) {
	db := newMemoryFragmentTestDB(t)

	saved, err := db.MemoryFragmentSave(0, `知识片段A`, "# 标题\n\n正文内容", []string{`缓存`, `知识库`})
	if err != nil {
		t.Fatalf("MemoryFragmentSave() error = %v", err)
	}

	fragmentID := cast.ToInt(saved[`id`])
	if _, err = db.MemoryFragmentSoftDelete(fragmentID); err != nil {
		t.Fatalf("MemoryFragmentSoftDelete() error = %v", err)
	}

	activeList, err := db.MemoryFragmentList(0)
	if err != nil {
		t.Fatalf("MemoryFragmentList() error = %v", err)
	}
	if len(activeList) != 0 {
		t.Fatalf("MemoryFragmentList() len = %d, want 0 after delete", len(activeList))
	}

	trashList, err := db.MemoryFragmentTrashList(0)
	if err != nil {
		t.Fatalf("MemoryFragmentTrashList() error = %v", err)
	}
	if len(trashList) != 1 {
		t.Fatalf("MemoryFragmentTrashList() len = %d, want 1", len(trashList))
	}
	if cast.ToInt(trashList[0][`id`]) != fragmentID {
		t.Fatalf("trash fragment id = %d, want %d", cast.ToInt(trashList[0][`id`]), fragmentID)
	}

	if _, err = db.MemoryFragmentRestore(fragmentID); err != nil {
		t.Fatalf("MemoryFragmentRestore() error = %v", err)
	}

	activeList, err = db.MemoryFragmentList(0)
	if err != nil {
		t.Fatalf("MemoryFragmentList() after restore error = %v", err)
	}
	if len(activeList) != 1 {
		t.Fatalf("MemoryFragmentList() len after restore = %d, want 1", len(activeList))
	}

	trashList, err = db.MemoryFragmentTrashList(0)
	if err != nil {
		t.Fatalf("MemoryFragmentTrashList() after restore error = %v", err)
	}
	if len(trashList) != 0 {
		t.Fatalf("MemoryFragmentTrashList() len after restore = %d, want 0", len(trashList))
	}

	if _, err = db.MemoryFragmentSoftDelete(fragmentID); err != nil {
		t.Fatalf("MemoryFragmentSoftDelete() second error = %v", err)
	}
	if err = db.MemoryFragmentHardDelete(fragmentID); err != nil {
		t.Fatalf("MemoryFragmentHardDelete() error = %v", err)
	}

	if _, err = db.MemoryFragmentInfo(fragmentID); err == nil {
		t.Fatalf("MemoryFragmentInfo() error = nil, want deleted fragment query to fail")
	}

	trashList, err = db.MemoryFragmentTrashList(0)
	if err != nil {
		t.Fatalf("MemoryFragmentTrashList() after hard delete error = %v", err)
	}
	if len(trashList) != 0 {
		t.Fatalf("MemoryFragmentTrashList() len after hard delete = %d, want 0", len(trashList))
	}
}

// newMemoryFragmentTestDB 构造知识片段专用内存数据库。
func newMemoryFragmentTestDB(t *testing.T) *CSqlite {
	t.Helper()

	sqliteClient, err := gsdb.NewSqlite(`:memory:`, false)
	if err != nil {
		t.Fatalf("NewSqlite() error = %v", err)
	}
	db := &CSqlite{Client: sqliteClient}

	sqlPath := filepath.Join(`..`, `database`, `2026`, `03`, `20260306.记忆片段.sql`)
	sqlBytes, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if _, err = db.Client.ExecBySql(string(sqlBytes)).Exec(); err != nil {
		t.Fatalf("ExecBySql() error = %v", err)
	}

	return db
}
