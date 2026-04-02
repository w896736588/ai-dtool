# dev_tool_master

## 功能简介

本工具是面向开发与运维场景的个人使用的本地化工作台，数据存储个人空间

### 模块说明

- 首页：通过快捷命令操作Git等模块
- Redis：用于 Redis 数据查询、键值查看与常用缓存操作。
- Supervisor：用于进程/服务管理，查看运行状态并执行启停相关操作。
- Git：用于代码仓库常用操作与结果查看。
- 自定义网页：配置并打开业务常用网页入口，支持快捷访问。
- 自定义脚本：维护并执行脚本化流程，支持变量参与和结果输出。
- Docker：用于容器与服务相关操作查看与管理。
- 接口开发：用于 API 目录管理、接口编辑、环境变量、调试执行与结果记录。
- 终端输出：统一查看命令执行输出，便于排查与追踪。
- 知识片段：方便的知识存储，搜索功能。
- 配置：维护系统基础配置与模块参数。
- 小工具：提供常用辅助工具（如编码转换、二维码、时间转换等）。

## 环境准备

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOPRIVATE=gitee.com

# gs扩展安装
go env -w GOPRIVATE=gitee.com
# 更新到最新tag
go get -u gitee.com/Sxiaobai/gs/v2@latest
# task安装
go install github.com/go-task/task/v3/cmd/task@latest
# 安装 Wails CLI（用于桌面端调试/构建）：
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

## 启动命令（task）


### Web 端（浏览器模式）

```bash
开发时
# 后端
task run-server-company
# 前端 
task run-web-dev
# 访问地址
http://localhost:8080
```

```bash
# 点击
网页版.bat
# 访问地址
http://localhost:17170
```


## 编译打包命令

```bash
# web端和桌面端
task package-windows
# web端
task package-linux
# web端和桌面端
task package-macos
```
