package espiot

import (
	"context"
	"github.com/Ullaakut/nmap"
	"time"
)

func Discover(addrs string) ([]*Device, error) {
	var devs []*Device
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	scanner, err := nmap.NewScanner(
		nmap.WithTargets(addrs),
		nmap.WithPorts("5000,5001"),
		nmap.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	result, _, err := scanner.Run()
	if err != nil {
		return nil, err
	}

	// Use the results to print an example output
	for _, host := range result.Hosts {
		valid := true
		for _, port := range host.Ports {
			if port.State.State == "closed" {
				valid = false
			}
		}

		if valid {
			dev := &Device{}
			if err := dev.Connect(host.Addresses[0].Addr); err == nil {
				devs = append(devs, dev)
			}
		}
	}

	return devs, nil
}
