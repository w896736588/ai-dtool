开发工具集合：
双击start.bat启动

注意：
如果编译遇到错误 那么修改包中的检测内容大小后再编译（我们的编译是32位的）

SSH：
cliConf := base.ClientConfig{}
cliConf.CreateClient("121.40.109.241", 22, "frog", "frog987^%$321_220")
//多条命令用;分割
fmt.Println(cliConf.RunShell("ls -l"))


```shell
export CGO_ENABLED=0  
#linux
#export GOARCH=amd64 
#export GOOS=linux 
#windows
export GOARCH=amd64 CC=x86_64-w64-mingw32-gcc   
export GOOS=windows
go build -o ./build/zhima.exe ./cmd/zhima/main.go
git add ./build/zhima.exe
git update-index --chmod=+x ./build/zhima.exe
git ls-files --stage ./build/zhima.exe
```