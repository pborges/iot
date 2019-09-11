package main

import (
	"fmt"
	"github.com/pborges/iot"
)

var broker iot.Broker
var clientOrd int

func getTestClient() *iot.Client {
	clientOrd++
	client, err := broker.CreateClient(fmt.Sprintf("client%d", clientOrd))
	if err != nil {
		panic(err)
	}
	return client
}

// demo
func main() {
	c1, _ := broker.CreateClient("client1")
	c2, _ := broker.CreateClient("client2")
	c3, _ := broker.CreateClient("client3")
	c4, _ := broker.CreateClient("client4")

	c2.Subscribe(">", func(name string, value iot.Datum, res iot.BrokerAccess) error {
		return nil
	})

	c3.Subscribe(">", func(name string, value iot.Datum, res iot.BrokerAccess) error {
		return nil
	})

	c4.Subscribe("*.bob", func(name string, value iot.Datum, res iot.BrokerAccess) error {
		res.Publish("client1.temp", 99)
		return nil
	})

	attr, _, _ := c1.Create("temp", iot.IntegerDefinition{Default: 3}, nil)
	attr.Update(55)

	c4.Create("bob", iot.IntegerDefinition{}, nil)

	c4.Publish("client1.temp", 67)
	c2.Publish("client4.bob", 666)

	fmt.Println("done")
}
