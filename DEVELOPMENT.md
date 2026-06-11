# 计划
1. 增加知识库检索（通过接口实现），可以在内置agent和外部编辑器直接使用
2. 完整的工作流git状态展示（或者操作？）
3. 调整工作流程的sse，跟任务ID绑定
4. 流式机器人

# 开发与运行说明

本文档保留项目开发、运行和打包相关说明，方便仓库首页 README 专注于系统功能介绍。

## 环境准备

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOPRIVATE=gitee.com

# task 安装
go install github.com/go-task/task/v3/cmd/task@latest

# air 监听启动
go install github.com/air-verse/air@latest
```

## 开发时启动命令

```bash
# 启动服务，启动后前端变更后都会自动热更新
task dun-dev-company

# 前端开发地址
http://localhost:8080
```

## 发布版启动命令

```bash
# windows
网页版.bat

# linux
web.sh

# macos
web.command

# 默认访问地址
http://localhost:17170
```

## 编译打包命令

```bash
# Windows Web 发行包
task package-windows -- 20260101

# Linux Web 发行包
task package-linux -- 20260101

# macOS Web 发行包
task package-macos -- 20260101

# 后台执行
nohup ./dtool --ConfigFile=xxxx >> /var/log/xxxx.$(date +%Y%m%d).log 2>&1 &
```
