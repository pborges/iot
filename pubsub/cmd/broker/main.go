package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/pborges/iot/espiot"
	"github.com/pborges/iot/pubsub"
	"log"
	"net"
	"os"
	"time"
)

func discoverAndHandle(broker *pubsub.Broker) {
	log.SetOutput(os.Stdout)
	log.SetPrefix("[main      ]")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	discovered, err := espiot.Discover("192.168.1.0/24")
	if err != nil {
		log.Println(err)
	}

	devs := make(map[string]*espiot.Device)
	for _, dev := range discovered {
		dev := dev
		log.Printf("Discovered addr:'%s' id:'%s' fw:'%s' hw:'%s' model:'%s' name:'%s'\n", dev.Address, dev.Id(), dev.FrameworkVersion(), dev.HardwareVersion(), dev.Model(), dev.GetString("config.name"))
		dev.AlwaysReconnect = true
		name := dev.Id()
		dev.Log = log.New(os.Stdout, fmt.Sprintf("[%s]", name), log.LstdFlags)

		if dev.Name() == "test" {
			devs[dev.Id()] = dev
		}
	}

	for k := range devs {
		dev := devs[k]
		node := pubsub.BasicNode{
			ID: dev.Id(),
		}
		for _, a := range dev.ListAttributes() {
			var attr pubsub.Attribute
			attr.Name = a.AttributeDef().Name
			switch a.(type) {
			case *espiot.StringAttributeValue:
				attr.Definition = pubsub.StringDefinition{AcceptFn: func(v string) error {
					return dev.SetString(attr.Name, v)
				}}
			case *espiot.IntegerAttributeValue:
				attr.Definition = pubsub.IntegerDefinition{AcceptFn: func(v int64) error {
					return dev.SetInteger(attr.Name, int(v))
				}}
			case *espiot.DoubleAttributeValue:
				attr.Definition = pubsub.DoubleDefinition{AcceptFn: func(v float64) error {
					return dev.SetDouble(attr.Name, v)
				}}
			case *espiot.BooleanAttributeValue:
				attr.Definition = pubsub.BooleanDefinition{AcceptFn: func(v bool) error {
					return dev.SetBool(attr.Name, v)
				}}
			}
			node.Attributes = append(node.Attributes, attr)
		}

		dev.OnUpdate(func(v espiot.AttributeAndValue) {
			id := fmt.Sprintf("%s.%s", dev.Id(), v.AttributeDef().Name)
			if err := broker.Publish(node, id, v.InspectValue()); err != nil {
				log.Println("error publishing", id, v.InspectValue(), err)
			}
		})

		if err := broker.Register(node); err != nil {
			log.Println("error registering node", node.ID, err)
		}

		dev.Connect()
	}
}

func main() {
	broker := &pubsub.Broker{
		Log: log.New(os.Stdout, "[BROKER] ", log.LstdFlags),
	}

	go discoverAndHandle(broker)

	ln, err := net.Listen("tcp", ":5000")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go handleConnection(broker, conn)
	}
}

func handleConnection(broker *pubsub.Broker, conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	in := make(chan string)
	go process(broker, conn, in)
	for scanner.Scan() {
		in <- scanner.Text()
	}
}

func process(broker *pubsub.Broker, conn net.Conn, in chan string) {
	fmt.Fprint(conn, "node: ")

	node := pubsub.BasicNode{
		ID: <-in,
	}

	fmt.Fprintf(conn, "Welcome %s!\n", node.ID)
	acceptFn := make(chan string)
	for {
		select {
		case value := <-acceptFn:
			fmt.Fprint(conn, "PUB value:'"+value+"' Press enter to accept, or type an error: ")
			acceptFn <- <-in
			fmt.Fprintln(conn, "ok")
		case line := <-in:
			packet, err := espiot.Decode(line)
			if err != nil {
				fmt.Fprintln(conn, err)
				continue
			}
			switch packet.Command {
			case "list":
				for _, a := range broker.Values(">", time.Now()) {
					fmt.Fprintf(conn, "%s: %s\n", a.AttributeID, a.Value.Inspect())
				}
				fmt.Fprintln(conn, "ok")
			case "pub":
				if err := broker.Publish(node, packet.Args["name"], packet.Args["value"]); err == nil {
					fmt.Fprintln(conn, "ok")
				} else {
					fmt.Fprintln(conn, "err", err)
				}
			case "sub":
				var sub pubsub.Subscription
				sub.Name = packet.Args["name"]
				sub.Filter = packet.Args["filter"]
				sub.Fn = func(ctx pubsub.Context, v pubsub.Value) {
					fmt.Fprintf(conn, "SUB[%s] attribute: %s value: %s published by %s @ %s\n",
						sub.Name,
						v.AttributeID,
						v.Inspect(),
						v.UpdatedBy,
						v.UpdatedAt.Local().Format(time.RubyDate),
					)
				}
				node.Subscriptions = append(node.Subscriptions, sub)
				broker.Register(node)
				fmt.Fprintln(conn, "ok")
			case "def":
				var attr pubsub.Attribute
				attr.Name = packet.Args["name"]
				switch packet.Args["type"] {
				case "string":
					def := pubsub.StringDefinition{}
					def.AcceptFn = func(v string) error {
						acceptFn <- v
						err := <-acceptFn
						if err != "" {
							return errors.New(err)
						}
						return nil
					}
					attr.Definition = def
				default:
					fmt.Fprintln(conn, "unknown type")
					continue
				}
				node.Attributes = append(node.Attributes, attr)
				broker.Register(node)
				fmt.Fprintln(conn, "ok")
			default:
				fmt.Fprintln(conn, "unknown command")
			}
		}
	}
}
