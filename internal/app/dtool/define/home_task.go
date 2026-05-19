package define

const (
	// HomeTaskArchivedNo 表示任务未归档。
	HomeTaskArchivedNo = 0
	// HomeTaskArchivedYes 表示任务已归档。
	HomeTaskArchivedYes = 1
)

const (
	// HomeTaskStatusTodo 表示任务待开始。
	HomeTaskStatusTodo = `待开始`
	// HomeTaskStatusDeveloping 表示任务处于开发中。
	HomeTaskStatusDeveloping = `开发中`
	// HomeTaskStatusDevCompleted 表示任务开发完成。
	HomeTaskStatusDevCompleted = `开发完`
	// HomeTaskStatusSelfTesting 表示任务处于自测中。
	HomeTaskStatusSelfTesting = `自测中`
	// HomeTaskStatusSelfTested 表示任务自测完成。
	HomeTaskStatusSelfTested = `自测完`
	// HomeTaskStatusPendingIntegration 表示任务待对接。
	HomeTaskStatusPendingIntegration = `待对接`
	// HomeTaskStatusIntegrating 表示任务处于对接中。
	HomeTaskStatusIntegrating = `对接中`
	// HomeTaskStatusTesting 表示任务处于测试中。
	HomeTaskStatusTesting = `测试中`
	// HomeTaskStatusReleasing 表示任务处于上线中。
	HomeTaskStatusReleasing = `上线中`
	// HomeTaskStatusOnline 表示任务已上线。
	HomeTaskStatusOnline = `已上线`
	// HomeTaskStatusPendingTest 表示任务待测试。
	HomeTaskStatusPendingTest = `待测试`
	// HomeTaskStatusAbandoned 表示任务已废弃。
	HomeTaskStatusAbandoned = `已废弃`
)

var (
	// HomeTaskStatusList 用于统一校验允许的任务状态。
	HomeTaskStatusList = []string{
		HomeTaskStatusTodo,
		HomeTaskStatusDeveloping,
		HomeTaskStatusDevCompleted,
		HomeTaskStatusSelfTesting,
		HomeTaskStatusSelfTested,
		HomeTaskStatusPendingIntegration,
		HomeTaskStatusIntegrating,
		HomeTaskStatusTesting,
		HomeTaskStatusReleasing,
		HomeTaskStatusOnline,
		HomeTaskStatusPendingTest,
		HomeTaskStatusAbandoned,
	}
)
