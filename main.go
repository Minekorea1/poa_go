package main

import (
	"flag"
	"fmt"
	paoUpdater "pao/updater"
)

const (
	VERSION_NAME               = "v0.0.1"
	APPLICATION_UPDATE_ADDRESS = "github.com/Minekorea1/poa_go"
)

func main() {
	testFlag := false
	flag.BoolVar(&testFlag, "testFlag", false, "")

	versionFlag := false
	flag.BoolVar(&versionFlag, "version", false, "prints the version and exit")
	flag.Parse()

	if versionFlag {
		fmt.Println(VERSION_NAME)
		return
	}

	updater := paoUpdater.NewUpdater()
	updater.Init(VERSION_NAME, APPLICATION_UPDATE_ADDRESS)
	updater.Update()
}
