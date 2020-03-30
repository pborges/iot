package pubsub

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func dump(broker *Broker) {
	fmt.Println("--------------------------------------------------")
	for _, rec := range broker.Values(">", time.Now()) {
		fmt.Printf("ATTRIBUTE: %-5s VALUE: %s\n", rec.AttributeID, rec.Inspect())
		for _, res := range rec.SubscriptionResponses {
			fmt.Printf("  SUBSCRIPTION: %-15s ERRORS: %d\n", res.SubscriptionID, len(res.Err))
			for _, err := range res.Err {
				fmt.Printf("         ERROR: %s\n", err.Error())
			}
		}
	}
	fmt.Println("--------------------------------------------------")
	fmt.Println()
}

func TestBroker(t *testing.T) {
	n1 := BasicNode{
		ID: "n1",
		Attributes: []Attribute{
			{
				Name:       "a1",
				Definition: StringDefinition{},
			},
		},
		Subscriptions: []Subscription{
			{
				Name:   "monkey_see_monkey_do",
				Filter: "n2.a2",
				Fn: func(ctx Context, v Value) {
					ctx.Error(ctx.Publish("n1.a1", "copy of "+v.Value.(string)))
				},
			},
		},
	}

	n2 := BasicNode{
		ID: "n2",
		Attributes: []Attribute{
			{
				Name: "a1",
				Definition: StringDefinition{
					AcceptFn: func(v string) error {
						if v == "hello" {
							return errors.New("i dont like hellos")
						}
						return nil
					},
				},
			},
			{
				Name:       "a2",
				Definition: StringDefinition{},
			},
		},
		Subscriptions: []Subscription{
			{
				Name:   "catch_all",
				Filter: ">",
				Fn: func(ctx Context, v Value) {
					if v.Inspect() == "h2" {
						ctx.Error(errors.New("idk what to do with this information"))
					}
					ctx.Error(errors.New("test error"))
				},
			},
		},
	}
	broker := &Broker{
		Log: log.New(os.Stdout, "[BROKER] ", log.LstdFlags),
	}

	fmt.Println(`broker.Register(n)`)
	for _, n := range []Node{n1, n2} {
		broker.Register(n)
	}

	fmt.Println()
	dump(broker)

	fmt.Println(`broker.Publish(n1, "n2.a1", "h1")`)
	broker.Publish(n1, "n2.a1", "h1")
	fmt.Println()
	dump(broker)

	fmt.Println(`broker.Publish(n1, "n1.a1", "h2")`)
	broker.Publish(n1, "n1.a1", "h2")
	fmt.Println()
	dump(broker)

	fmt.Println(`broker.Publish(n2, "n2.a1", "h3")`)
	broker.Publish(n2, "n2.a1", "h3")
	fmt.Println()
	dump(broker)

	fmt.Println(`broker.Publish(n2, "n2.a1", "hello")`)
	broker.Publish(n2, "n2.a1", "hello")
	fmt.Println()
	dump(broker)

	fmt.Println(`broker.Publish(n2, "n2.a2", "gzorp")`)
	broker.Publish(n2, "n2.a2", "gzorp")
	fmt.Println()
	dump(broker)
}
