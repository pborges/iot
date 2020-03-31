package main

import (
	"github.com/pborges/iot/espiot"
	"log"
	"os"
)

func main() {
	d := espiot.Device{
		AlwaysReconnect: true,
		Address:         "192.168.1.155",
		Log:             log.New(os.Stdout, "", log.LstdFlags),
	}

	if err := d.Connect(); err != nil {
		d.Log.Println("error connecting:", err)
	}

	d.OnConnect(func() {
		d.SetOnDisconnect("led.0", false)
		d.Set("led.0", true)
	})
	d.OnDisconnect(func() {
		d.Log.Println("lol im ded")
	})

	select {}
}
