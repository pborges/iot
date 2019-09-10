package definition

import (
	"errors"
	"fmt"
	"github.com/pborges/iot/pubsub"
	"testing"
)

func TestIntegerBroker_Create(t *testing.T) {
	b := &pubsub.BasicBroker{}

	intDef := IntegerDefinition{Min: 4, Max: 8}
	intDef.OnAccept(func(value interface{}) (err error) {
		if value == 7 {
			err = errors.New("it is impossible to send a 7 over the network")
		}
		fmt.Println("doNetworkStuffToSetValue(", value, ") ->", err)
		return err
	})

	intDef.OnAccept(func(value interface{}) error {
		fmt.Println("am i a logger?", value)
		return nil
	})

	_, err := b.Create("test", intDef.Accept)
	if err != nil {
		t.Error(err)
	}

	err, _ = b.Publish("test", 6)
	if err != nil {
		t.Error(err)
	}

	err, _ = b.Publish("test", 7)
	if err == nil {
		t.Error("expected error, didnt get one")
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
