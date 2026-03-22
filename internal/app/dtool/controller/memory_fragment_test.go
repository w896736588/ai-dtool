package controller

import (
	"strings"
	"testing"
)

func TestBuildMemoryArrangeUserPrompt(t *testing.T) {
	prompt := buildMemoryArrangeUserPrompt(`整理结构，不改内容`, `测试标题`, "# 标题\n\n正文")
	if !strings.Contains(prompt, "整理结构，不改内容") {
		t.Fatalf("buildMemoryArrangeUserPrompt() missing custom prompt: %q", prompt)
	}
	if !strings.Contains(prompt, "测试标题") {
		t.Fatalf("buildMemoryArrangeUserPrompt() missing title: %q", prompt)
	}
	if !strings.Contains(prompt, "```markdown") {
		t.Fatalf("buildMemoryArrangeUserPrompt() missing markdown fence: %q", prompt)
	}
}

func TestStripMarkdownCodeFence(t *testing.T) {
	got := stripMarkdownCodeFence("```markdown\n# 标题\n\n正文\n```")
	want := "# 标题\n\n正文"
	if got != want {
		t.Fatalf("stripMarkdownCodeFence() = %q, want %q", got, want)
	}
}
