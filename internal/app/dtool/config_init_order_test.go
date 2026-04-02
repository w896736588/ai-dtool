package dtool

import (
	"errors"
	"reflect"
	"testing"
)

func TestInitBasePreparesMemoryStoreBeforeAnyDatabaseInit(t *testing.T) {
	t.Parallel()

	calls := make([]string, 0, 7)
	restore := stubInitBaseHooks(func(step string) {
		calls = append(calls, step)
	})
	defer restore()

	InitBase(`config`)

	want := []string{
		`initComponent`,
		`prepareMainDBStoreBeforeDB`,
		`prepareMemoryStoreBeforeDB`,
		`initSqlite`,
		`initGin`,
		`initOther`,
		`initPlaywright`,
		`stdLog`,
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("InitBase() call order = %v, want %v", calls, want)
	}
}

func TestInitBasePanicsWhenPrepareMemoryStoreFails(t *testing.T) {
	t.Parallel()

	restore := stubInitBaseHooks(func(string) {})
	defer restore()

	prepareMemoryStoreBeforeDBFunc = func() error {
		panicMessage := errors.New(`prepare memory failed`)
		return panicMessage
	}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("InitBase() panic = nil, want panic")
		}
		if recovered != `prepare memory failed` {
			t.Fatalf("InitBase() panic = %v, want %q", recovered, `prepare memory failed`)
		}
	}()

	InitBase(`config`)
}

func TestInitBasePanicsWhenPrepareMainDBStoreFails(t *testing.T) {
	t.Parallel()

	restore := stubInitBaseHooks(func(string) {})
	defer restore()

	prepareMainDBStoreBeforeDBFunc = func() error {
		panicMessage := errors.New(`prepare main db failed`)
		return panicMessage
	}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("InitBase() panic = nil, want panic")
		}
		if recovered != `prepare main db failed` {
			t.Fatalf("InitBase() panic = %v, want %q", recovered, `prepare main db failed`)
		}
	}()

	InitBase(`config`)
}

// stubInitBaseHooks 统一替换 InitBase 依赖步骤，避免测试执行真实初始化 / replace InitBase steps to avoid real bootstrap in tests.
func stubInitBaseHooks(record func(string)) func() {
	originalInitComponent := initComponentFunc
	originalPrepareMainDBStore := prepareMainDBStoreBeforeDBFunc
	originalPrepareMemoryStore := prepareMemoryStoreBeforeDBFunc
	originalInitSqlite := initSqliteFunc
	originalInitGin := initGinFunc
	originalInitOther := initOtherFunc
	originalInitPlaywright := initPlaywrightFunc
	originalStdLog := stdLogFunc

	initComponentFunc = func(string, string) {
		record(`initComponent`)
	}
	prepareMainDBStoreBeforeDBFunc = func() error {
		record(`prepareMainDBStoreBeforeDB`)
		return nil
	}
	prepareMemoryStoreBeforeDBFunc = func() error {
		record(`prepareMemoryStoreBeforeDB`)
		return nil
	}
	initSqliteFunc = func() {
		record(`initSqlite`)
	}
	initGinFunc = func() {
		record(`initGin`)
	}
	initOtherFunc = func() {
		record(`initOther`)
	}
	initPlaywrightFunc = func() {
		record(`initPlaywright`)
	}
	stdLogFunc = func() {
		record(`stdLog`)
	}

	return func() {
		initComponentFunc = originalInitComponent
		prepareMainDBStoreBeforeDBFunc = originalPrepareMainDBStore
		prepareMemoryStoreBeforeDBFunc = originalPrepareMemoryStore
		initSqliteFunc = originalInitSqlite
		initGinFunc = originalInitGin
		initOtherFunc = originalInitOther
		initPlaywrightFunc = originalInitPlaywright
		stdLogFunc = originalStdLog
	}
}
