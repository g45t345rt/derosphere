FLAGS="-X main.DEBUG_MODE=false"
export GOARCH=amd64
export GOOS=windows
go build -o bin/derosphere-windows.exe -ldflags "$FLAGS"
export GOOS=linux
go build -o bin/derosphere-linux -ldflags "$FLAGS"
export GOOS=darwin
go build -o bin/derosphere-darwin -ldflags "$FLAGS"
