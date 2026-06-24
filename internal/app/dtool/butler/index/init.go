package index

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/w896736588/go-tool/gstool"
)

// SkillInfo 扫描到的 skill 信息。
type SkillInfo struct {
	Name        string   // skill 名称（目录名）
	Description string   // SKILL.md 中的 description
	Functions   []string // 功能索引列表
	Steps       []string // step/ 下的步骤文件名
}

// GenerateStepIndex 扫描 skills/ 目录，生成 step.md 索引内容。
// skillsRoot 为 skills 目录的绝对路径（项目根目录下的 skills/）。
func GenerateStepIndex(skillsRoot string) (string, error) {
	skills, err := scanSkills(skillsRoot)
	if err != nil {
		return ``, fmt.Errorf(`扫描 skills 目录失败: %w`, err)
	}
	if len(skills) == 0 {
		return `# 步骤文件索引\n\n暂无可用步骤文件。`, nil
	}
	return buildStepMarkdown(skills), nil
}

// InitIndex 执行索引初始化：扫描 skills/ → 生成 step.md + capabilities.md + apis.md。
// 返回生成的 step.md 内容和错误。
func InitIndex(skillsRoot, indexPath string) (string, error) {
	// 确保目录存在
	if err := EnsureIndexDir(indexPath); err != nil {
		return ``, fmt.Errorf(`创建索引目录失败: %w`, err)
	}
	// 1. 扫描并生成 step.md
	stepContent, err := GenerateStepIndex(skillsRoot)
	if err != nil {
		return ``, err
	}
	if err := WriteIndexFile(indexPath, StepFileName, stepContent); err != nil {
		return ``, fmt.Errorf(`写入 step.md 失败: %w`, err)
	}
	// 2. 生成 capabilities.md（管家总能力清单）
	capabilitiesContent := GenerateCapabilitiesIndex()
	if err := WriteIndexFile(indexPath, CapabilitiesFileName, capabilitiesContent); err != nil {
		return ``, fmt.Errorf(`写入 capabilities.md 失败: %w`, err)
	}
	// 3. 生成 apis.md（dtool HTTP 接口索引）
	apisContent := GenerateApisIndex()
	if err := WriteIndexFile(indexPath, ApisFileName, apisContent); err != nil {
		return ``, fmt.Errorf(`写入 apis.md 失败: %w`, err)
	}
	return stepContent, nil
}

// scanSkills 扫描 skills/ 下所有子目录，提取 skill 信息。
func scanSkills(skillsRoot string) ([]SkillInfo, error) {
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		return nil, err
	}
	skills := make([]SkillInfo, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillDir := filepath.Join(skillsRoot, entry.Name())
		info := scanSkillDir(skillDir)
		if info != nil {
			skills = append(skills, *info)
		}
	}
	// 按名称排序
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	return skills, nil
}

// scanSkillDir 扫描单个 skill 目录，提取信息。
func scanSkillDir(skillDir string) *SkillInfo {
	skillName := filepath.Base(skillDir)
	info := &SkillInfo{
		Name: skillName,
	}
	// 读取 SKILL.md
	skillMDPath := filepath.Join(skillDir, `SKILL.md`)
	content, err := gstool.FileGetContent(skillMDPath)
	if err != nil {
		// 无 SKILL.md 仍然收集脚本信息
		gstool.FmtPrintlnLogTime(`[butler-index] SKILL.md 不存在 %s`, skillMDPath)
	} else {
		parseSkillMD(content, info)
	}
	// 扫描 step/ 目录
	stepsDir := filepath.Join(skillDir, `step`)
	entries, err := os.ReadDir(stepsDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), `.md`) {
				info.Steps = append(info.Steps, entry.Name())
			}
		}
		sort.Strings(info.Steps)
	}
	// 仅包含有步骤文件的 skill，step.md 索引只列出有操作步骤的模块
	if len(info.Steps) == 0 {
		return nil
	}
	return info
}

// parseSkillMD 解析 SKILL.md 内容，提取 front matter 和功能索引。
func parseSkillMD(content string, info *SkillInfo) {
	// 解析 YAML front matter（--- 之间的内容）
	parts := strings.SplitN(content, `---`, 3)
	if len(parts) >= 3 {
		frontMatter := parts[1]
		// 提取 description
		for _, line := range strings.Split(frontMatter, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, `description:`) {
				info.Description = strings.TrimSpace(strings.TrimPrefix(line, `description:`))
				// 去除可能的引号
				info.Description = strings.Trim(info.Description, `"`)
			}
		}
	}
	// 提取功能索引（## 功能索引 下的列表项）
	body := content
	if len(parts) >= 3 {
		body = parts[2]
	}
	inFunctionsSection := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, `## 功能索引`) {
			inFunctionsSection = true
			continue
		}
		if inFunctionsSection {
			if strings.HasPrefix(trimmed, `## `) || trimmed == `` && len(info.Functions) > 0 {
				break
			}
			if strings.HasPrefix(trimmed, `- `) {
				info.Functions = append(info.Functions, strings.TrimPrefix(trimmed, `- `))
			}
		}
	}
}

// buildStepMarkdown 根据扫描到的 skill 信息构建 step.md 内容。
func buildStepMarkdown(skills []SkillInfo) string {
	var sb strings.Builder
	sb.WriteString(`# 步骤文件索引`)
	totalSteps := 0
	for _, s := range skills {
		totalSteps += len(s.Steps)
	}
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf(`共 %d 个步骤文件。`, totalSteps))
	sb.WriteString("\n\n")

	for _, skill := range skills {
		sb.WriteString(fmt.Sprintf(`## [%s] %s`, skill.Name, skill.Description))
		sb.WriteString("\n\n")
		// 步骤文件列表
		if len(skill.Steps) > 0 {
			sb.WriteString(`- 步骤: `)
			stepPaths := make([]string, len(skill.Steps))
			for i, s := range skill.Steps {
				stepPaths[i] = fmt.Sprintf(`skills/%s/step/%s`, skill.Name, s)
			}
			sb.WriteString(strings.Join(stepPaths, `, `))
			sb.WriteString("\n")
		}
		// 功能索引
		for _, fn := range skill.Functions {
			sb.WriteString(fmt.Sprintf(`- %s`, fn))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// GetSkillsRoot 获取项目 skills 目录的绝对路径。
func GetSkillsRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ``
	}
	rootPath, err := gstool.GetRootPath(wd)
	if err != nil {
		return ``
	}
	return filepath.Join(rootPath, `skills`)
}
