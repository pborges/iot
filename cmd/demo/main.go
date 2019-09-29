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

	c2.Subscribe("sub1", ">", func(name string, value pubsub.Datum, ctx pubsub.Context) error {
		return nil
	})
	c2.Subscribe("sub2", ">", func(name string, value pubsub.Datum, ctx pubsub.Context) error {
		return nil
	})

	c3.Subscribe("sub1", ">", func(name string, value pubsub.Datum, ctx pubsub.Context) error {
		return nil
	})

	c4.Subscribe("sub1", "*.bob", func(name string, value pubsub.Datum, ctx pubsub.Context) error {
		fmt.Println("\nDEBUG publish from subscription")
		ctx.Publish("client1.temp", 99)
		return nil
	})

	fmt.Println("DEBUG create attribute")
	attr, _, _ := c1.CreateAttribute("temp", pubsub.IntegerDefinition{Default: 3})
	fmt.Println("\nDEBUG self update")
	attr.Update(55)

	fmt.Println("\nDEBUG create attribute")
	c4.CreateAttribute("bob", pubsub.IntegerDefinition{}, func(s pubsub.Source, i interface{}) error {
		if i.(int64) == 3 {
			return errors.New("threes are evil")
		}
		return nil
	})

	fmt.Println("\nDEBUG client publish")
	c4.Publish("client1.temp", 67)
	fmt.Println("\nDEBUG client publish")
	c2.Publish("client4.bob", 666)

	c5.Subscribe("sub1", "client4.*")

	for i := 0; i < 5; i++ {
		fmt.Println("\nDEBUG client publish")
		c2.Publish("client4.bob", i)
	}

	fmt.Println("done")
	fmt.Println()
	fmt.Println("list all keys")

	for _, d := range broker.List(">") {
		fmt.Println(d.Name, d.Value)
	}
}
