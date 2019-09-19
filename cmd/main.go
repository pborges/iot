package main

import (
	"errors"
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
	c5, _ := broker.CreateClient("client5")

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

	c4.CreateAttribute("bob", pubsub.IntegerDefinition{}, func(s pubsub.Source, i interface{}) error {
		if i.(int64) == 3 {
			return errors.New("threes are evil")
		}
		return nil
	})

	c4.Publish("client1.temp", 67)
	c2.Publish("client4.bob", 666)

	c5.Subscribe("client4.*", nil)

	for i := 0; i < 5; i++ {
		c2.Publish("client4.bob", i)
	}

	fmt.Println("done")
}
