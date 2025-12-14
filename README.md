### 启动
双击start.bat启动

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
8. 输出监控：优化拦截展示页面，把拦截的标题和次数用表格更清晰展示 0k
9. docker服务列表固定按自然顺序排序 wait

```shell
#公司编译
export CGO_ENABLED=1  
export GOARCH=amd64   
export GOOS=windows
# DbPath数据库目录
# DbName数据库文件名 为空的话取服务名
go build -tags timetzdata -ldflags " -s -w" -o ./build/dtool.exe ./cmd/dtool/main.go
#git add ./build/dtool.exe
#git update-index --chmod=+x ./build/dtool.exe
git ls-files --stage ./build/dtool.exe
go build -tags timetzdata -ldflags " -s -w" -o ./build/zhimaPub.exe ./cmd/dtool/main.go
git ls-files --stage ./build/zhimaPub.exe
```

```shell
export CGO_ENABLED=1
go run -ldflags " -s -w" cmd/dtool/main.go --ConfigFile=zkzf_home
```
