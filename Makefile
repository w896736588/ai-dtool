.PHONY:windows
windows:
	set CGO_ENABLED=0;GOOS=windows;GOARCH=amd64 CC=x86_64-w64-mingw32-gcc&&go build -o main.exe main.go
.PHONY:linux
linux:
	set GOOS=linux&&go build -o cache_managerr -ldflags "-s -w"
