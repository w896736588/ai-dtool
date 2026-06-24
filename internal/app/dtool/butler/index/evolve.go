package index

import (
	"fmt"
	"strings"

	"github.com/w896736588/go-tool/gstool"
)

// EvolveAppend 向 step.md 追加新步骤条目（自进化）。
// 当子管家创建了新步骤文件时，调用此函数将步骤信息追加到索引中。
// 追加前会检查 step.md 中是否已存在同名步骤条目，若已存在则跳过。
func EvolveAppend(indexPath, skillName, stepName, description string) error {
	stepContent := ReadIndexFile(indexPath, StepFileName)

	// 检查 step.md 中是否已存在同名步骤条目
	if strings.Contains(stepContent, stepName) {
		gstool.FmtPrintlnLogTime(`[butler-evolve] 步骤 %s 已存在于索引中，跳过追加`, stepName)
		return nil
	}

	entry := buildEvolveEntry(skillName, stepName, description)
	// 追加到文件末尾
	newContent := stepContent
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += "\n" + entry
	if err := WriteIndexFile(indexPath, StepFileName, newContent); err != nil {
		return fmt.Errorf(`追加索引条目失败: %w`, err)
	}
	gstool.FmtPrintlnLogTime(`[butler-evolve] 已追加索引条目 skill=%s step=%s`, skillName, stepName)
	return nil
}

// buildEvolveEntry 构建自进化追加的索引条目。
func buildEvolveEntry(skillName, stepName, description string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`## [%s] %s`, skillName, description))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf(`- 步骤: %s`, stepName))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`- 来源: 自进化生成`))
	sb.WriteString("\n\n")
	return sb.String()
}
