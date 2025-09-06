### 启动
双击start.bat启动

- 如果编译遇到错误 那么修改包中的检测内容大小后再编译（我们的编译是32位的）
- 如果发布后打开报错，那么打开一个cmd窗口，然后直接输入.exe完整目录执行

### 新环境安装
go env -w GOPROXY=https://goproxy.cn,direct

### 插件
- Batch Scripts support 允许在README.md中直接执行命令
- MarkDown Editor Markdown编辑器

### 各个master分支
- master负责修改代码，其他代码永远使用master的代码，master不保留数据库配置文件
- 其他master只保留有自己的编译文件，数据库文件

### 编译及运行设置的参数说明
- -X main.DbPath 本地数据库文件
- -X main.ViewPath 前端页面dist目录

```shell
#公司编译
export CGO_ENABLED=1  
export GOARCH=amd64   
export GOOS=windows
go build -tags timetzdata -ldflags "-X main.DbPath=D:/go/cache_manager_api/config/zhima/ -X main.ViewPath=D:/go/devtool/dist -s -w" -o ./build/zhima.exe ./cmd/zhima/main.go
#git add ./build/zhima.exe
#git update-index --chmod=+x ./build/zhima.exe
git ls-files --stage ./build/zhima.exe
go build -tags timetzdata -ldflags "-X main.DbPath= -X main.ViewPath= -s -w" -o ./build/zhimaPub.exe ./cmd/zhima/main.go
git ls-files --stage ./build/zhimaPub.exe
```

```shell
#公司go run
export CGO_ENABLED=1
go run -ldflags "-X main.DbPath=D:/go/cache_manager_api/config/zhima/ -X main.ViewPath=D:/go/devtool/dist" cmd/zhima/main.go
```

```shell
#家里go run
export CGO_ENABLED=1
go run -ldflags "-X main.DbPath=C:\work\frog\dev_tool_master\config\zhima\ -X main.ViewPath=C:\work\frog\cache_manager_web\dist" cmd/zhima/main.go
```
