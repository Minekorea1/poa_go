package autostart

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

var template string = `
[Unit]
Description=PoA Unit Service
After=network.target

[Service]
Type=forking

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

	file, err := os.Create(service.path)
	if err != nil {
		log.Println(err)
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
