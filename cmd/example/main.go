package main

import (
	"fmt"
	"github.com/pborges/iot/espiot"
	"github.com/robfig/cron/v3"
	logger "log"
	"os"
	"strings"
	"time"
)

func StaticDiscovery(ip ...string) (res []*espiot.Device, err error) {
	for _, addr := range ip {
		dev := &espiot.Device{}
		if err := dev.Connect(addr); err != nil {
			return nil, err
		}
		res = append(res, dev)
	}
	return res, nil
}

func main() {
	log := logger.New(os.Stdout, "", logger.LstdFlags)

	discovered, err := espiot.Discover("192.168.1.0/24")
	if err != nil {
		log.Println(err)
	}

	devs := make(map[string]*espiot.Device)
	for _, dev := range discovered {
		log.Println("Discovered", dev.String(), dev.ControlAddress.String())
		dev.Log = logger.New(os.Stdout, fmt.Sprintf("[%s] ", dev.String()), logger.LstdFlags|logger.Lshortfile)
		if !strings.HasPrefix(dev.Model, "s31") {
			continue
		}
		if dev.Name != "" {
			devs[dev.Name] = dev
		} else {
			devs[dev.Id] = dev
		}
	}

	for _, dev := range devs {
		dev.OnUpdate(func(v espiot.AttributeAndValue) {
			dev.Log.Println(v.AttributeDef().Name, v.InspectValue())
		})
		dev.OnConnect(func() {
			if err := dev.SetBool("gpio.0", true); err != nil {
				dev.Log.Println("error setting OnConnect:", err)
			}
			if err := dev.SetBoolOnDisconnect("gpio.0", false); err != nil {
				dev.Log.Println("error setting OnDisconnect:", err)
			}
		})

		dev.OnDisconnect(func() {
			for {
				time.Sleep(5 * time.Second)
				if err := dev.Reconnect(); err != nil {
					dev.Log.Println("error reconnecting", err)
				} else {
					return
				}
			}
		})
	}

	c := cron.New(
		cron.WithSeconds(),
		cron.WithLogger(cron.VerbosePrintfLogger(log)),
	)
	c.AddFunc("@every "+(1*time.Second).String(), func() {
		for _, dev := range devs {
			go func(dev *espiot.Device) {
				if err := dev.SetBool("led.0", !dev.GetBool("led.0")); err != nil {
					dev.Log.Println("error setting state:", err)
				}
			}(dev)
		}
	})

	c.Start()

	select {}
}
