开发工具集合：
双击start.bat启动

下一步：
增加ssh断线检测及重连机制
增加前端页面缓存，不要每次都刷新

注意：
如果编译遇到错误 那么修改包中的检测内容大小后再编译（我们的编译是32位的）

SSH：
cliConf := base.ClientConfig{}
cliConf.CreateClient("121.40.109.241", 22, "frog", "frog987^%$321_220")
//多条命令用;分割
fmt.Println(cliConf.RunShell("ls -l"))


静默浏览器打开 环境安装
https://github.com/playwright-community/playwright-go
```go
##安装扩展
go get -u github.com/playwright-community/playwright-go   
##安装浏览器核心 通过代码执行（可以设置一个lock文件来判断是否安装）
err := playwright.Install()
```



```shell
export CGO_ENABLED=1  
export GOARCH=amd64   
export GOOS=windows
go build -ldflags "-X main.IsBuild=1 -X main.DbPath=D:/go/redis_manager/config/.db/" -o ./build/zhima.exe ./cmd/zhima/main.go
#git add ./build/zhima.exe
#git update-index --chmod=+x ./build/zhima.exe
git ls-files --stage ./build/zhima.exe
go build -ldflags "-X main.IsBuild=1 -X main.DbPath=" -o ./build/zhimaPub.exe ./cmd/zhima/main.go
git ls-files --stage ./build/zhimaPub.exe
```