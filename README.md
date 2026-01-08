### 新环境安装
go env -w GOPROXY=https://goproxy.cn,direct

### 更新计划
1. 数据库变更：文件支持按年按月目录分类执行 ok
2. 接口开发：执行后保存最后一次的结果 ok
3. 接口开发：执行时需要等待保存完后再执行，且自动切换到结果 ok
4. 接口开发：接口详情页面，右上角环境变量支持查看 ok
5. 接口开发：支持文件夹详情中显示所有接口文档，支持一键复制 ok
6. 接口开发：支持从curl命令生成接口 ok
7. 输出监控：修复某些情况下丢失数据问题（可能是缓冲区不够，还没到换行） ok
8. 输出监控：优化拦截展示页面，把拦截的标题和次数用表格更清晰展示 ok
9. docker：docker服务列表固定按自然顺序排序 ok
10. 自定义网页：执行逻辑支持复制新增 ok
11. nginx：增加nginx配置文件一览 wait

### go run 启动
ConfigFile设置为config/dtool下面的某个文件
```shell
export CGO_ENABLED=1
go run -ldflags " -s -w" cmd/dtool/main.go --ConfigFile=company
```

### bat启动
将下面的内容保存为xxx.bat放到build目录即可双击运行，ctrl + c结束
```bat
@echo off
REM 打开网页（异步）
start "" "http://localhost:17170/"

REM 同步运行 dtool，保持在当前控制台
dtool.exe --ConfigFile=xxx

REM 可选：运行结束后暂停（方便看最后日志）
pause
```
