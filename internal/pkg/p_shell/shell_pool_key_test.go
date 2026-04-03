package p_shell

import "testing"

func TestMaxShellPoolSizeIsTwenty(t *testing.T) {
	if maxShellPoolSize != 20 {
		t.Fatalf("maxShellPoolSize = %d, want 20", maxShellPoolSize)
	}
}

func TestResolvePoolKeyBySshConfigId(t *testing.T) {
	sshConfig := map[string]any{
		"id": "7",
	}
	shellClientId := "7#git_query_current_branch_1772261391886_sse_distribute_id_1772261391886_97147"

	got := resolvePoolKey(sshConfig, shellClientId)
	want := shellClientId
	if got != want {
		t.Fatalf("resolvePoolKey() = %q, want %q", got, want)
	}
}

func TestResolvePoolKeyByShellClientPrefix(t *testing.T) {
	sshConfig := map[string]any{}
	shellClientId := "9#docker#abc"

	got := resolvePoolKey(sshConfig, shellClientId)
	want := shellClientId
	if got != want {
		t.Fatalf("resolvePoolKey() = %q, want %q", got, want)
	}
}

func TestResolvePoolKeyKeepsDifferentDashboardSessionsSeparated(t *testing.T) {
	sshConfig := map[string]any{
		"id": "7",
	}
	gitShellClientID := "7#dashboard_git_1775207067233_sse_distribute_id_1775207067233_65562"
	dockerShellClientID := "7#dashboard_docker_1775207077610_sse_distribute_id_1775207077610_66680"

	gitPoolKey := resolvePoolKey(sshConfig, gitShellClientID)
	dockerPoolKey := resolvePoolKey(sshConfig, dockerShellClientID)

	if gitPoolKey == dockerPoolKey {
		t.Fatalf("different dashboard sessions should not share pool key: git=%q docker=%q", gitPoolKey, dockerPoolKey)
	}
}
