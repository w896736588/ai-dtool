package business

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type MemoryGit struct {
}

func NewMemoryGit() *MemoryGit {
	return &MemoryGit{}
}

func (h *MemoryGit) IsGitRepo(dir string) (bool, error) {
	output, err := h.run(dir, `rev-parse`, `--is-inside-work-tree`)
	if err != nil {
		if strings.Contains(output, `not a git repository`) {
			return false, nil
		}
		return false, err
	}
	return strings.TrimSpace(output) == `true`, nil
}

func (h *MemoryGit) RootDir(dir string) (string, error) {
	output, err := h.run(dir, `rev-parse`, `--show-toplevel`)
	if err != nil {
		return ``, err
	}
	return strings.TrimSpace(output), nil
}

func (h *MemoryGit) Pull(dir string) error {
	_, err := h.run(dir, `pull`)
	return err
}

func (h *MemoryGit) HasFileChanges(dir, fileName string) (bool, error) {
	output, err := h.run(dir, `status`, `--porcelain`, `--`, fileName)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != ``, nil
}

func (h *MemoryGit) ListChangedFiles(dir, fileName string) ([]string, error) {
	output, err := h.run(dir, `status`, `--porcelain`, `--`, fileName)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(output, "\r", ""), "\n")
	result := make([]string, 0, len(lines))
	seen := make(map[string]bool)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == `` {
			continue
		}
		if len(line) < 4 {
			continue
		}
		pathText := strings.TrimSpace(line[3:])
		if pathText == `` {
			continue
		}
		if strings.Contains(pathText, ` -> `) {
			parts := strings.Split(pathText, ` -> `)
			pathText = strings.TrimSpace(parts[len(parts)-1])
		}
		pathText = filepath.ToSlash(pathText)
		if !seen[pathText] {
			seen[pathText] = true
			result = append(result, pathText)
		}
	}
	return result, nil
}

func (h *MemoryGit) AddFile(dir, fileName string) error {
	_, err := h.run(dir, `add`, `--`, fileName)
	return err
}

func (h *MemoryGit) Commit(dir, fileName, message string) error {
	output, err := h.run(dir, `commit`, `-m`, message, `--`, fileName)
	if err != nil && strings.Contains(output, `nothing to commit`) {
		return nil
	}
	return err
}

func (h *MemoryGit) Push(dir string) error {
	_, err := h.run(dir, `push`)
	return err
}

func (h *MemoryGit) run(dir string, args ...string) (string, error) {
	cmd := exec.Command(`git`, args...)
	cmd.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	output := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
	if err != nil {
		return output, fmt.Errorf(`git %s 失败: %s`, strings.Join(args, ` `), output)
	}
	return output, nil
}
