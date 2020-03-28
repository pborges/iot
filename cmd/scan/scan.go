package main

import (
	"fmt"
	"github.com/pborges/iot/espiot"
	"github.com/robfig/cron"
	logger "log"
	"os"
	"strings"
	"time"
)

func main() {
	log := logger.New(os.Stdout, "", logger.LstdFlags)

	//d := espiot.Device{}
	//d.OnConnect(func() {
	//	d.SetBool("led.0", true)
	//	d.SetBoolOnDisconnect("led.0", false)
	//})
	//d.OnUpdate(func(v espiot.AttributeAndValue) {
	//	fmt.Println(v.AttributeDef().Name, v.InspectValue())
	//})
	//if err := d.Connect("192.168.1.155"); err != nil {
	//	panic(err)
	//}
	//select {}

	discovered, err := espiot.Discover("192.168.1.0/24")
	if err != nil {
		log.Println(err)
	}

	devs := make(map[string]*espiot.Device)
	for _, dev := range discovered {
		log.Println("Discovered", dev.String(), dev.ControlAddress.String())
		if strings.HasPrefix(dev.Model, "s31") {
			dev.Log = logger.New(os.Stdout, fmt.Sprintf("[%s] ", dev.String()), logger.LstdFlags|logger.Lshortfile)
			devs[dev.Id] = dev
		}
	}

	for _, dev := range devs {
		dev.OnUpdate(func(v espiot.AttributeAndValue) {
			dev.Log.Println(v.AttributeDef().Name, v.InspectValue())
		})
		dev.OnConnect(func() {
			if err := dev.SetBool("led.0", true); err != nil {
				dev.Log.Println("error setting OnConnect:", err)
			}
			if err := dev.SetBoolOnDisconnect("led.0", false); err != nil {
				dev.Log.Println("error setting OnDisconnect:", err)
			}
		})

		dev.OnDisconnect(func() {
			for {
				if err := dev.Reconnect(); err != nil {
					dev.Log.Println("error reconnecting", err)
					time.Sleep(5 * time.Second)
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
	b := false
	c.AddFunc("@every "+(1*time.Second).String(), func() {
		for _, dev := range devs {
			if err := dev.SetBool("gpio.0", b); err != nil {
				dev.Log.Println("error setting state:", err)
			}
		}
		b = !b
	})

	c.Start()

	select {}
}
