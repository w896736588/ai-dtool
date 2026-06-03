package p_shell

import (
	"errors"
	"sync"
	"time"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsssh"
	"github.com/w896736588/go-tool/gstool"
)

// ClientFactory 连接工厂函数类型，用于解耦连接池与具体的 Shell 管理器实现。
// 调用方在创建连接池时注入此函数，避免包循环依赖。
// sshConfig 由连接池持有并自动传入，工厂只需关注 shellClientId 的路由逻辑。
type ClientFactory func(sshConfig map[string]any, shellClientId string) (*gsssh.SshTerminal, error)

type ClientClose func(shellClientId string)

const (
	// DefaultPoolSize 连接池默认槽位数量
	DefaultPoolSize = 5
	// DefaultCoolDown 连接使用后（含新建后）的冷却等待时间
	DefaultCoolDown = time.Second
	// acquireRetryInterval 获取连接时的轮询间隔
	acquireRetryInterval = 100 * time.Millisecond
)

// PoolConfig 连接池配置
type PoolConfig struct {
	PoolSize int           // 槽位数量，默认 DefaultPoolSize
	CoolDown time.Duration // 冷却时间，默认 DefaultCoolDown
}

// DefaultPoolConfig 返回默认连接池配置
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		PoolSize: DefaultPoolSize,
		CoolDown: DefaultCoolDown,
	}
}

// poolSlot 连接池单个槽位，管理一个 SSH 连接的生命周期和冷却状态。
type poolSlot struct {
	client        *gsssh.SshTerminal // SSH 客户端实例
	clientId      string             //shell client_id
	shellClientId string             // 当前绑定的 client key（用于判断复用）
	sshConfig     map[string]any     // 当前绑定的 SSH 配置
	lastUsedAt    time.Time          // 最后释放/创建时间（用于冷却判断）
	inUse         bool               // 是否正在被占用
}

// ShellPool 通用 SSH 连接池，支持：
//   - 固定大小的槽位数组
//   - 连接复用（相同 shellClientId 命中时直接复用）
//   - 冷却机制：新建或释放后的连接需经过 CoolDown 时间才能再次分配
//   - 阻塞等待：所有槽位忙或未冷却时轮询等待
//   - sshConfig 在初始化时绑定，一个连接池对应一台 SSH 服务器
type ShellPool struct {
	prefix    string
	config    PoolConfig
	sshConfig map[string]any // 初始化时绑定的 SSH 配置（不可变）
	slots     []poolSlot
	factory   ClientFactory // 连接工厂函数（依赖注入）
	close     ClientClose   //释放shell
	mu        sync.Mutex
}

// NewShellPoolWithConfig 使用自定义配置创建 Shell 连接池。
func NewShellPoolWithConfig(prefix string, sshConfig map[string]any, cfg PoolConfig, factory ClientFactory, close ClientClose) *ShellPool {
	sshId := cast.ToString(sshConfig["id"])
	if sshId == "" {
		panic("[ShellPool:" + prefix + "] sshConfig id is empty on init")
	}
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = DefaultPoolSize
	}
	if cfg.CoolDown <= 0 {
		cfg.CoolDown = DefaultCoolDown
	}
	pool := &ShellPool{
		prefix:    prefix,
		config:    cfg,
		sshConfig: sshConfig,
		slots:     make([]poolSlot, cfg.PoolSize),
		factory:   factory,
		close:     close,
	}
	gstool.FmtPrintlnLogTime("[ShellPool:%s] 初始化完成, sshId=%s, 槽位数=%d, 冷却时间=%v", prefix, sshId, cfg.PoolSize, cfg.CoolDown)
	return pool
}

// GetOne 从连接池中获取一个可用的 SSH 客户端。
//
// 规则：
//   - 优先复用已有且空闲的连接（已过冷却期）
//   - 无可用连接时在空闲槽位上新建
//   - 所有槽位忙或未冷却时轮询等待（100ms 间隔）
//   - 新建连接的 lastUsedAt 设为 now，确保首次使用前也经过冷却期
//
// timeout: 最大等待超时时间，<=0 表示无限等待
// 返回: SSH 客户端、槽位索引、错误
func (h *ShellPool) GetOne(timeout time.Duration) (*gsssh.SshTerminal, int, error) {
	deadline := time.Time{}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	for {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return nil, -1, errors.New("[ShellPool:" + h.prefix + "] 获取连接超时")
		}

		h.mu.Lock()
		for i := range h.slots {
			slot := &h.slots[i]
			if slot.inUse {
				continue
			}
			if time.Since(slot.lastUsedAt) < h.config.CoolDown {
				continue
			}
			// 有现成的空闲连接 → 直接复用
			if slot.client != nil {
				slot.inUse = true
				h.mu.Unlock()
				gstool.FmtPrintlnLogTime("[ShellPool:%s] 复用连接 slot=%d", h.prefix, i)
				return slot.client, i, nil
			}
			// 空闲槽位无连接 → 新建
			if h.factory == nil {
				h.mu.Unlock()
				return nil, -1, errors.New("[ShellPool:" + h.prefix + "] client factory is nil")
			}
			shellClientId := "git_group_" + gstool.RandStringAll(10)
			gstool.FmtPrintlnLogTime(`创建连接 %s`, shellClientId)
			client, err := h.factory(h.sshConfig, shellClientId)
			if err != nil {
				h.mu.Unlock()
				return nil, -1, err
			}
			slot.client = client
			slot.clientId = shellClientId
			slot.sshConfig = h.sshConfig
			slot.lastUsedAt = time.Now()
			slot.inUse = true
			h.mu.Unlock()
			gstool.FmtPrintlnLogTime("[ShellPool:%s] 新建连接 slot=%d", h.prefix, i)
			return client, i, nil
		}
		h.mu.Unlock()

		time.Sleep(acquireRetryInterval)
	}
}

// ReleaseByIndex 根据索引释放指定槽位。
func (h *ShellPool) ReleaseByIndex(idx int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if idx < 0 || idx >= len(h.slots) {
		return
	}
	h.slots[idx].inUse = false
	h.slots[idx].lastUsedAt = time.Now()
	gstool.FmtPrintlnLogTime("[ShellPool:%s] 按索引归还连接 slot=%d", h.prefix, idx)
}

// Close 关闭连接池中所有连接并释放资源。
func (h *ShellPool) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := range h.slots {
		h.close(h.slots[i].clientId)
	}
	h.slots = nil
	gstool.FmtPrintlnLogTime("[ShellPool:%s] 已关闭", h.prefix)
}

// Size 返回连接池总槽数。
func (h *ShellPool) Size() int {
	return len(h.slots)
}

// InUseCount 返回当前正在使用的连接数。
func (h *ShellPool) InUseCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	count := 0
	for i := range h.slots {
		if h.slots[i].inUse {
			count++
		}
	}
	return count
}

// PoolInfo 描述连接池当前状态的快照信息。
type PoolInfo struct {
	Prefix   string     `json:"prefix"`
	PoolSize int        `json:"pool_size"`
	InUse    int        `json:"in_use"`
	Idle     int        `json:"idle"`
	CoolDown string     `json:"cool_down"`
	Slots    []SlotInfo `json:"slots"`
}

// SlotInfo 单个槽位的详细信息。
type SlotInfo struct {
	Index         int    `json:"index"`
	InUse         bool   `json:"in_use"`
	ShellClientId string `json:"shell_client_id"`
	Cooling       bool   `json:"cooling"` // 是否仍在冷却期
}

// Info 返回连接池当前的详细状态信息。
func (h *ShellPool) Info() PoolInfo {
	h.mu.Lock()
	defer h.mu.Unlock()

	info := PoolInfo{
		Prefix:   h.prefix,
		PoolSize: len(h.slots),
		CoolDown: h.config.CoolDown.String(),
		Slots:    make([]SlotInfo, len(h.slots)),
	}

	for i, slot := range h.slots {
		si := SlotInfo{
			Index:         i,
			InUse:         slot.inUse,
			ShellClientId: slot.shellClientId,
			Cooling:       !slot.inUse && time.Since(slot.lastUsedAt) < h.config.CoolDown,
		}
		info.Slots[i] = si
		if slot.inUse {
			info.InUse++
		} else if si.Cooling {
			// 冷却中的算半空闲
		} else {
			info.Idle++
		}
	}

	return info
}
