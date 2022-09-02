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

	var lastAliveInfo *watcher.AliveInfo

	for {
		aliveInfo := watcher.RunClient()

		if lastAliveInfo != nil {
			if aliveInfo.MqttPublishTimestamp-lastAliveInfo.MqttPublishTimestamp == 0 {
				logger.LogW("PoA is not responding. restart PoA.")
				watcher.StartProcess(aliveInfo.GetPoaArgs()...)
			}
		}

		lastAliveInfo = aliveInfo

		time.Sleep(time.Minute * 5)
	}
}
