package controller

import (
	"bufio"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_define"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

type browserPortSlotStatus string

const (
	// browserPortSlotStatusIdle 表示固定入口端口仅监听未绑定浏览器。
	browserPortSlotStatusIdle browserPortSlotStatus = "idle"
	// browserPortSlotStatusBinding 表示已为租约预占，等待浏览器调试端口 ready。
	browserPortSlotStatusBinding browserPortSlotStatus = "binding"
	// browserPortSlotStatusBound 表示固定入口端口已经转发到内部调试端口。
	browserPortSlotStatusBound browserPortSlotStatus = "bound"
	// browserPortSlotStatusReleasing 表示正在解绑并释放浏览器租约。
	browserPortSlotStatusReleasing browserPortSlotStatus = "releasing"
	// browserPortSlotStatusError 表示 listener 或绑定流程发生错误。
	browserPortSlotStatusError browserPortSlotStatus = "error"
)

const (
	browserPortPoolPlaywrightInitTimeout = 60
	// browserPortAcquireMaxAttempts 表示分配内部调试端口时最多重试 3 次。 // Retry random internal debug port allocation at most 3 times.
	browserPortAcquireMaxAttempts       = 3
	browserPortReadyTimeout             = 15 * time.Second
	browserPortReadyInterval            = 300 * time.Millisecond
	browserPortHTTPProbeURLTemplate     = "http://127.0.0.1:%d/json/version"
	browserPortHealthCheckInterval      = 3 * time.Second
	browserPortHealthCheckFailThreshold = 3
)

// BrowserPortLeaseInfo 描述一次浏览器端口租约，同时暴露对外入口端口和内部调试端口。
type BrowserPortLeaseInfo struct {
	Config            define.McpChromeDevtoolsConfigItem
	LeaseID           string
	PublicPort        int
	InternalDebugPort int
}

// browserPortItem 表示一个长期存在的固定入口端口槽位。
type browserPortItem struct {
	Config         define.McpChromeDevtoolsConfigItem
	Status         browserPortSlotStatus
	LeaseID        string
	SessionID      string
	BoundDebugPort int
	Forwarder      *tcpForwarder
	LastError      string
}

// browserLease 表示一次真实浏览器调试租约。
type browserLease struct {
	LeaseID           string
	SlotPort          int
	InternalDebugPort int
	SessionID         string
	CreatedAt         time.Time
	stopCh            chan struct{}
	stopOnce          sync.Once
}

// browserPortRuntimeState 输出端口槽位的运行态，供接口层组装响应。
type browserPortRuntimeState struct {
	Status         string
	IsUsed         int
	LeaseID        string
	SessionID      string
	BoundDebugPort int
	LastError      string
}

// tcpForwarder 在固定入口端口上提供透明 TCP 代理。
type tcpForwarder struct {
	listenAddr string
	listener   net.Listener
	slot       *browserPortItem

	connMu sync.Mutex
	conns  map[string]net.Conn
	closed bool
}

// browserPortPool 管理固定入口端口、动态浏览器租约和 TCP 转发器。
type browserPortPool struct {
	mu     sync.Mutex
	items  []*browserPortItem
	leases map[string]*browserLease
}

var globalBrowserPortPool *browserPortPool

// InitBrowserPortPool 初始化 Chrome DevTools 双层端口池。
// 中文：启动时只建立固定入口 listener，不再预热浏览器进程。
// English: Initialize stable public listeners and skip browser prewarm.
func InitBrowserPortPool() {
	for i := 0; i < browserPortPoolPlaywrightInitTimeout; i++ {
		if component.PlaywrightClient != nil && component.PlaywrightClient.Pw != nil {
			break
		}
		time.Sleep(time.Second)
	}
	if component.PlaywrightClient == nil || component.PlaywrightClient.Pw == nil {
		gstool.FmtPrintlnLogTime("[端口池] Playwright 初始化超时，跳过端口监听初始化")
		return
	}

	rows, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_chrome_devtools_config ORDER BY id`,
	).All()
	if err != nil {
		gstool.FmtPrintlnLogTime("[端口池] 读取端口配置失败: %v", err)
		return
	}
	if len(rows) == 0 {
		gstool.FmtPrintlnLogTime("[端口池] 无端口配置，跳过入口监听初始化")
		return
	}

	pool := &browserPortPool{
		items:  make([]*browserPortItem, 0, len(rows)),
		leases: make(map[string]*browserLease),
	}
	for _, row := range rows {
		item := &browserPortItem{
			Config: define.McpChromeDevtoolsConfigItem{
				Id:         cast.ToInt(row["id"]),
				Name:       cast.ToString(row["name"]),
				Port:       cast.ToInt(row["port"]),
				Remark:     cast.ToString(row["remark"]),
				CreateTime: cast.ToInt64(row["create_time"]),
				UpdateTime: cast.ToInt64(row["update_time"]),
			},
			Status: browserPortSlotStatusIdle,
		}
		forwarder, listenErr := newTCPForwarder(item)
		if listenErr != nil {
			item.Status = browserPortSlotStatusError
			item.LastError = listenErr.Error()
			gstool.FmtPrintlnLogTime("[端口池] 端口 %d (%s) 入口监听启动失败: %v", item.Config.Port, item.Config.Name, listenErr)
		} else {
			item.Forwarder = forwarder
			gstool.FmtPrintlnLogTime("[端口池] 端口 %d (%s) 入口监听已启动", item.Config.Port, item.Config.Name)
		}
		pool.items = append(pool.items, item)
	}

	globalBrowserPortPool = pool
	gstool.FmtPrintlnLogTime("[端口池] 初始化完成，共管理 %d 个端口", len(pool.items))
	broadcastChromeDevtoolsPortStatusChange()
}

// Acquire 预占一个空闲固定入口端口并分配内部调试端口。
// 中文：这里只做 slot+lease 分配，不等待浏览器启动。
// English: Reserve slot and allocate internal debug port without launching browser.
func (p *browserPortPool) Acquire() (*BrowserPortLeaseInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, item := range p.items {
		if item.Status != browserPortSlotStatusIdle {
			continue
		}
		if item.Forwarder == nil {
			p.setSlotStatusLocked(item, browserPortSlotStatusError, "listener not ready")
			continue
		}
		leaseID := uuid.NewString()
		lease := &browserLease{
			LeaseID:   leaseID,
			SlotPort:  item.Config.Port,
			CreatedAt: time.Now(),
			stopCh:    make(chan struct{}),
		}
		item.LeaseID = leaseID
		item.SessionID = ""
		item.BoundDebugPort = 0
		p.leases[leaseID] = lease
		p.setSlotStatusLocked(item, browserPortSlotStatusBinding, "")
		// 中文：完全使用当前固定入口端口配置，内部调试端口改为每次随机生成。 // Keep the configured public port and randomize only the internal debug port per lease.
		internalDebugPort, err := p.allocateInternalDebugPortLocked(leaseID)
		if err != nil {
			delete(p.leases, leaseID)
			item.LeaseID = ""
			item.SessionID = ""
			item.BoundDebugPort = 0
			p.setSlotStatusLocked(item, browserPortSlotStatusError, err.Error())
			return nil, err
		}
		gstool.FmtPrintlnLogTime("[端口池] 分配端口 %d (%s)，租约 %s，内部调试端口 %d", item.Config.Port, item.Config.Name, leaseID, internalDebugPort)
		broadcastChromeDevtoolsPortStatusChange()
		return &BrowserPortLeaseInfo{
			Config:            cloneChromeDevtoolsConfigItem(item),
			LeaseID:           leaseID,
			PublicPort:        item.Config.Port,
			InternalDebugPort: internalDebugPort,
		}, nil
	}
	return nil, fmt.Errorf("没有可用的调试端口")
}

// allocateInternalDebugPortLocked 为租约随机挑选内部调试端口，失败最多重试 3 次。 // Pick a random internal debug port for the lease, retrying up to 3 times.
func (p *browserPortPool) allocateInternalDebugPortLocked(leaseID string) (int, error) {
	lease, item := p.findLeaseAndSlotLocked(leaseID)
	if lease == nil || item == nil {
		return 0, fmt.Errorf("租约不存在: %s", leaseID)
	}
	var lastErr error
	for attempt := 1; attempt <= browserPortAcquireMaxAttempts; attempt++ {
		internalDebugPort, err := allocateRandomPort()
		if err != nil {
			lastErr = err
			continue
		}
		lease.InternalDebugPort = internalDebugPort
		gstool.FmtPrintlnLogTime("[端口池] 租约 %s 第 %d 次使用随机内部调试端口 %d", leaseID, attempt, internalDebugPort)
		return internalDebugPort, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("随机端口分配失败")
	}
	return 0, fmt.Errorf("分配内部调试端口失败，已重试 %d 次: %w", browserPortAcquireMaxAttempts, lastErr)
}

// ConfirmLeaseBound 在浏览器调试端口 ready 后，把固定入口端口绑定到内部调试端口。
// 中文：该方法会校验 lease 归属，避免过期绑定写入错误槽位。
// English: Commit slot binding only if lease still owns the slot.
func (p *browserPortPool) ConfirmLeaseBound(leaseID string) error {
	p.mu.Lock()
	lease, item := p.findLeaseAndSlotLocked(leaseID)
	if lease == nil || item == nil {
		p.mu.Unlock()
		return fmt.Errorf("租约不存在: %s", leaseID)
	}
	if item.LeaseID != leaseID || item.Status != browserPortSlotStatusBinding {
		p.mu.Unlock()
		return fmt.Errorf("租约状态不允许绑定: %s", leaseID)
	}
	item.BoundDebugPort = lease.InternalDebugPort
	p.setSlotStatusLocked(item, browserPortSlotStatusBound, "")
	p.mu.Unlock()

	go p.monitorLeaseHealth(leaseID, lease.InternalDebugPort)
	gstool.FmtPrintlnLogTime("[端口池] 端口 %d (%s) 已绑定到调试端口 %d", item.Config.Port, item.Config.Name, lease.InternalDebugPort)
	broadcastChromeDevtoolsPortStatusChange()
	return nil
}

// BindLeaseSession 把 MCP sessionID 记录到租约槽位中，便于状态观测。
func (p *browserPortPool) BindLeaseSession(leaseID, sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	lease, item := p.findLeaseAndSlotLocked(leaseID)
	if lease == nil || item == nil {
		return
	}
	lease.SessionID = sessionID
	item.SessionID = sessionID
}

// Release 根据租约 ID 释放双层端口绑定与连接。
// 中文：Release 设计为幂等，重复释放不会报错。
// English: Release is idempotent for repeated lease cleanup.
func (p *browserPortPool) Release(leaseID string) {
	p.mu.Lock()
	lease, item := p.findLeaseAndSlotLocked(leaseID)
	if lease == nil || item == nil {
		p.mu.Unlock()
		return
	}
	if item.Status == browserPortSlotStatusReleasing || item.Status == browserPortSlotStatusIdle {
		lease.stop()
		delete(p.leases, leaseID)
		item.LeaseID = ""
		item.SessionID = ""
		item.BoundDebugPort = 0
		p.setSlotStatusLocked(item, browserPortSlotStatusIdle, "")
		p.mu.Unlock()
		broadcastChromeDevtoolsPortStatusChange()
		return
	}
	lease.stop()
	item.BoundDebugPort = 0
	item.SessionID = ""
	delete(p.leases, leaseID)
	p.setSlotStatusLocked(item, browserPortSlotStatusReleasing, "")
	forwarder := item.Forwarder
	publicPort := item.Config.Port
	slotName := item.Config.Name
	p.mu.Unlock()

	if forwarder != nil {
		forwarder.CloseActiveConnections()
	}

	p.mu.Lock()
	item.LeaseID = ""
	item.SessionID = ""
	item.BoundDebugPort = 0
	p.setSlotStatusLocked(item, browserPortSlotStatusIdle, "")
	p.mu.Unlock()

	gstool.FmtPrintlnLogTime("[端口池] 端口 %d (%s) 已解绑，恢复空闲监听", publicPort, slotName)
	broadcastChromeDevtoolsPortStatusChange()
}

// MarkLeaseError 在 Acquire/启动异常路径记录错误并回收租约。
func (p *browserPortPool) MarkLeaseError(leaseID string, err error) {
	if err == nil {
		p.Release(leaseID)
		return
	}
	p.mu.Lock()
	_, item := p.findLeaseAndSlotLocked(leaseID)
	if item != nil {
		p.setSlotStatusLocked(item, browserPortSlotStatusError, err.Error())
	}
	p.mu.Unlock()
	broadcastChromeDevtoolsPortStatusChange()
	p.Release(leaseID)
}

func (p *browserPortPool) monitorLeaseHealth(leaseID string, internalDebugPort int) {
	ticker := time.NewTicker(browserPortHealthCheckInterval)
	defer ticker.Stop()

	failCount := 0
	for range ticker.C {
		if !p.isLeaseStillBound(leaseID, internalDebugPort) {
			return
		}
		probeErr := probeBrowserDebugPort(internalDebugPort)
		if probeErr == nil {
			failCount = 0
			continue
		}

		failCount++
		gstool.FmtPrintlnLogTime("[端口池] 调试端口健康检查失败 lease=%s port=%d fail=%d/%d err=%v", leaseID, internalDebugPort, failCount, browserPortHealthCheckFailThreshold, probeErr)
		if failCount < browserPortHealthCheckFailThreshold {
			continue
		}

		// 中文：连续探活失败视为浏览器已关闭，主动触发释放逻辑。
		// English: Consecutive probe failures mean the browser is likely gone, so release the lease.
		gstool.FmtPrintlnLogTime("[端口池] 调试端口连续探活失败，释放租约 lease=%s port=%d", leaseID, internalDebugPort)
		p.Release(leaseID)
		return
	}
}

func (p *browserPortPool) isLeaseStillBound(leaseID string, internalDebugPort int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	lease, item := p.findLeaseAndSlotLocked(leaseID)
	if lease == nil || item == nil {
		return false
	}
	if lease.InternalDebugPort != internalDebugPort {
		return false
	}
	if item.Status != browserPortSlotStatusBound {
		return false
	}
	select {
	case <-lease.stopCh:
		return false
	default:
	}
	return true
}

// RuntimeStateMap 返回固定端口的运行态快照。
func (p *browserPortPool) RuntimeStateMap() map[int]browserPortRuntimeState {
	p.mu.Lock()
	defer p.mu.Unlock()

	stateMap := make(map[int]browserPortRuntimeState, len(p.items))
	for _, item := range p.items {
		stateMap[item.Config.Port] = browserPortRuntimeState{
			Status:         string(item.Status),
			IsUsed:         slotStatusToIsUsed(item.Status),
			LeaseID:        item.LeaseID,
			SessionID:      item.SessionID,
			BoundDebugPort: item.BoundDebugPort,
			LastError:      item.LastError,
		}
	}
	return stateMap
}

// Shutdown 停止全部 listener 和存量连接。
func (p *browserPortPool) Shutdown() {
	p.mu.Lock()
	items := append([]*browserPortItem(nil), p.items...)
	p.mu.Unlock()

	for _, item := range items {
		if item.Forwarder != nil {
			_ = item.Forwarder.Shutdown()
		}
	}
	gstool.FmtPrintlnLogTime("[端口池] 已停止所有入口监听")
}

// ShutdownBrowserPortPool 停止全局端口池 listener。
func ShutdownBrowserPortPool() {
	if globalBrowserPortPool != nil {
		globalBrowserPortPool.Shutdown()
	}
}

func acquireBrowserPort() (*BrowserPortLeaseInfo, error) {
	if globalBrowserPortPool == nil {
		return nil, fmt.Errorf("端口池未初始化，无法分配端口")
	}
	return globalBrowserPortPool.Acquire()
}

func bindBrowserPortLease(leaseID string) error {
	if globalBrowserPortPool == nil {
		return fmt.Errorf("端口池未初始化，无法绑定租约")
	}
	return globalBrowserPortPool.ConfirmLeaseBound(leaseID)
}

func releaseBrowserLease(leaseID string) {
	if globalBrowserPortPool != nil {
		globalBrowserPortPool.Release(leaseID)
	}
}

func setBrowserLeaseSession(leaseID, sessionID string) {
	if globalBrowserPortPool != nil {
		globalBrowserPortPool.BindLeaseSession(leaseID, sessionID)
	}
}

func markBrowserLeaseError(leaseID string, err error) {
	if globalBrowserPortPool != nil {
		globalBrowserPortPool.MarkLeaseError(leaseID, err)
	}
}

func waitForBrowserDebugPortReady(port int) error {
	deadline := time.Now().Add(browserPortReadyTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := probeBrowserDebugPort(port); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(browserPortReadyInterval)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("timeout")
	}
	return fmt.Errorf("等待调试端口 %d 就绪失败: %w", port, lastErr)
}

func probeBrowserDebugPort(port int) error {
	httpClient := &http.Client{Timeout: 2 * time.Second}
	probeURL := fmt.Sprintf(browserPortHTTPProbeURLTemplate, port)
	resp, err := httpClient.Get(probeURL)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status=%d", resp.StatusCode)
	}
	return nil
}

func allocateRandomPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer func() { _ = ln.Close() }()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok || addr.Port <= 0 {
		return 0, fmt.Errorf("随机端口分配失败")
	}
	return addr.Port, nil
}

func slotStatusToIsUsed(status browserPortSlotStatus) int {
	switch status {
	case browserPortSlotStatusBinding, browserPortSlotStatusBound, browserPortSlotStatusReleasing:
		return 1
	default:
		return 0
	}
}

func cloneChromeDevtoolsConfigItem(item *browserPortItem) define.McpChromeDevtoolsConfigItem {
	config := item.Config
	config.IsUsed = slotStatusToIsUsed(item.Status)
	config.Status = string(item.Status)
	config.LeaseID = item.LeaseID
	config.SessionID = item.SessionID
	config.BoundDebugPort = item.BoundDebugPort
	config.LastError = item.LastError
	return config
}

func (p *browserPortPool) findLeaseAndSlotLocked(leaseID string) (*browserLease, *browserPortItem) {
	lease, ok := p.leases[leaseID]
	if !ok {
		return nil, nil
	}
	for _, item := range p.items {
		if item.Config.Port == lease.SlotPort {
			return lease, item
		}
	}
	return nil, nil
}

func (p *browserPortPool) setSlotStatusLocked(item *browserPortItem, status browserPortSlotStatus, errText string) {
	item.Status = status
	item.LastError = strings.TrimSpace(errText)
	item.Config.IsUsed = slotStatusToIsUsed(status)
	item.Config.Status = string(status)
	item.Config.LeaseID = item.LeaseID
	item.Config.SessionID = item.SessionID
	item.Config.BoundDebugPort = item.BoundDebugPort
	item.Config.LastError = item.LastError
}

func newTCPForwarder(slot *browserPortItem) (*tcpForwarder, error) {
	listenAddr := fmt.Sprintf("127.0.0.1:%d", slot.Config.Port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	forwarder := &tcpForwarder{
		listenAddr: listenAddr,
		listener:   listener,
		slot:       slot,
		conns:      make(map[string]net.Conn),
	}
	go forwarder.acceptLoop()
	return forwarder, nil
}

func (f *tcpForwarder) acceptLoop() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			if f.isClosed() || errors.Is(err, net.ErrClosed) {
				return
			}
			gstool.FmtPrintlnLogTime("[端口池] listener %s accept失败: %v", f.listenAddr, err)
			time.Sleep(200 * time.Millisecond)
			continue
		}
		go f.handleConn(conn)
	}
}

func (f *tcpForwarder) handleConn(clientConn net.Conn) {
	connID := uuid.NewString()
	f.addConn(connID, clientConn)
	defer func() {
		f.removeConn(connID)
		_ = clientConn.Close()
	}()

	targetPort := f.currentTargetPort()
	if targetPort <= 0 {
		writeForwarder503(clientConn)
		return
	}

	upstreamConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", targetPort))
	if err != nil {
		writeForwarder503(clientConn)
		return
	}
	upstreamConnID := connID + "-upstream"
	f.addConn(upstreamConnID, upstreamConn)
	defer func() {
		f.removeConn(upstreamConnID)
		_ = upstreamConn.Close()
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(upstreamConn, clientConn)
		if tcpConn, ok := upstreamConn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(clientConn, upstreamConn)
		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}()
	wg.Wait()
}

func (f *tcpForwarder) currentTargetPort() int {
	if globalBrowserPortPool == nil {
		return 0
	}
	globalBrowserPortPool.mu.Lock()
	defer globalBrowserPortPool.mu.Unlock()

	if f.slot == nil {
		return 0
	}
	if f.slot.Status != browserPortSlotStatusBound || f.slot.BoundDebugPort <= 0 {
		return 0
	}
	return f.slot.BoundDebugPort
}

func (f *tcpForwarder) addConn(connID string, conn net.Conn) {
	f.connMu.Lock()
	defer f.connMu.Unlock()
	if f.closed {
		_ = conn.Close()
		return
	}
	f.conns[connID] = conn
}

func (f *tcpForwarder) removeConn(connID string) {
	f.connMu.Lock()
	defer f.connMu.Unlock()
	delete(f.conns, connID)
}

func (f *tcpForwarder) CloseActiveConnections() {
	f.connMu.Lock()
	connList := make([]net.Conn, 0, len(f.conns))
	for _, conn := range f.conns {
		connList = append(connList, conn)
	}
	f.connMu.Unlock()

	for _, conn := range connList {
		_ = conn.Close()
	}
}

func (f *tcpForwarder) Shutdown() error {
	f.connMu.Lock()
	if f.closed {
		f.connMu.Unlock()
		return nil
	}
	f.closed = true
	listener := f.listener
	f.connMu.Unlock()

	f.CloseActiveConnections()
	if listener != nil {
		return listener.Close()
	}
	return nil
}

func (f *tcpForwarder) isClosed() bool {
	f.connMu.Lock()
	defer f.connMu.Unlock()
	return f.closed
}

func writeForwarder503(conn net.Conn) {
	writer := bufio.NewWriter(conn)
	body := "browser not bound\n"
	_, _ = writer.WriteString("HTTP/1.1 503 Service Unavailable\r\n")
	_, _ = writer.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	_, _ = writer.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(body)))
	_, _ = writer.WriteString("Connection: close\r\n\r\n")
	_, _ = writer.WriteString(body)
	_ = writer.Flush()
}

func (l *browserLease) stop() {
	l.stopOnce.Do(func() {
		close(l.stopCh)
	})
}

const chromeDevtoolsPortStatusSsePrefix = `ClientId:`

// broadcastChromeDevtoolsPortStatusChange 向所有已连接的 SSE 客户端广播端口状态变更通知。
// 前端 McpBinding 组件收到后自动刷新配置列表。
func broadcastChromeDevtoolsPortStatusChange() {
	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseChromeDevtoolsPortStatus,
		Data:            "changed",
		Type:            p_define.SseContentTypeMsg,
	})

	for _, item := range gsgin.SseStatus() {
		clientID := strings.TrimSpace(strings.TrimPrefix(item, chromeDevtoolsPortStatusSsePrefix))
		if clientID == "" || clientID == item {
			continue
		}
		sse := gsgin.SseGetByClientId(clientID)
		if sse == nil {
			continue
		}
		_ = sse.SendToChan(msg)
	}
}
