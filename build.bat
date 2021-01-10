SET GOOS=linux
go build -o bin/httpproxy cmd/main.go

SET GOOS=windows
go build -o bin/httpproxy.exe cmd/main.go
