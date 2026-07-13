// Package variable 提供 E2E 用例执行期间的变量解析、生成和上下文存储。
// 设计原则：每个用例执行时创建独立 Context，步骤间共享变量，支持快照回滚。
package variable

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BuiltinGenerators 内置变量生成器（通过 {{$name}} 访问）。
// 注意：本系统变量插值采用 {{var_name}} 语法（与设计文档一致），
// 为兼顾历史用例，额外允许 $name 形式的内置变量名。
var BuiltinGenerators = map[string]func() string{
	"$rand_8":   randomString(8),
	"$rand_16":  randomString(16),
	"$rand_32":  randomString(32),
	"$uuid":     func() string { return uuid.New().String() },
	"$date":     func() string { return time.Now().Format("2006-01-02") },
	"$datetime": func() string { return time.Now().Format("2006-01-02 15:04:05") },
	"$ts":       func() string { return strconv.FormatInt(time.Now().Unix(), 10) },
	"$ts_ms":    func() string { return strconv.FormatInt(time.Now().UnixMilli(), 10) },
}

func randomString(n int) func() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	return func() string {
		b := make([]byte, n)
		t := time.Now().UnixNano()
		for i := range b {
			b[i] = charset[int(t)%len(charset)]
			t = (t*1103515245 + 12345) >> 16
		}
		return string(b)
	}
}

// Context 变量上下文，线程安全。
type Context struct {
	mu      sync.RWMutex
	vars    map[string]string
	history []map[string]string
}

func NewContext(initial map[string]string) *Context {
	c := &Context{vars: make(map[string]string)}
	for k, v := range initial {
		c.vars[k] = v
	}
	return c
}

// Set 设置变量。
func (c *Context) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.vars[key] = value
}

// Get 获取变量。
func (c *Context) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.vars[key]
	return v, ok
}

// Delete 删除变量。
func (c *Context) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.vars, key)
}

// Snapshot 保存当前快照。
func (c *Context) Snapshot() {
	c.mu.Lock()
	defer c.mu.Unlock()
	snap := make(map[string]string, len(c.vars))
	for k, v := range c.vars {
		snap[k] = v
	}
	c.history = append(c.history, snap)
}

// Rollback 回滚到上一次快照。
func (c *Context) Rollback() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.history) == 0 {
		return
	}
	c.vars = c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
}

// Vars 返回所有变量快照（只读）。
func (c *Context) Vars() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]string, len(c.vars))
	for k, v := range c.vars {
		out[k] = v
	}
	return out
}

// Resolver 变量解析器，将字符串中的 {{var_name}} 替换为对应值。
type Resolver struct {
	ctx *Context
}

func NewResolver(ctx *Context) *Resolver {
	return &Resolver{ctx: ctx}
}

// Resolve 执行变量解析。支持：
// - {{name}} 上下文变量
// - {{$builtin}} 内置生成器（每次调用生成新值）
// - 转义：\\{{ 不解析
func (r *Resolver) Resolve(input string) string {
	if input == "" {
		return input
	}
	if !strings.Contains(input, "{{") {
		return input
	}
	var b strings.Builder
	b.Grow(len(input))
	i := 0
	for i < len(input) {
		// 转义：跳过 \\{{
		if i+1 < len(input) && input[i] == '\\' && input[i+1] == '{' && i+2 < len(input) && input[i+2] == '{' {
			b.WriteString("{{")
			i += 3
			continue
		}
		// 寻找 {{ ... }}
		if i+1 < len(input) && input[i] == '{' && input[i+1] == '{' {
			end := strings.Index(input[i+2:], "}}")
			if end < 0 {
				b.WriteString(input[i:])
				break
			}
			name := strings.TrimSpace(input[i+2 : i+2+end])
			// 转义处理
			if name == "" {
				b.WriteString("{{}}")
				i += 2 + end + 2
				continue
			}
			if strings.HasPrefix(name, "$") {
				if gen, ok := BuiltinGenerators[name]; ok {
					b.WriteString(gen())
					i += 2 + end + 2
					continue
				}
				b.WriteString("{{" + name + "}}")
				i += 2 + end + 2
				continue
			}
			if v, ok := r.ctx.Get(name); ok {
				b.WriteString(v)
			} else {
				// 未定义的变量保留原文，避免静默吞错
				b.WriteString("{{" + name + "}}")
			}
			i += 2 + end + 2
			continue
		}
		b.WriteByte(input[i])
		i++
	}
	return b.String()
}

// MustResolve 解析失败时返回错误。
func (r *Resolver) MustResolve(input string) (string, error) {
	out := r.Resolve(input)
	if strings.Contains(out, "{{") && strings.Contains(out, "}}") {
		// 仍含未解析的占位符
		return out, fmt.Errorf("存在未解析变量: %s", input)
	}
	return out, nil
}
