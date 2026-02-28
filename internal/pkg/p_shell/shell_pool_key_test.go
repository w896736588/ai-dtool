package p_shell

import "testing"

func TestResolvePoolKeyBySshConfigId(t *testing.T) {
	sshConfig := map[string]any{
		"id": "7",
	}
	shellClientId := "7#git_query_current_branch_1772261391886_sse_distribute_id_1772261391886_97147"

	got := resolvePoolKey(sshConfig, shellClientId)
	want := "7"
	if got != want {
		t.Fatalf("resolvePoolKey() = %q, want %q", got, want)
	}
}

func TestResolvePoolKeyByShellClientPrefix(t *testing.T) {
	sshConfig := map[string]any{}
	shellClientId := "9#docker#abc"

	got := resolvePoolKey(sshConfig, shellClientId)
	want := "9"
	if got != want {
		t.Fatalf("resolvePoolKey() = %q, want %q", got, want)
	}
}

