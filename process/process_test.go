package process

import (
	"github.com/pborges/iot/pubsub"
	"testing"
)

func TestProcess_Subscribe(t *testing.T) {
	broker := &pubsub.CoreBroker{}
	p := Process{broker: broker}
	p.Subscribe("*", nil)
	if len(p.subscriptions) != 1 {
		t.Error("subscription count is incorrect")
	}
	if p.subscriptions[0].Filter() != "*" {
		t.Error("subscription filter is incorrect")
	}
}

func TestProcess_Publish(t *testing.T) {
	broker := &pubsub.CoreBroker{}
	_, err := broker.Create("test", nil)
	if err != nil {
		t.Error(err)
	}

	p := Process{Name: "Process One", broker: broker}
	p.Subscribe("*", func(m Message) error {
		if m.Value != 1234 {
			t.Error("unexpected value")
		}
		if m.Process != p.Name {
			t.Error("invalid process name")
		}
		return nil
	})

	err, _ = p.Publish("test", 1234)
	if err != nil {
		t.Error(err)
	}
}
