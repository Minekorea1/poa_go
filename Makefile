# uncomment if for windows
# SHELL=C:/Program Files/Git/usr/bin/bash.exe
POA=poa
WATCHER=watcher

$(info $(SHELL))

all: windows linux

windows: export GOOS=windows
windows: export GOARCH=amd64
windows:
#	fyne-cross windows -arch=amd64 -output $(POA)_${GOOS}_${GOARCH}.exe
#	fyne-cross windows -arch=amd64 -console -output $(POA)_${GOOS}_${GOARCH}.exe
#	go build -o bin/$(POA)_${GOOS}_${GOARCH}.exe -ldflags '-H windowsgui' $(POA)
	go build -o bin/$(POA)_${GOOS}_${GOARCH}.exe $(POA)
	go build -o bin/$(WATCHER)_${GOOS}_${GOARCH}.exe $(WATCHER)

linux: export GOOS=linux
linux: export GOARCH=amd64
# linux: export CGO_ENABLED=1
linux:
#	fyne-cross linux -arch=amd64 -output ../$(POA)_${GOOS}_${GOARCH}
	go build -o bin/$(POA)_${GOOS}_${GOARCH} $(POA)
	go build -o bin/$(WATCHER)_${GOOS}_${GOARCH} $(WATCHER)

# install:
# 	cp fyne-cross/bin/windows-amd64/* bin
# 	cp fyne-cross/bin/linux-amd64/* bin

gen_grpc:
	protoc -I watcher/. watcher/grpc/protos/* --go_out=watcher/. --go-grpc_out=watcher/.