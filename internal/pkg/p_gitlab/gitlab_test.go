package p_gitlab

import "testing"

func TestMergeUserOpSummaryBranchActiveToday(t *testing.T) {
	testCases := []struct {
		name    string
		summary mergeUserOpSummary
		want    bool
	}{
		{
			name:    "no activity",
			summary: mergeUserOpSummary{},
			want:    false,
		},
		{
			name: "author commit today",
			summary: mergeUserOpSummary{
				authorCommitToday: true,
			},
			want: true,
		},
		{
			name: "other commit today",
			summary: mergeUserOpSummary{
				otherCommitToday: true,
			},
			want: true,
		},
		{
			name: "author merge like commit today",
			summary: mergeUserOpSummary{
				authorMergeLikeCommitToday: true,
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.summary.branchActiveToday(); got != tc.want {
				t.Fatalf("branchActiveToday() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsMergeLikeCommit(t *testing.T) {
	gitlab := &TGitlab{}
	testCases := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "classic merge branch",
			message: "Merge branch 'main' into feature/test",
			want:    true,
		},
		{
			name:    "remote tracking merge",
			message: "Merge remote-tracking branch 'origin/main' into feature/test",
			want:    true,
		},
		{
			name:    "pull request merge",
			message: "Merge pull request #42 from team/feature",
			want:    true,
		},
		{
			name:    "bitbucket style merged in",
			message: "Merged in feature/test (pull request #123)",
			want:    true,
		},
		{
			name:    "gitlab see merge request",
			message: "feature/test\n\nSee merge request group/project!1",
			want:    true,
		},
		{
			name:    "merge prefix fallback",
			message: " merge release/2026-03 ",
			want:    true,
		},
		{
			name:    "regular feature commit",
			message: "feat: optimize mr filter",
			want:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := gitlab.isMergeLikeCommit(tc.message); got != tc.want {
				t.Fatalf("isMergeLikeCommit(%q) = %v, want %v", tc.message, got, tc.want)
			}
		})
	}
}
