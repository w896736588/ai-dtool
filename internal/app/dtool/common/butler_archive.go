package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// FindPendingArchiveBySession 根据会话 ID 查找已有的 pending 状态归档记录。
// 返回归档记录 ID，不存在时返回 0。
func (c *CSqlite) FindPendingArchiveBySession(sessionId string) (int, error) {
	row, err := c.Client.QueryBySql(
		`SELECT id FROM tbl_butler_archive WHERE session_id = ? AND status = 'pending' ORDER BY id DESC LIMIT 1`,
		sessionId,
	).One()
	if err != nil || len(row) == 0 {
		return 0, err
	}
	return cast.ToInt(row[`id`]), nil
}

// CreateArchiveRecord 主管家任务完成后提交归档记录。
func (c *CSqlite) CreateArchiveRecord(configId, taskId int, sessionId string, files []string, conversation string) (int, error) {
	// 读取各文件的内容拼入对话记录
	var sb strings.Builder
	sb.WriteString(conversation)
	sb.WriteString("\n\n=== 产生的文件内容 ===\n")
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			sb.WriteString(fmt.Sprintf("[%s] 读取失败: %s\n", f, err.Error()))
		} else {
			sb.WriteString(fmt.Sprintf("[%s]\n%s\n", f, string(data)))
		}
	}
	filesJSON, _ := json.Marshal(files)
	now := time.Now().Unix()
	_, err := c.Client.QuickCreate(`tbl_butler_archive`, map[string]any{
		`config_id`:    configId,
		`task_id`:      taskId,
		`session_id`:   sessionId,
		`files`:        string(filesJSON),
		`conversation`: sb.String(),
		`status`:       `pending`,
		`created_at`:   now,
		`updated_at`:   now,
	}).Exec()
	if err != nil {
		return 0, err
	}
	// 使用 last_insert_rowid() 获取刚插入的自增 ID，比 WHERE session_id 查询更可靠
	one, qErr := c.Client.QueryBySql(`SELECT last_insert_rowid() as id`).One()
	if qErr == nil && len(one) > 0 {
		id := cast.ToInt(one[`id`])
		if id > 0 {
			return id, nil
		}
	}
	return 0, fmt.Errorf(`创建归档记录后无法获取自增ID session=%s`, sessionId)
}

// ListPendingArchives 查询待处理的归档记录。
func (c *CSqlite) ListPendingArchives(limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 10
	}
	return c.Client.QueryBySql(
		`SELECT * FROM tbl_butler_archive WHERE status = 'pending' ORDER BY id ASC LIMIT ?`, limit,
	).All()
}

// UpdateArchiveContent 更新归档记录的文件列表和对话内容（用于同一会话后续轮次合并归档）。
// 仅当记录状态为 pending 时执行更新，避免覆盖正在处理或已完成的记录。
func (c *CSqlite) UpdateArchiveContent(id int, files []string, conversation string, taskId int) error {
	filesJSON, _ := json.Marshal(files)
	_, err := c.Client.QuickUpdate(`tbl_butler_archive`, map[string]any{`id`: id}, map[string]any{
		`files`:        string(filesJSON),
		`conversation`: conversation,
		`task_id`:      taskId,
		`updated_at`:   time.Now().Unix(),
	}).Exec()
	return err
}

// UpdateArchiveStatus 更新归档记录的状态、日志和结果。
func (c *CSqlite) UpdateArchiveStatus(id int, status, logContent, result, resultFile, resultIndex string) error {
	updateData := map[string]any{
		`status`:       status,
		`log`:          logContent,
		`result`:       result,
		`result_file`:  resultFile,
		`result_index`: resultIndex,
		`updated_at`:   time.Now().Unix(),
	}
	_, err := c.Client.QuickUpdate(`tbl_butler_archive`, map[string]any{`id`: id}, updateData).Exec()
	return err
}

// WriteArchiveStep 将归档管家生成的步骤文件内容写入 skills/dtool-butler/step/ 目录。
func WriteArchiveStep(rootPath, stepName, content string) (string, error) {
	stepDir := filepath.Join(rootPath, `skills`, `dtool-butler`, `step`)
	if err := os.MkdirAll(stepDir, 0755); err != nil {
		return ``, err
	}
	filePath := filepath.Join(stepDir, stepName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return ``, err
	}
	return filePath, nil
}

// AppendArchiveIndex 向 step.md 追加一条归档步骤索引条目（一行格式）。
// description 为 AI 产出的完整索引行，格式: `- skills/dtool-butler/step/xxx.md — 任务说明`
// 追加前会检查 step.md 中是否已存在同名步骤条目，若已存在则跳过追加，避免重复索引。
func AppendArchiveIndex(rootPath string, description string) error {
	indexPath := filepath.Join(rootPath, `skills`, `dtool-butler`, `index`, `step.md`)
	// 读取现有内容
	existing, err := os.ReadFile(indexPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	// 去除 description 中的换行和多余空白，确保是单行
	desc := strings.TrimSpace(description)
	// 去掉可能的前缀换行
	desc = strings.TrimLeft(desc, "\n\r")

	// 从索引行中提取步骤文件名（如 "query_repo_branch.md"）
	// 格式: `- skills/dtool-butler/step/xxx.md — ...`
	stepName := extractStepNameFromIndex(desc)
	if stepName != `` && strings.Contains(string(existing), stepName) {
		return fmt.Errorf(`step.md 中已存在步骤 %s 的索引条目，跳过追加以避免重复`, stepName)
	}

	var sb strings.Builder
	sb.Write(existing)
	if len(existing) > 0 && !strings.HasSuffix(string(existing), "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(desc)
	sb.WriteString("\n")
	return os.WriteFile(indexPath, []byte(sb.String()), 0644)
}

// extractStepNameFromIndex 从索引描述行中提取步骤文件名。
// 输入: `- skills/dtool-butler/step/query_repo_branch.md — 任务说明`
// 输出: `query_repo_branch.md`
func extractStepNameFromIndex(desc string) string {
	// 找到 "step/" 后的 .md 文件名
	idx := strings.Index(desc, `step/`)
	if idx == -1 {
		return ``
	}
	rest := desc[idx+len(`step/`):]
	// 截取到空格、—、中文标点之前
	end := strings.IndexAny(rest, ` —-，。`)
	if end == -1 {
		end = len(rest)
	}
	name := strings.TrimSpace(rest[:end])
	// 确保是 .md 文件
	if !strings.HasSuffix(name, `.md`) {
		return ``
	}
	return name
}
