# uncomment if for windows
# SHELL=C:/Program Files/Git/usr/bin/bash.exe
MODULE=poa
BIN=poa

$(info $(SHELL))

all: windows linux


windows: export GOOS=windows
windows: export GOARCH=amd64
windows:
	go build -o bin/$(BIN)_${GOOS}_${GOARCH}.exe -ldflags '-H windowsgui' $(MODULE)

linux: export GOOS=linux
linux: export GOARCH=amd64
linux:
	go build -o bin/$(BIN)_${GOOS}_${GOARCH} $(MODULE)
