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


```shell
#需要sqlite 开启
export CGO_ENABLED=1  
#zhima
#linux
#export GOARCH=amd64 
#export GOOS=linux 
#windows
export GOARCH=amd64 CC=x86_64-w64-mingw32-gcc   
export GOOS=windows
go build -ldflags "-X main.IsBuild=1" -o ./build/tool.exe ./cmd/tool/main.go
git add ./build/tool.exe
#git update-index --chmod=+x ./build/zhima.exe
git ls-files --stage ./build/tool.exe
```