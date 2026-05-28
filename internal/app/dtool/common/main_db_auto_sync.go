package common

// MainDBAutoSync 保留类型占位，避免旧引用在本次去自动同步改造中继续携带行为。
type MainDBAutoSync struct{}

func NewMainDBAutoSync() *MainDBAutoSync {
	return &MainDBAutoSync{}
}

func (h *MainDBAutoSync) SyncNow() error {
	return nil
}

func (h *MainDBAutoSync) SyncPendingTaskNow() error {
	return nil
}

func (h *MainDBAutoSync) HasPendingTask() bool {
	return false
}

func (h *MainDBAutoSync) Stop() {}
