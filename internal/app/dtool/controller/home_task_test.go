package controller

import "testing"

func TestShouldAutoCreateHomeTaskMemoryFragment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		taskID     int
		fragmentID string
		want       bool
	}{
		{
			name:       "create task without selected fragment auto creates",
			taskID:     0,
			fragmentID: ``,
			want:       true,
		},
		{
			name:       "edit task without selected fragment does not auto create",
			taskID:     12,
			fragmentID: ``,
			want:       false,
		},
		{
			name:       "selected fragment never auto creates",
			taskID:     12,
			fragmentID: `6da2b5cd-6f93-442d-80ce-d28dce02dfb1`,
			want:       false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := shouldAutoCreateHomeTaskMemoryFragment(testCase.taskID, testCase.fragmentID)
			if got != testCase.want {
				t.Fatalf("shouldAutoCreateHomeTaskMemoryFragment(%d, %q) = %v, want %v", testCase.taskID, testCase.fragmentID, got, testCase.want)
			}
		})
	}
}

func TestNormalizeHomeTaskMemoryFragmentID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		raw   any
		want  string
	}{
		{name: "uuid id", raw: "6da2b5cd-6f93-442d-80ce-d28dce02dfb1", want: "6da2b5cd-6f93-442d-80ce-d28dce02dfb1"},
		{name: "empty string", raw: "", want: ""},
		{name: "legacy zero string", raw: "0", want: ""},
		{name: "trim spaces", raw: "  abc-123  ", want: "abc-123"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeHomeTaskMemoryFragmentID(testCase.raw)
			if got != testCase.want {
				t.Fatalf("normalizeHomeTaskMemoryFragmentID(%v) = %q, want %q", testCase.raw, got, testCase.want)
			}
		})
	}
}
