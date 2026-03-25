package controller

import (
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestParseGitStatusEntries(t *testing.T) {
	output := " M app/service.go\nA  docs/readme.md\nR  old.txt -> new.txt\n?? tmp/demo.txt\n"
	got := parseGitStatusEntries(output)
	want := []string{
		"app/service.go",
		"docs/readme.md",
		"new.txt",
		"tmp/demo.txt",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %#v, want %#v", got, want)
	}
}

func TestParseRecentUsageOwners(t *testing.T) {
	now := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)
	output := "" +
		"alice|" + strconv.FormatInt(now.Add(-30*time.Minute).Unix(), 10) + "|app/service.go\n" +
		"bob|" + strconv.FormatInt(now.Add(-90*time.Minute).Unix(), 10) + "|docs/readme.md\n" +
		"carol|" + strconv.FormatInt(now.Add(-3*time.Hour).Unix(), 10) + "|old/demo.txt\n" +
		"alice|bad-ts|bad.txt\n"
	got := parseRecentUsageOwners(output, now, 2*time.Hour)
	want := []string{"alice", "bob"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("owners = %#v, want %#v", got, want)
	}
}

func TestBuildBranchUsageDisplay_RemoteBranchWins(t *testing.T) {
	got := buildBranchUsageDisplay("feature/local", "feature/demo")
	if got != gitBranchUsageUsedDisplay {
		t.Fatalf("display = %q, want %q", got, gitBranchUsageUsedDisplay)
	}
}

func TestBuildBranchUsageDisplay_OwnerFallback(t *testing.T) {
	got := buildBranchUsageDisplay("feature/local", "", "alice", "bob")
	if got != "alice, bob" {
		t.Fatalf("display = %q, want %q", got, "alice, bob")
	}
}

func TestBuildBranchUsageDisplay_NoUsage(t *testing.T) {
	got := buildBranchUsageDisplay("feature/local", "")
	if got != gitBranchUsageNoneDisplay {
		t.Fatalf("display = %q, want %q", got, gitBranchUsageNoneDisplay)
	}
}

func TestBuildBranchUsageDisplay_LocalBranchMissing(t *testing.T) {
	got := buildBranchUsageDisplay("", "feature/demo", "alice")
	if got != gitBranchUsageNoneDisplay {
		t.Fatalf("display = %q, want %q", got, gitBranchUsageNoneDisplay)
	}
}
