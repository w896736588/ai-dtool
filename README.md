# dev_tool_master

## 功能简介

本工具是面向开发与运维场景的本地化工作台，支持 Web 端和桌面端两种模式。

### 菜单总览

当前主界面菜单名称如下：

1. 首页
2. Redis
3. Supervisor
4. Git
5. 自定义网页
6. 自定义脚本
7. Docker
8. 接口开发
9. 终端输出
10. 配置
11. 小工具（侧栏底部入口）

说明：`Redis / Supervisor / Git / 自定义网页 / 自定义脚本 / Docker / 接口开发 / 终端输出` 这些菜单会受模块开关控制，可能在部分环境中隐藏。

### 模块说明

1. 首页：系统工作台入口，展示全局状态并承载各模块跳转。
2. Redis：用于 Redis 数据查询、键值查看与常用缓存操作。
3. Supervisor：用于进程/服务管理，查看运行状态并执行启停相关操作。
4. Git：用于代码仓库常用操作与结果查看。
5. 自定义网页：配置并打开业务常用网页入口，支持快捷访问。
6. 自定义脚本：维护并执行脚本化流程，支持变量参与和结果输出。
7. Docker：用于容器与服务相关操作查看与管理。
8. 接口开发：用于 API 目录管理、接口编辑、环境变量、调试执行与结果记录。
9. 终端输出：统一查看命令执行输出，便于排查与追踪。
10. 配置：维护系统基础配置与模块参数。
11. 小工具：提供常用辅助工具（如编码转换、二维码、时间转换等）。

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

## 开发启动命令（task）

```bash
开发时
# 网页版，前后端一起，启动company.ini配置文件
task run-dev-company
# 桌面版,启动company.ini配置文件
task run-dev-wails3-company
```
正式运行时
默认访问地址：`http://localhost:17170/`（以配置中的 `run.ports` 为准）。

## 一键打包

打完包的产物，在build文件夹下
```bash
# windows
task package-windows
# linux
task package-linux
# macos
task package-macos
```

