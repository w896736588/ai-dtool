package controller

import "testing"

func TestParseApiIDs(t *testing.T) {
	// 中文注释：校验数组入参与非法值过滤逻辑。
	got := parseApiIDs([]any{1, 2, 2, 0, -1})
	want := []int{1, 2}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("got[%d] = %d, want %d", index, got[index], want[index])
		}
	}

	// 中文注释：校验字符串入参与去重逻辑。
	got = parseApiIDs("3,2,2,0,a,1")
	want = []int{3, 2, 1}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("got[%d] = %d, want %d", index, got[index], want[index])
		}
	}

	if len(parseApiIDs("")) != 0 {
		t.Fatal("expected empty string returns empty ids")
	}
	if len(parseApiIDs([]any{})) != 0 {
		t.Fatal("expected empty array returns empty ids")
	}
}

func TestBuildApiBasicInfo(t *testing.T) {
	input := map[string]any{
		"id":               1,
		"folder_id":        2,
		"collection_id":    3,
		"name":             "获取用户",
		"method":           "GET",
		"url":              "https://example.com/user",
		"desc":             "测试接口",
		"env_id":           4,
		"weight":           5,
		"create_time":      6,
		"update_time":      7,
		"headers":          "{}",
		"query_params":     "[]",
		"body_form":        "[]",
		"body_json":        "{}",
		"response_take":    "[]",
		"take_result":      "[]",
		"take_result_desc": "desc",
		"last_result":      "{}",
	}

	got := buildApiBasicInfo(input)
	if got["id"] != 1 {
		t.Fatalf("id = %v, want 1", got["id"])
	}
	if got["type"] != "api" {
		t.Fatalf("type = %v, want api", got["type"])
	}
	if got["uniqueid"] != "api1" {
		t.Fatalf("uniqueid = %v, want api1", got["uniqueid"])
	}

	disallowedKeys := []string{
		"headers",
		"query_params",
		"body_form",
		"body_json",
		"response_take",
		"take_result",
		"take_result_desc",
		"last_result",
	}
	for _, key := range disallowedKeys {
		if _, ok := got[key]; ok {
			t.Fatalf("unexpected key %s exists", key)
		}
	}
}

func TestBuildCollectionBasicInfo(t *testing.T) {
	input := map[string]any{
		"id":         11,
		"name":       "用户中心",
		"child_count": 4,
		"create_time": 100,
		"update_time": 200,
	}

	got := buildCollectionBasicInfo(input)
	if got["id"] != 11 {
		t.Fatalf("id = %v, want 11", got["id"])
	}
	if got["type"] != "collection" {
		t.Fatalf("type = %v, want collection", got["type"])
	}
	if got["uniqueid"] != "collection11" {
		t.Fatalf("uniqueid = %v, want collection11", got["uniqueid"])
	}
	if got["child_count"] != 4 {
		t.Fatalf("child_count = %v, want 4", got["child_count"])
	}
}

func TestBuildFolderBasicInfo(t *testing.T) {
	input := map[string]any{
		"id":            21,
		"collection_id": 11,
		"name":          "登录接口",
		"child_count":   7,
		"create_time":   300,
		"update_time":   400,
	}

	got := buildFolderBasicInfo(input)
	if got["id"] != 21 {
		t.Fatalf("id = %v, want 21", got["id"])
	}
	if got["type"] != "folder" {
		t.Fatalf("type = %v, want folder", got["type"])
	}
	if got["uniqueid"] != "folder21" {
		t.Fatalf("uniqueid = %v, want folder21", got["uniqueid"])
	}
	if got["child_count"] != 7 {
		t.Fatalf("child_count = %v, want 7", got["child_count"])
	}
}

func TestSortAPIListByIDs(t *testing.T) {
	list := []map[string]any{
		{"id": 2, "name": "b"},
		{"id": 1, "name": "a"},
		{"id": 3, "name": "c"},
	}
	got := sortAPIListByIDs(list, []int{3, 1, 2})
	want := []int{3, 1, 2}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index]["id"] != want[index] {
			t.Fatalf("got[%d].id = %v, want %d", index, got[index]["id"], want[index])
		}
	}
}
