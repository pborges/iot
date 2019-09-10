package process

import (
	"github.com/pborges/iot/pubsub"
	"testing"
)

type TestBroker struct {
	pubsub.Broker
}
type TestPublication struct {
	Name string
	pubsub.Publication
}

func (p TestPublication) Key() string {
	return p.Name
}

func (b TestBroker) Create(key string, fn pubsub.AcceptFn) (pubsub.Publication, error) {
	return TestPublication{Name: key}, nil
}

func TestProcess_CreateAddsPrefix(t *testing.T) {
	proc := Process{name: "process", Broker: TestBroker{}}

	pub, err := proc.Create("test", nil)
	if err != nil {
		t.Error(err)
	}

	if pub.Key() != "process.test" {
		t.Error("invalid name")
	}
}
