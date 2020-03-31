package espiot

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Ullaakut/nmap"
	"net"
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
			if dev, err := validateDevice(host.Addresses[0].Addr); err == nil {
				devs = append(devs, dev)
			}
		}
	}

	return devs, nil
}

func validateDevice(addr string) (*Device, error) {
	dev := &Device{
		Address: addr,
	}
	conn, err := net.DialTimeout("tcp", addr+":5000", 3*time.Second)
	if err != nil {
		return nil, err
	}

	if err = conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintln(conn, Encode(Packet{Command: "info"})); err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(conn)

	if err = conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return nil, err
	}
	if res, err := readResponse(scanner); err == nil {
		if err := dev.setMetadata(res[0]); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	if _, err := fmt.Fprintln(conn, Encode(Packet{Command: "list"})); err != nil {
		return nil, err
	}

	if err = conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		return nil, err
	}
	if res, err := readResponse(scanner); err == nil {
		if err := dev.handleList(res); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return dev, nil
}

func readResponse(scanner *bufio.Scanner) ([]Packet, error) {
	var listPackets []Packet

	for scanner.Scan() {
		if scanner.Text() != "ok" {
			if p, err := Decode(scanner.Text()); err == nil {
				listPackets = append(listPackets, p)
			} else {
				return nil, err
			}
		} else {
			break
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return listPackets, nil
}
