package main

import (
	"github.com/pborges/iot/espiot"
	logger "log"
	"os"
)

func main() {
	log := logger.New(os.Stdout, "", logger.LstdFlags)

	discovered, err := espiot.Discover("192.168.1.0/24")
	if err != nil {
		log.Println(err)
	}

	for _, dev := range discovered {
		log.Printf("Discovered addr:'%s' id:'%s' fw:'%s' hw:'%s' model:'%s' name:'%s'\n", dev.Address, dev.Id(), dev.FrameworkVersion(), dev.HardwareVersion(), dev.Model(), dev.GetString("config.name"))
	}
}
