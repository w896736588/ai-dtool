package define

const (
	CronTaskTypeDailyReport = `daily_report`
)

// CronTaskRegistry 定义所有定时任务类型及其名称和执行函数注册信息。
var CronTaskRegistry = map[string]CronTaskDef{
	CronTaskTypeDailyReport: {Name: `AI 生成工作日报`},
}

// DeprecatedCronTaskTypes 已废弃并需在启动时彻底清理的定时任务类型。
// 历史上用于“兜底”同步，相关类型已从 CronTaskRegistry 移除，但数据库中的记录需要删除。
var DeprecatedCronTaskTypes = []string{
	`memory_sync`,  // 同步知识片段（兜底）
	`main_db_sync`, // 同步主库（兜底）
}

// CronTaskDef 描述一种定时任务类型。
type CronTaskDef struct {
	Name string
}
