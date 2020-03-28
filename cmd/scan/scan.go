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
		log.Println("Discovered", dev.ControlAddress.String(), dev.String())
	}
}
