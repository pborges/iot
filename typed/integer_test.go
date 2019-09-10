package typed

import (
	"github.com/pborges/iot/pubsub"
	"testing"
)

func TestIntegerBroker_Create(t *testing.T) {
	b := &pubsub.BasicBroker{}

	_, err := b.Create("test", IntegerAcceptor{Min: 4, Max: 8}.Accept)
	if err != nil {
		t.Error(err)
	}

	err, _ = b.Publish("test", 6)
	if err != nil {
		t.Error(err)
	}

	err, _ = b.Publish("test", "123")
	if err == nil {
		t.Error("expected error, didnt get one")
	}

	err, _ = b.Publish("test", 1)
	if err == nil {
		t.Error("expected error, didnt get one")
	}

	err, _ = b.Publish("test", 100)
	if err == nil {
		t.Error("expected error, didnt get one")
	}
}
