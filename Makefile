
.PHONY:dev_tool_widows
dev_tool_widows:
	go env -w CGO_ENABLED=1 GOOS=windows GOARCH=amd64
	go mod tidy
	go build -ldflags "-s -w" -o ./build/dtool.exe ./cmd/dtool/main.go
	git ls-files --stage build/dtool.exe


.PHONY:make_all
make_all:
	make dev_tool_widows
