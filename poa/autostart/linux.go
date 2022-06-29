package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"poa/log"
	"strings"
)

var logger = log.NewLogger("autostart-linux")

var template string = `
[Unit]
Description=PoA Unit Service
After=network.target

[Service]
Environment="DISPLAY=:0"
Environment="XAUTHORITY=/home/user/.Xauthority"
User=user
Type=simple

ExecStart=TARGET_PATH
WorkingDirectory=WORKING_DIRECTORY

Restart=always
RestartSec=30
KillSignal=SIGINT

[Install]
WantedBy=multi-user.target
`

type Service struct {
	path string
}

func newService() *Service {
	return &Service{
		path: "/etc/systemd/system/poa.service",
	}
}

func (service *Service) createServiceFile() {
	appService := template
	// exec_path := os.Args[0]

	exec_path, err := os.Executable()
	if err != nil {
		logger.LogE(err)
	}

	appService = strings.Replace(appService, "TARGET_PATH", exec_path, 1)
	appService = strings.Replace(appService, "WORKING_DIRECTORY", filepath.Dir(exec_path), 1)

	file, err := os.Create(service.path)
	if err != nil {
		logger.LogE(err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "%s", appService)
}

func (service *Service) registerAutoStart() {
	exec.Command("systemctl", "daemon-reload")
	exec.Command("systemctl", "enable", "poa.service")
	exec.Command("systemctl", "start", "poa.service")
}
