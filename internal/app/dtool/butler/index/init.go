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
	Scripts     []string // scripts/ 下的脚本文件名
}

// GenerateScriptsIndex 扫描 skills/ 目录，生成 scripts.md 索引内容。
// skillsRoot 为 skills 目录的绝对路径（项目根目录下的 skills/）。
func GenerateScriptsIndex(skillsRoot string) (string, error) {
	skills, err := scanSkills(skillsRoot)
	if err != nil {
		return ``, fmt.Errorf(`扫描 skills 目录失败: %w`, err)
	}
	if len(skills) == 0 {
		return `# 脚本工具索引\n\n暂无可用脚本工具。`, nil
	}
	return buildScriptsMarkdown(skills), nil
}

// InitIndex 执行索引初始化：扫描 skills/ → 生成 scripts.md + capabilities.md + apis.md。
// 返回生成的 scripts.md 内容和错误。
func InitIndex(skillsRoot, indexPath string) (string, error) {
	// 确保目录存在
	if err := EnsureIndexDir(indexPath); err != nil {
		return ``, fmt.Errorf(`创建索引目录失败: %w`, err)
	}
	// 1. 扫描并生成 scripts.md
	scriptsContent, err := GenerateScriptsIndex(skillsRoot)
	if err != nil {
		return ``, err
	}
	if err := WriteIndexFile(indexPath, ScriptsFileName, scriptsContent); err != nil {
		return ``, fmt.Errorf(`写入 scripts.md 失败: %w`, err)
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
	return scriptsContent, nil
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
	// 扫描 scripts/ 目录
	scriptsDir := filepath.Join(skillDir, `scripts`)
	entries, err := os.ReadDir(scriptsDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), `.py`) {
				info.Scripts = append(info.Scripts, entry.Name())
			}
		}
		sort.Strings(info.Scripts)
	}
	// 无脚本也无描述的 skill 跳过
	if info.Description == `` && len(info.Scripts) == 0 && len(info.Functions) == 0 {
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

// buildScriptsMarkdown 根据扫描到的 skill 信息构建 scripts.md 内容。
func buildScriptsMarkdown(skills []SkillInfo) string {
	var sb strings.Builder
	sb.WriteString(`# 脚本工具索引`)
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf(`共 %d 个脚本工具。`, len(skills)))
	sb.WriteString("\n\n")

	for _, skill := range skills {
		sb.WriteString(fmt.Sprintf(`## [%s] %s`, skill.Name, skill.Description))
		sb.WriteString("\n\n")
		// 脚本列表
		if len(skill.Scripts) > 0 {
			sb.WriteString(`- 脚本: `)
			scriptPaths := make([]string, len(skill.Scripts))
			for i, s := range skill.Scripts {
				scriptPaths[i] = fmt.Sprintf(`skills/%s/scripts/%s`, skill.Name, s)
			}
			sb.WriteString(strings.Join(scriptPaths, `, `))
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
