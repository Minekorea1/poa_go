package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/mouuff/go-rocket-update/pkg/provider"
	"github.com/mouuff/go-rocket-update/pkg/updater"
)

const (
	VERSION_NAME = "v0.1.0"
)

func main() {
	u := &updater.Updater{
		Provider: &provider.Github{
			RepositoryURL: "github.com/Minekorea1/poa_go",
			ArchiveName:   "poa.zip",
		},
		ExecutableName: fmt.Sprintf("poa_%s_%s", runtime.GOOS, runtime.GOARCH),
		Version:        VERSION_NAME,
	}

	versionFlag := false
	flag.BoolVar(&versionFlag, "version", false, "prints the version and exit")
	flag.Parse()

	if versionFlag {
		fmt.Println(u.Version)
		return
	}

	fmt.Println(u.Version)

	if _, err := u.Update(); err != nil {
		log.Println(err)
	}
}
