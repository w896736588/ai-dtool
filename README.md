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
```

首次使用建议安装 Wails CLI（用于桌面端调试/构建）：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 启动命令（go run）

`ConfigFile` 对应 `config/dtool/*.ini` 的文件名（不带 `.ini` 后缀）。

### Web 端（浏览器模式）

```bash
go run ./cmd/dtool --ConfigFile=company
```

默认访问地址：`http://localhost:17170/`（以配置中的 `run.ports` 为准）。

### 桌面端（Wails）

```bash
go run -tags production ./cmd/dtool_wails --ConfigFile=company
```

说明：
1. Wails 使用 `go run/go build` 时必须带 tag（建议 `production`）。
2. 如果不方便传 `--ConfigFile`，可用环境变量 `DTOOL_CONFIG_FILE` 指定。

## 编译命令

### 构建前端 dist

```bash
cd web
npm ci
npm run prod
```

### 构建 Web 模式后端 exe

```bash
go build -ldflags "-s -w" -o build/dtool.exe ./cmd/dtool
```

### 构建桌面端 exe

```bash
go build -tags production -ldflags "-s -w -H=windowsgui" -o build/dtool_wails.exe ./cmd/dtool_wails
```

## 配置项说明（`config/dtool/*.ini`）

### `[run]`

1. `host`：监听地址（可选，未配置时按默认行为）。
2. `ports`：服务端口列表，逗号分隔，例如 `17170,17171`。前端会从该列表中选端口请求 API。

### `[path]`

1. `webkit_driver_path`：webkit 驱动目录，支持占位符 `{DRIVE}`。
2. `webkit_data_path`：webkit 用户数据目录，支持占位符 `{DRIVE}`。
3. `webkit_download_path`：webkit 下载目录，支持占位符 `{DRIVE}`。

占位符说明：
1. `{DRIVE}`：优先 `D:`，若不存在则回退到 `C:`。

### `[base]`

1. `dbPath`：数据库目录；为空时默认 `config/dtool`。
2. `dbFileName`：数据库文件名；为空时默认 `dtool.db`（项目内逻辑为 `AppName.db`）。
3. `webPath`：前端 dist 目录绝对路径。

`webPath` 为空时默认使用当前项目 `web/dist`。

## 一键打包

在 `Windows PowerShell` 或 `CMD` 中，先切到项目根目录再执行：

```bat
script\build.bat
```

脚本会自动执行：
1. 构建 `web/dist`。
2. 构建 `dtool.exe`（Web 模式）和 `dtool_wails.exe`（桌面模式）。
3. 复制运行所需目录（`config/dtool`、`web/dist`、`internal/pkg/p_js`、数据库升级 SQL）。
4. 输出 `build/dtool_release_时间戳.zip`。

## 相关约定

开发约定见 [AGENTS.md](AGENTS.md)。
