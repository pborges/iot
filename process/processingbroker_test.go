package process

import (
	"github.com/pborges/iot/pubsub"
	"testing"
)

func TestProcess_Subscribe(t *testing.T) {
	p := ProcessingBroker{}
	p.Wrap(&pubsub.Broker{})

	p.Subscribe("*", nil)
	if len(p.subscriptions) != 1 {
		t.Error("subscription count is incorrect")
	}
	if p.subscriptions[0].Filter() != "*" {
		t.Error("subscription filter is incorrect")
	}
}

func TestProcess_Publish(t *testing.T) {
	broker := &pubsub.Broker{}

	p1 := ProcessingBroker{name: "ProcessingBroker 1", broker: broker}
	p1.Subscribe("*", func(m Message) error {
		if m.Process != p1.name {
			t.Error("unexpected process name")
		}
		if m.Value != 1234 {
			t.Error("unexpected value")
		}
		if m.Process != p1.name {
			t.Error("invalid process name")
		}
		if m.Key != "test" {
			t.Error("invalid key name")
		}
		return nil
	})

	p2 := ProcessingBroker{name: "ProcessingBroker 2", broker: broker}

	pub, _ := p2.Create("test", nil)
	pub.Update(1234)
}
