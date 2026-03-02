package plw

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeFileInfo struct {
	isDir bool
}

func (f fakeFileInfo) Name() string       { return "" }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() os.FileMode  { return 0 }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return f.isDir }
func (f fakeFileInfo) Sys() any           { return nil }

// TestResolveNodePathWithDeps 验证 Node 路径解析优先级与回退逻辑
func TestResolveNodePathWithDeps(t *testing.T) {
	fileNodePath := filepath.Join(`C:`, `tools`, `node`, `node.exe`)
	dirNodePath := filepath.Join(`D:`, `runtime`, `node`)
	dirNodeExePath := filepath.Join(dirNodePath, `node.exe`)
	customNodePath := filepath.Join(`E:`, `node-custom`, `node.exe`)
	pathNodePath := filepath.Join(`C:`, `Program Files`, `nodejs`, `node.exe`)
	notExistPath := filepath.Join(`X:`, `not-exist`, `node.exe`)

	tests := []struct {
		name           string
		configNodePath string
		lookPathMap    map[string]string
		existPathMap   map[string]bool
		dirPathMap     map[string]bool
		want           string
	}{
		{
			name:           "配置为可执行文件路径",
			configNodePath: fileNodePath,
			existPathMap: map[string]bool{
				fileNodePath: true,
			},
			want: fileNodePath,
		},
		{
			name:           "配置为目录时自动拼接 node.exe",
			configNodePath: dirNodePath,
			existPathMap: map[string]bool{
				dirNodePath:    true,
				dirNodeExePath: true,
			},
			dirPathMap: map[string]bool{
				dirNodePath: true,
			},
			want: dirNodeExePath,
		},
		{
			name:           "配置为命令名时走 LookPath",
			configNodePath: `node-custom`,
			lookPathMap: map[string]string{
				`node-custom`: customNodePath,
			},
			want: customNodePath,
		},
		{
			name:           "配置无效时回退系统 node",
			configNodePath: notExistPath,
			lookPathMap: map[string]string{
				`node`: pathNodePath,
			},
			want: pathNodePath,
		},
		{
			name:           "全部失败返回空",
			configNodePath: notExistPath,
			want:           ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookPath := func(file string) (string, error) {
				if path, ok := tt.lookPathMap[file]; ok {
					return path, nil
				}
				return ``, errors.New("not found")
			}
			stat := func(name string) (os.FileInfo, error) {
				if !tt.existPathMap[name] {
					return nil, os.ErrNotExist
				}
				return fakeFileInfo{isDir: tt.dirPathMap[name]}, nil
			}
			got := resolveNodePathWithDeps(tt.configNodePath, lookPath, stat)
			if got != tt.want {
				t.Fatalf("resolveNodePathWithDeps() = %q, want %q", got, tt.want)
			}
		})
	}
}
