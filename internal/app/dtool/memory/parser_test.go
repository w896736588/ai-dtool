package memory

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeFragmentTitleFallsBackToFirstH1(t *testing.T) {
	title := NormalizeFragmentTitle(``, "# 标题一\n\n正文内容")
	if title != `标题一` {
		t.Fatalf("NormalizeFragmentTitle() = %q, want %q", title, `标题一`)
	}
}

func TestNormalizeFragmentTitlePrefersFrontMatterTitle(t *testing.T) {
	title := NormalizeFragmentTitle(`显式标题`, "# 标题一\n\n正文内容")
	if title != `显式标题` {
		t.Fatalf("NormalizeFragmentTitle() = %q, want %q", title, `显式标题`)
	}
}

func TestRenderFragmentMarkdownWritesFrontMatter(t *testing.T) {
	createdAt := time.Date(2026, 4, 10, 10, 30, 0, 0, time.FixedZone(`CST`, 8*3600))
	updatedAt := createdAt.Add(5 * time.Minute)
	content, err := RenderFragmentMarkdown(Fragment{
		ID:        `fragment-1`,
		Title:     ``,
		Content:   "# Redis 缓存\n\n正文内容",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("RenderFragmentMarkdown() error = %v", err)
	}
	assertContains(t, content, "title: Redis 缓存")
	assertContains(t, content, "created_at: 2026-04-10T10:30:00+08:00")
	assertContains(t, content, "updated_at: 2026-04-10T10:35:00+08:00")
	assertContains(t, content, "# Redis 缓存")
}

func assertContains(t *testing.T, content, want string) {
	t.Helper()
	if !strings.Contains(content, want) {
		t.Fatalf("content missing %q in:\n%s", want, content)
	}
}
