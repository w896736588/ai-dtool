package plw

import "testing"

// TestSanitizeWindowsFilename 验证 Windows 非法文件名字符会被替换
func TestSanitizeWindowsFilename(t *testing.T) {
	got := sanitizeWindowsFilename(`报表:2026/03*04?.xlsx`)
	want := `报表_2026_03_04_.xlsx`
	if got != want {
		t.Fatalf("sanitizeWindowsFilename() = %q, want %q", got, want)
	}
}
