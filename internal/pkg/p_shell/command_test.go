package p_shell

import (
	"strings"
	"testing"
)

const (
	// dockerLogTruncateCommandSnippet 约束日志清理命令必须直接作用于 Docker 容器日志目录。
	dockerLogTruncateCommandSnippet = `truncate -s 0 /var/lib/docker/containers/*/*-json.log`
	// gitPullOriginCurrentBranchSnippet 约束“拉取当前分支”命令必须在远端拉取时动态解析本地分支。
	gitPullOriginCurrentBranchSnippet = `git pull --quiet origin "$(git symbolic-ref --short -q HEAD)"`
)

func TestDockerContainerLogTruncateCommand(t *testing.T) {
	commandText := NewCommand().Sudo().DockerContainerLogTruncate().GetCommand().ToStr()

	if !strings.Contains(commandText, dockerLogTruncateCommandSnippet) {
		t.Fatalf("DockerContainerLogTruncate command = %q, want contains %q", commandText, dockerLogTruncateCommandSnippet)
	}
}

func TestGitPullOriginCurrentBranchCommand(t *testing.T) {
	commandText := NewCommand().GitPullOriginCurrentBranch().GetCommand().ToStr()

	if !strings.Contains(commandText, gitPullOriginCurrentBranchSnippet) {
		t.Fatalf("GitPullOriginCurrentBranch command = %q, want contains %q", commandText, gitPullOriginCurrentBranchSnippet)
	}
}
