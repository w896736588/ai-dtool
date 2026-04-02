package controller

import "testing"

func TestParseDockerComposeServiceNames(t *testing.T) {
	raw := "\x1b[?2004hservice_api\r\nservice-web\r\ninvalid service\r\nservice_api\r\nservice-web\r\n"

	got := parseDockerComposeServiceNames(raw)

	want := []string{"service_api", "service-web"}
	if len(got) != len(want) {
		t.Fatalf("service count = %d, want %d, got=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("service[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseDockerImageRows(t *testing.T) {
	raw := "nginx\tlatest\tsha256:abc123\t2 weeks ago\t187MB\n<none>\t<none>\tsha256:def456\t3 days ago\t120MB\nbad-row\n"

	got := parseDockerImageRows(raw)

	if len(got) != 2 {
		t.Fatalf("image count = %d, want 2, got=%v", len(got), got)
	}
	if got[0]["image_ref"] != "nginx:latest" {
		t.Fatalf("first image_ref = %q, want %q", got[0]["image_ref"], "nginx:latest")
	}
	if got[1]["image_ref"] != "sha256:def456" {
		t.Fatalf("dangling image_ref = %q, want %q", got[1]["image_ref"], "sha256:def456")
	}
}

func TestParseDockerContainerRows(t *testing.T) {
	raw := "c1a2b3\tweb-1\tnginx:latest\trunning\tUp 2 hours\nc9d8e7\tworker-1\tnginx:latest\texited\tExited (0) 3 hours ago\nbad-row\n"

	got := parseDockerContainerRows(raw)

	if len(got) != 2 {
		t.Fatalf("container count = %d, want 2, got=%v", len(got), got)
	}
	if got[0]["container_name"] != "web-1" {
		t.Fatalf("first container_name = %q, want %q", got[0]["container_name"], "web-1")
	}
	if got[1]["state"] != "exited" {
		t.Fatalf("second state = %q, want %q", got[1]["state"], "exited")
	}
}

func TestApplyDockerComposeSshNames(t *testing.T) {
	composeList := []map[string]any{
		{"id": 1, "ssh_id": 10, "name": "api"},
		{"id": 2, "ssh_id": 0, "name": "worker"},
		{"id": 3, "ssh_id": 11, "name": "admin"},
	}
	sshList := []map[string]any{
		{"id": 10, "name": "prod-a"},
		{"id": 11, "name": "prod-b"},
	}

	applyDockerComposeSshNames(composeList, sshList)

	if composeList[0]["ssh_name"] != "prod-a" {
		t.Fatalf("composeList[0].ssh_name = %v, want prod-a", composeList[0]["ssh_name"])
	}
	if composeList[1]["ssh_name"] != "" {
		t.Fatalf("composeList[1].ssh_name = %v, want empty string", composeList[1]["ssh_name"])
	}
	if composeList[2]["ssh_name"] != "prod-b" {
		t.Fatalf("composeList[2].ssh_name = %v, want prod-b", composeList[2]["ssh_name"])
	}
}
