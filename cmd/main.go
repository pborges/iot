package main

import (
	"fmt"
	"github.com/pborges/iot/pubsub"
)

// demo
func main() {
	var broker pubsub.Broker
	c1, _ := broker.CreateClient("client1")
	c2, _ := broker.CreateClient("client2")
	c3, _ := broker.CreateClient("client3")
	c4, _ := broker.CreateClient("client4")

	c2.Subscribe(">", func(name string, value pubsub.Datum, ctx pubsub.Context) error {
		return nil
	})
	c2.Subscribe(">", func(name string, value pubsub.Datum, ctx pubsub.Context) error {
		return nil
	})

	c3.Subscribe(">", func(name string, value pubsub.Datum, ctx pubsub.Context) error {
		return nil
	})

	c4.Subscribe("*.bob", func(name string, value pubsub.Datum, ctx pubsub.Context) error {
		ctx.Publish("client1.temp", 99)
		return nil
	})

	attr, _, _ := c1.CreateAttribute("temp", pubsub.IntegerDefinition{Default: 3}, nil)
	attr.Update(55)

	c4.CreateAttribute("bob", pubsub.IntegerDefinition{}, func(i interface{}) error {
		fmt.Println("[AcceptFN          ] ATTR: bob VALUE:", i)
		return nil
	})

	c4.Publish("client1.temp", 67)
	c2.Publish("client4.bob", 666)

	fmt.Println("done")
}
