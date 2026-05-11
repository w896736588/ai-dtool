package mcp

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"net/http"
	"sync"
	"time"

	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_gin"

	"github.com/gin-gonic/gin"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/playwright-community/playwright-go"
)

// BrowserSession 保存一个已登录的浏览器上下文及其 MCP SSE 服务。
type BrowserSession struct {
	ID           string
	ContextPage  *plw.ContextPage
	Page         *playwright.Page
	McpServer    *mcpserver.MCPServer
	SseServer    *mcpserver.SSEServer
	CreatedAt    time.Time
	LastActiveAt time.Time
	OnClose      func() // 浏览器关闭时的回调（用于释放调试端口等资源）
	snapshot     *AccessibilitySnapshot
	mu           sync.Mutex
	log          *gstool.GsSlog
}

// sessionManager 管理所有浏览器 MCP 会话。
type sessionManager struct {
	sessions sync.Map // sessionID → *BrowserSession
	log      *gstool.GsSlog
}

var globalSessionManager *sessionManager

func initSessionManager() {
	if globalSessionManager == nil {
		globalSessionManager = &sessionManager{
			log: component.PlaywrightClient.Log,
		}
	}
}

// CreateSession 为已登录的浏览器上下文创建 MCP 会话。
func CreateSession(contextPage *plw.ContextPage, page *playwright.Page, baseURL string) (*BrowserSession, error) {
	initSessionManager()

	sessionID := p_common.TBaseClient.GetUnique("mcp-br-")
	session := &BrowserSession{
		ID:           sessionID,
		ContextPage:  contextPage,
		Page:         page,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
		log:          component.PlaywrightClient.Log,
	}

	mcpSrv := mcpserver.NewMCPServer(
		"dtool-browser",
		"1.0.0",
		mcpserver.WithToolCapabilities(true),
	)
	session.McpServer = mcpSrv
	registerTools(mcpSrv, session)

	sseSrv := mcpserver.NewSSEServer(
		mcpSrv,
		mcpserver.WithBaseURL(baseURL),
		mcpserver.WithStaticBasePath("/mcp/ai-browser/"+sessionID),
		mcpserver.WithSSEEndpoint("/sse"),
		mcpserver.WithMessageEndpoint("/message"),
		mcpserver.WithUseFullURLForMessageEndpoint(true),
	)
	session.SseServer = sseSrv

	globalSessionManager.sessions.Store(sessionID, session)
	return session, nil
}

// GetSession 根据 sessionID 查找会话。
func GetSession(sessionID string) (*BrowserSession, bool) {
	initSessionManager()
	val, ok := globalSessionManager.sessions.Load(sessionID)
	if !ok {
		return nil, false
	}
	return val.(*BrowserSession), true
}

// RemoveSession 关闭并移除会话。
func RemoveSession(sessionID string) {
	initSessionManager()
	if val, ok := globalSessionManager.sessions.LoadAndDelete(sessionID); ok {
		s := val.(*BrowserSession)
		if s.ContextPage != nil && s.ContextPage.Context != nil && *s.ContextPage.Context != nil {
			if err := (*s.ContextPage.Context).Close(); err != nil {
				component.PlaywrightClient.Log.Errof("关闭MCP浏览器会话失败: %v", err)
			}
		}
		if s.OnClose != nil {
			s.OnClose()
		}
	}
}

// updateActivity 更新会话活跃时间。
func (s *BrowserSession) updateActivity() {
	s.LastActiveAt = time.Now()
}

// getActivePage 获取当前活跃页面，如果已有页面已关闭则取第一个可用页面。
func (s *BrowserSession) getActivePage() (playwright.Page, error) {
	if s.Page != nil && *s.Page != nil && !(*s.Page).IsClosed() {
		return *s.Page, nil
	}
	pages := s.ContextPage.Pages()
	for _, p := range pages {
		if !p.IsClosed() {
			s.Page = &p
			return p, nil
		}
	}
	return nil, fmt.Errorf("没有可用的页面，浏览器可能已关闭")
}

// setSnapshot 保存最近的 accessibility snapshot。
func (s *BrowserSession) setSnapshot(snapshot *AccessibilitySnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = snapshot
}

// getSnapshot 获取最近的 accessibility snapshot。
func (s *BrowserSession) getSnapshot() *AccessibilitySnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshot
}

// RegisterGinRoutes 将 MCP SSE 的两个端点挂载到 Gin 路由组。
func RegisterGinRoutes(rg *gin.RouterGroup) {
	initSessionManager()
	rg.GET("/:sessionId/sse", handleSSE)
	rg.POST("/:sessionId/message", handleMessage)
}

// RegisterGinRoutesDirect 使用 p_gin.Gin 直接注册 MCP SSE 路由。
func RegisterGinRoutesDirect(tGin *p_gin.Gin) {
	initSessionManager()
	tGin.GinGet("/mcp/ai-browser/:sessionId/sse", handleSSE)
	tGin.GinPost("/mcp/ai-browser/:sessionId/message", handleMessage)
}

func handleSSE(c *gin.Context) {
	sessionID := c.Param("sessionId")
	session, ok := GetSession(sessionID)
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	session.SseServer.SSEHandler().ServeHTTP(c.Writer, c.Request)
}

func handleMessage(c *gin.Context) {
	sessionID := c.Param("sessionId")
	session, ok := GetSession(sessionID)
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	session.SseServer.MessageHandler().ServeHTTP(c.Writer, c.Request)
}

// StartCleanupTimer 启动定时清理过期会话的协程。
func StartCleanupTimer(maxIdle time.Duration) {
	initSessionManager()
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			globalSessionManager.sessions.Range(func(key, value any) bool {
				s := value.(*BrowserSession)
				if now.Sub(s.LastActiveAt) > maxIdle {
					component.PlaywrightClient.Log.Infof("清理过期MCP会话: %s", s.ID)
					RemoveSession(s.ID)
				}
				return true
			})
		}
	}()
}
