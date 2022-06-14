export GOARCH=amd64
export GOOS=windows
go build -o bin/derosphere-windows.exe
export GOOS=linux
go build -o bin/derosphere-linux
export GOOS=darwin
go build -o bin/derosphere-darwin
