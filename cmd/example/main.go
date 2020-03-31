package main

import (
	"fmt"
	"github.com/pborges/iot/espiot"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	crontab := cron.New(
		cron.WithSeconds(),
	)

	discovered, err := espiot.Discover("192.168.1.0/24")
	if err != nil {
		log.Println(err)
	}

	devices := make(map[string]*espiot.Device)
	for _, dev := range discovered {
		log.Printf("Discovered addr:'%s' id:'%s' fw:'%s' hw:'%s' model:'%s' name:'%s'\n", dev.Address, dev.Id(), dev.FrameworkVersion(), dev.HardwareVersion(), dev.Model(), dev.GetString("config.name"))
		dev.AlwaysReconnect = true
		name := dev.Id()
		if dev.GetString("config.name") != "" {
			name = dev.GetString("config.name")
		}
		dev.Log = log.New(os.Stdout, fmt.Sprintf("[%s]", name), log.LstdFlags)

		dev.OnConnect(func() {
			dev.SetBoolOnDisconnect("gpio.0", false)
			dev.SetBool("gpio.0", true)
		})

		if strings.HasPrefix(dev.Model(), "s31") {
			devices[name] = dev
		}
	}

	for k := range devices {
		devices[k].Connect()
		if _, err := crontab.AddFunc("@every "+(1*time.Second).String(), func() {
			if err := devices[k].SetBool("led.0", !devices[k].GetBool("led.0")); err != nil {
				fmt.Println(err)
			}
		}); err != nil {
			fmt.Println("error in cron", err)
		}
	}

	crontab.Start()
	select {}
}
