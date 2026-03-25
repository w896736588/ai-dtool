package api

import "testing"

func TestMergeHeadersWithFolderDefaults(t *testing.T) {
	// 中文注释：目录默认请求头应先加载，接口请求头再覆盖同名键。
	folderHeaders := map[string]string{
		"Authorization": "Bearer folder-token",
		"X-Trace":       "folder-trace",
	}
	apiHeaders := map[string]string{
		"Authorization": "Bearer api-token",
		"Content-Type":  "application/json",
	}

	got := mergeHeadersWithFolderDefaults(folderHeaders, apiHeaders)

	if got["Authorization"] != "Bearer api-token" {
		t.Fatalf("Authorization = %q, want api override", got["Authorization"])
	}
	if got["X-Trace"] != "folder-trace" {
		t.Fatalf("X-Trace = %q, want folder default", got["X-Trace"])
	}
	if got["Content-Type"] != "application/json" {
		t.Fatalf("Content-Type = %q, want api content type", got["Content-Type"])
	}
	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3", len(got))
	}
}
