package main

import (
	"fmt"
	"github.com/pborges/iot/espiot"
	"log"
)

func main() {
	devs, err := espiot.Discover("192.168.1.0/24")
	if err != nil {
		log.Println(err)
	}

	for _, dev := range devs {
		fmt.Println(dev.ControlAddress.String(), dev.String())
		//fmt.Println("  Attrs----------------")
		//for _, a := range dev.ListAttributes() {
		//	fmt.Printf("    %-15s: %s\n", a.AttributeDef().Name, a.InspectValue())
		//}
		//fmt.Println("  Funcs----------------")
		//for _, fn := range dev.ListFunctions() {
		//	fmt.Printf("    %s\n", fn.Name)
		//	for _, a := range fn.Args {
		//		fmt.Printf("      %s -> %s\n", a.Name, a.Type)
		//	}
		//}
	}
}
