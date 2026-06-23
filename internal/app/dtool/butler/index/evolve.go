package index

import (
	"fmt"
	"strings"

	"github.com/w896736588/go-tool/gstool"
)

// EvolveAppend 向 scripts.md 追加新脚本条目（自进化）。
// 当子管家创建了新脚本时，调用此函数将脚本信息追加到索引中。
// 追加前会检查 scripts.md 中是否已存在同名脚本条目，若已存在则跳过。
func EvolveAppend(indexPath, skillName, scriptName, description string) error {
	scriptsContent := ReadIndexFile(indexPath, ScriptsFileName)

	// 检查 scripts.md 中是否已存在同名脚本条目
	if strings.Contains(scriptsContent, scriptName) {
		gstool.FmtPrintlnLogTime(`[butler-evolve] 脚本 %s 已存在于索引中，跳过追加`, scriptName)
		return nil
	}

	entry := buildEvolveEntry(skillName, scriptName, description)
	// 追加到文件末尾
	newContent := scriptsContent
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += "\n" + entry
	if err := WriteIndexFile(indexPath, ScriptsFileName, newContent); err != nil {
		return fmt.Errorf(`追加索引条目失败: %w`, err)
	}
	gstool.FmtPrintlnLogTime(`[butler-evolve] 已追加索引条目 skill=%s script=%s`, skillName, scriptName)
	return nil
}

// buildEvolveEntry 构建自进化追加的索引条目。
func buildEvolveEntry(skillName, scriptName, description string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`## [%s] %s`, skillName, description))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf(`- 脚本: %s`, scriptName))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`- 来源: 自进化生成`))
	sb.WriteString("\n\n")
	return sb.String()
}
