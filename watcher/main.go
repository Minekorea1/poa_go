package main

import (
	"poa/log"
	"time"

	"watcher/watcher"
)

const GRPC_PORT = "8889"

var logger = log.NewLogger("watcher main")

func main() {
	logger.LogI("start watcher")

	go watcher.RunServer()

	for {
		watcher.RunClient()

		time.Sleep(time.Second * 10)
	}
}
