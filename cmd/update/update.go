package main

import (
	"bufio"
	"github.com/pborges/iot/espiot"
	logger "log"
	"os"
	"os/exec"
	"os/user"
)

func main() {
	log := logger.New(os.Stdout, "", logger.LstdFlags)

	discovered, err := espiot.Discover("192.168.1.0/24")
	if err != nil {
		log.Println(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for _, dev := range discovered {
		log.Println("Discovered", dev.ControlAddress.String(), dev.String())
		usr, _ := user.Current()
		dir := usr.HomeDir + "/src/iot/espiot_" + dev.Model
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			log.Println("Found firmware directory", dir)
			log.Println("Do update? [y/n]")
			if scanner.Scan() && scanner.Text() == "y" {
				log.Println("Updating...")
				updateCmd := exec.Command("pio", "run", "-t", "upload", "--upload-port", dev.IpAddr())
				updateCmd.Dir = dir
				updateCmd.Stdout = os.Stdout
				if err := updateCmd.Run(); err != nil {
					log.Println(err)
				}
			}
		} else {
			log.Println("Unable to locate firmware directory", dir)
		}
	}
}
