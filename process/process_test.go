package process

import (
	"github.com/pborges/iot/pubsub"
	"testing"
)

func TestProcess_Subscribe(t *testing.T) {
	p := Process{broker: &pubsub.CoreBroker{}}
	p.Subscribe("*", nil)
	if len(p.subscriptions) != 1 || p.subscriptions[0].Filter() != "*" {
		t.Fail()
	}
}
