# dev_tool_master

## 功能简介

本工具是面向开发与运维场景的个人使用的本地化工作台，当前仅保留 Web/API 运行模式，数据存储在个人空间。

### 模块说明

- 首页：通过快捷命令操作 Git 等模块
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
# air监听启动
go install github.com/air-verse/air@latest
```

## 开发时启动命令（task）

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
nohup ./dtool --ConfigFile=space >> /var/log/space.$(date +%Y%m%d).log 2>&1 &
```

## dtool-agent 构建命令

`dtool-agent` 支持编译时默认值和运行时环境变量两种配置方式：

- `DTOOL_SERVER_URL`：服务端地址，例如 `http://localhost:17170`
- `DTOOL_CLIENT_VERSION`：客户端版本号，需要和服务端配置的 `smart_link.client_version` 保持一致

优先级说明：

- 运行时环境变量优先
- 如果运行时未设置，则回退到编译时通过 `-ldflags -X` 注入的默认值
- 如果编译时也未注入，则回退到代码内默认值

### macOS 版本

```powershell
# 构建 macOS agent
$env:GOOS="darwin"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -ldflags "-X main.defaultServerURL=http://localhost:17170 -X main.defaultClientVersion=2.0.0" -o build/dtool-agent ./cmd/dtool-agent
```

### Windows 版本

```powershell
# 构建 Windows agent
$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CGO_ENABLED="0"
go build -ldflags "-X main.defaultServerURL=http://localhost:17170 -X main.defaultClientVersion=2.0.0" -o build/dtool-agent.exe ./cmd/dtool-agent
```

如果只想在运行时临时覆盖，也可以在启动前设置环境变量。

```powershell
# macOS
$env:DTOOL_SERVER_URL="http://localhost:17170"; $env:DTOOL_CLIENT_VERSION="2.0.0"; ./build/dtool-agent
```

```powershell
# Windows
$env:DTOOL_SERVER_URL="http://localhost:17170"; $env:DTOOL_CLIENT_VERSION="2.0.0"; .\build\dtool-agent.exe
```
