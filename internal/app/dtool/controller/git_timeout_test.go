package controller

import (
	"testing"
	"time"
)

func TestGetGitOperationTimeout_BranchChange(t *testing.T) {
	timeout := getGitOperationTimeout(gitOperationBranchChange)
	if timeout != 5*time.Minute {
		t.Fatalf("timeout = %v, want %v", timeout, 5*time.Minute)
	}
}

func TestGetGitOperationTimeout_Pull(t *testing.T) {
	timeout := getGitOperationTimeout(gitOperationPull)
	if timeout != 5*time.Minute {
		t.Fatalf("timeout = %v, want %v", timeout, 5*time.Minute)
	}
}

func TestGetGitOperationTimeout_Default(t *testing.T) {
	timeout := getGitOperationTimeout(`unknown`)
	if timeout != 40*time.Second {
		t.Fatalf("timeout = %v, want %v", timeout, 40*time.Second)
	}
}
