

.PHONY:zhima_tool_widows
zhima_tool_widows:
	go env -w CGO_ENABLED=1 GOOS=windows GOARCH=amd64
	go mod tidy
	go build -tags timetzdata -ldflags "-X main.IsBuild=1 -X main.DbPath=D:/go/cache_manager_api/config/zhima/ -X main.ViewPath=D:/go/devtool/dist -s -w" -o ./build/zhima.exe ./cmd/zhima/main.go
	git ls-files --stage build/zhima.exe

.PHONE:zhima_tool_pub_widows
zhima_tool_pub_widows:
	go env -w CGO_ENABLED=1 GOOS=windows GOARCH=amd64
	go mod tidy
	go build -tags timetzdata -ldflags "-X main.IsBuild=1 -X main.DbPath= -X main.ViewPath= -s -w" -o ./build/zhima.exe ./cmd/zhima/main.go
	git ls-files --stage build/zhimaPub.exe



.PHONY:make_all
make_all:
	make zhima_tool_widows
	make zhima_tool_pub_widows
