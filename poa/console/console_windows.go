//go:build windows
// +build windows

package console

import (
	"fmt"
	"os"
	"syscall"
)

const attach_parent_process = ^uint32(0) // (DWORD)-1

var (
	modkernel32       = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole = modkernel32.NewProc("AttachConsole")
)

func attachConsole(dwParentProcess uint32) (ok bool) {
	r0, _, _ := syscall.Syscall(procAttachConsole.Addr(), 1, uintptr(dwParentProcess), 0, 0)
	ok = bool(r0 != 0)
	return
}

func AttachConsole() {
	ok := attachConsole(attach_parent_process)
	if ok {
		fmt.Println("Okay, attached")
	}
	hout, err1 := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err1 != nil {
		os.Exit(2)
	}
	os.Stdout = os.NewFile(uintptr(hout), "/dev/stdout")
	fmt.Println("console attach complete")
}
