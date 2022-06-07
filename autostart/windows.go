package autostart

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PowerShell struct {
	exec string
}

var WIN_CREATE_SHORTCUT = `$WshShell = New-Object -comObject WScript.Shell
						   $Shortcut = $WshShell.CreateShortcut("$HOME\AppData\Roaming\Microsoft\Windows\Start Menu\Programs\Startup\LINK_NAME.lnk")
						   $Shortcut.TargetPath = "TARGET_PATH"
						   $Shortcut.WorkingDirectory = "WORKING_DIRECTORY"
						   $Shortcut.Save()`

func newPowerShell() *PowerShell {
	ps, _ := exec.LookPath("powershell.exe")
	return &PowerShell{
		exec: ps,
	}
}

func (powerShell *PowerShell) execute(args ...string) (stdOut string, stdErr string, err error) {
	args = append([]string{"-NoProfile", "-NonInteractive"}, args...)
	cmd := exec.Command(powerShell.exec, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	stdOut, stdErr = stdout.String(), stderr.String()
	return
}

func (powerShell *PowerShell) registerAutoStart() {
	ps := newPowerShell()
	exec_path := os.Args[0]
	WIN_CREATE_SHORTCUT = strings.Replace(WIN_CREATE_SHORTCUT, "LINK_NAME", "PoA", 1)
	WIN_CREATE_SHORTCUT = strings.Replace(WIN_CREATE_SHORTCUT, "TARGET_PATH", exec_path, 1)
	WIN_CREATE_SHORTCUT = strings.Replace(WIN_CREATE_SHORTCUT, "WORKING_DIRECTORY", filepath.Dir(exec_path), 1)
	_, _, err := ps.execute(WIN_CREATE_SHORTCUT)
	if err != nil {
		log.Println(err)
	}
}

func (powerShell *PowerShell) removeOldShortCut() {
	home, _ := os.UserHomeDir()
	shortCutPath := home + "\\AppData\\Roaming\\Microsoft\\Windows\\Start Menu\\Programs\\Startup\\poa_windows_amd64.exe - 바로 가기.lnk"
	if _, err := os.Stat(shortCutPath); err == nil {
		err = os.Remove(shortCutPath)
		if err != nil {
			log.Println(err)
		}
	}

	shortCutPath = home + "\\AppData\\Roaming\\Microsoft\\Windows\\Start Menu\\Programs\\Startup\\poa_windows_amd64 - 바로 가기.lnk"
	if _, err := os.Stat(shortCutPath); err == nil {
		err = os.Remove(shortCutPath)
		if err != nil {
			log.Println(err)
		}
	}
}

func (powerShell *PowerShell) removeShortCut() {
	home, _ := os.UserHomeDir()
	shortCutPath := home + "\\AppData\\Roaming\\Microsoft\\Windows\\Start Menu\\Programs\\Startup\\PoA.lnk"
	if _, err := os.Stat(shortCutPath); err == nil {
		err = os.Remove(shortCutPath)
		if err != nil {
			log.Println(err)
		}
	}
}
