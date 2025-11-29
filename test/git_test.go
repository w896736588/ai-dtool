package test

import (
	"dev_tool/base"
	"fmt"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitDiff(t *testing.T) {
	text1 := "Hello, World!"
	text2 := "Hello, Git!"
	if err := diffText(text1, text2); err != nil {
		fmt.Printf("对比失败: %v\n", err)
	}
}

func diffText(text1, text2 string) error {
	base.Component = base.TComponent{
		Env: &base.Env{
			LogPath: `D:\go\cache_manager_api\logs`,
		},
	}
	// 创建临时文件
	tmpFile1 := filepath.Join(base.Component.Env.LogPath, `tmp1.txt`)
	err := gstool.FileCreate(base.Component.Env.LogPath, `tmp1.txt`, text1)
	if err != nil {
		return err
	}
	//defer gstool.FileDelete(tmpFile1)

	tmpFile2 := filepath.Join(base.Component.Env.LogPath, `tmp2.txt`)
	err = gstool.FileCreate(base.Component.Env.LogPath, `tmp2.txt`, text2)
	if err != nil {
		return err
	}

	// 执行 git diff --numstat file1 file2
	cmd := exec.Command("git", "diff", "--no-index", "--shortstat", tmpFile1, tmpFile2)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("执行 Git 命令失败: %v\n输出: %s", err, output)
	}

	// 解析输出（示例：3    1    old_file.txt    new_file.txt）
	stats := strings.TrimSpace(string(output))
	if stats == "" {
		fmt.Println("两个文件内容相同")
		return err
	}

	// 提取变更数据
	var added, deleted int
	var oldFile, newFile string
	_, err = fmt.Sscanf(stats, "%d\t%d\t%s\t%s", &added, &deleted, &oldFile, &newFile)
	if err != nil {
		log.Fatalf("解析 Git 输出失败: %v\n原始输出: %s", err, stats)
	}

	// 打印结果
	fmt.Printf("文件对比结果 (%s → %s):\n", oldFile, newFile)
	fmt.Printf("- 新增行数: %d\n", added)
	fmt.Printf("- 删除行数: %d\n", deleted)
	fmt.Printf("- 变更总结: %d 处修改\n", added+deleted)
	return nil
}
