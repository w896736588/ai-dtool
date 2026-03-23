package common

import (
	"testing"
	"time"

	"dev_tool/internal/app/dtool/define"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"github.com/spf13/cast"
)

func TestHomeTaskSaveCreatesTaskWithLastOperateTime(t *testing.T) {
	t.Parallel()

	db := newHomeTaskTestDB(t)

	startTime := time.Date(2026, 3, 23, 0, 0, 0, 0, time.Local).Unix()
	info, err := db.HomeTaskSave(0, `任务A`, define.HomeTaskStatusTodo, `需要和测试确认缓存命中`, startTime)
	if err != nil {
		t.Fatalf("HomeTaskSave() error = %v", err)
	}

	if cast.ToString(info[`name`]) != `任务A` {
		t.Fatalf("task name = %q, want %q", cast.ToString(info[`name`]), `任务A`)
	}
	if cast.ToString(info[`task_status`]) != define.HomeTaskStatusTodo {
		t.Fatalf("task_status = %q, want %q", cast.ToString(info[`task_status`]), define.HomeTaskStatusTodo)
	}
	if cast.ToString(info[`remark`]) != `需要和测试确认缓存命中` {
		t.Fatalf("remark = %q, want %q", cast.ToString(info[`remark`]), `需要和测试确认缓存命中`)
	}
	if cast.ToInt64(info[`start_time`]) != startTime {
		t.Fatalf("start_time = %d, want %d", cast.ToInt64(info[`start_time`]), startTime)
	}
	if cast.ToInt64(info[`last_operated_at`]) <= 0 {
		t.Fatalf("last_operated_at = %d, want > 0", cast.ToInt64(info[`last_operated_at`]))
	}
	if cast.ToString(info[`start_time_desc`]) != `2026-03-23` {
		t.Fatalf("start_time_desc = %q, want %q", cast.ToString(info[`start_time_desc`]), `2026-03-23`)
	}
}

func TestHomeTaskStatusQuickUpdateSetsStartTimeWhenRunning(t *testing.T) {
	t.Parallel()

	db := newHomeTaskTestDB(t)

	info, err := db.HomeTaskSave(0, `任务B`, define.HomeTaskStatusTodo, `联调前补充接口说明`, 0)
	if err != nil {
		t.Fatalf("HomeTaskSave() error = %v", err)
	}
	beforeOperateTime := cast.ToInt64(info[`last_operated_at`])

	time.Sleep(1 * time.Second)

	updated, err := db.HomeTaskStatusQuickUpdate(cast.ToInt(info[`id`]), define.HomeTaskStatusDeveloping)
	if err != nil {
		t.Fatalf("HomeTaskStatusQuickUpdate() error = %v", err)
	}

	if cast.ToString(updated[`task_status`]) != define.HomeTaskStatusDeveloping {
		t.Fatalf("task_status = %q, want %q", cast.ToString(updated[`task_status`]), define.HomeTaskStatusDeveloping)
	}
	if cast.ToInt64(updated[`start_time`]) <= 0 {
		t.Fatalf("start_time = %d, want > 0", cast.ToInt64(updated[`start_time`]))
	}
	if cast.ToInt64(updated[`last_operated_at`]) <= beforeOperateTime {
		t.Fatalf("last_operated_at = %d, want > %d", cast.ToInt64(updated[`last_operated_at`]), beforeOperateTime)
	}
}

func newHomeTaskTestDB(t *testing.T) *CSqlite {
	t.Helper()

	sqliteClient, err := gsdb.NewSqlite(`:memory:`, false)
	if err != nil {
		t.Fatalf("NewSqlite() error = %v", err)
	}
	db := &CSqlite{Client: sqliteClient}

	_, err = db.Client.ExecBySql(`
CREATE TABLE "tbl_home_task" (
  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  "name" TEXT NOT NULL DEFAULT '',
  "task_status" TEXT NOT NULL DEFAULT '',
  "remark" TEXT NOT NULL DEFAULT '',
  "is_archived" INTEGER NOT NULL DEFAULT 0,
  "start_time" INTEGER NOT NULL DEFAULT 0,
  "last_operated_at" INTEGER NOT NULL DEFAULT 0,
  "create_time" INTEGER NOT NULL DEFAULT 0,
  "update_time" INTEGER NOT NULL DEFAULT 0
);`).Exec()
	if err != nil {
		t.Fatalf("create tbl_home_task error = %v", err)
	}

	return db
}
