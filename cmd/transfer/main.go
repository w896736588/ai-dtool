package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 命令行参数
	host := flag.String("host", "localhost", "管理服务监听地址")
	port := flag.Int("port", 9090, "管理服务监听端口")
	flag.Parse()

	// 创建内存记录存储
	store := NewStore()

	// 加载规则配置
	config, err := NewConfigStore(configFileName)
	if err != nil {
		log.Fatalf("加载配置失败: %s", err.Error())
	}

	// 创建服务管理器
	manager := NewServerManager(store, *port)

	// 启动所有已配置规则的代理服务
	for _, rule := range config.List() {
		ruleCopy := *rule // 复制以避免循环变量问题
		if err := manager.StartRule(&ruleCopy); err != nil {
			log.Printf("[启动失败] 规则 %s (端口 %d): %s", rule.Name, rule.Port, err.Error())
		}
	}

	// 创建管理服务器
	adminAddr := fmt.Sprintf("%s:%d", *host, *port)
	adminSrv := &http.Server{
		Addr:    adminAddr,
		Handler: &adminHandler{store: store, config: config, manager: manager},
	}

	// 在后台启动管理服务器
	go func() {
		fmt.Printf("管理服务启动: http://%s/\n", adminAddr)
		if err := adminSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("管理服务启动失败: %s", err.Error())
		}
	}()

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 优雅关闭
	fmt.Println("\n正在关闭所有服务...")
	manager.StopAll()
	fmt.Println("已退出")
}
