package autostart

import (
	"fmt"
	"runtime"
)

type shortcutCreator interface {
	registerAutoStart()
}

type AutoStart struct {
	creator shortcutCreator
}

func NewAutoStart() *AutoStart {
	autoStart := &AutoStart{}

	os := runtime.GOOS
	switch os {
	case "windows":
		autoStart.creator = newPowerShell()
	case "darwin":
	case "linux":
		autoStart.creator = newService()
		newService().createServiceFile()
	default:
		fmt.Printf("%s.\n", os)
	}

	return autoStart
}

func (autoStart *AutoStart) CreateShortcut() {
	os := runtime.GOOS
	switch os {
	case "windows":
		autoStart.creator.registerAutoStart()
	case "darwin":
	case "linux":
		newService().createServiceFile()
	default:
		fmt.Printf("%s.\n", os)
	}
}

func (autoStart *AutoStart) DeleteShortcut() {
	os := runtime.GOOS
	switch os {
	case "windows":
		autoStart.creator.(*PowerShell).removeOldShortCut()
		autoStart.creator.(*PowerShell).removeShortCut()
	case "darwin":
	case "linux":
	default:
		fmt.Printf("%s.\n", os)
	}
}
