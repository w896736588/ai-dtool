package p_shell

import (
	"strings"
	"testing"
)

const (
	// dockerLogTruncateCommandSnippet 约束日志清理命令必须直接作用于 Docker 容器日志目录。
	dockerLogTruncateCommandSnippet = `truncate -s 0 /var/lib/docker/containers/*/*-json.log`
)

func TestDockerContainerLogTruncateCommand(t *testing.T) {
	commandText := NewCommand().Sudo().DockerContainerLogTruncate().GetCommand().ToStr()

	if !strings.Contains(commandText, dockerLogTruncateCommandSnippet) {
		t.Fatalf("DockerContainerLogTruncate command = %q, want contains %q", commandText, dockerLogTruncateCommandSnippet)
	}
}
