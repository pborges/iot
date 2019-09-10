package pubsub

import (
	"testing"
)

func TestBroker_Create(t *testing.T) {
	b := &BasicBroker{}
	onCreateCalled := false
	_, err := b.Create("test", func(in interface{}) error {
		onCreateCalled = true
		return nil
	})

	if err != nil {
		t.Error(err)
	}

	if _, ok := b.publications.db["test"]; !ok {
		t.Error("key not found")
	}

	err, _ = b.Publish("test", true)

	if err != nil {
		t.Error(err)
	}

	if !onCreateCalled {
		t.Error("create onCreate was not called")
	}
}

func TestPublication_Update(t *testing.T) {
	b := &BasicBroker{}
	publication, err := b.Create("test", nil)
	if err != nil {
		t.Error(err)
	}

	onSubscribeCalled := false
	b.Subscribe("*", func(key string, value interface{}) error {
		onSubscribeCalled = true
		return nil
	})

	publication.Update(123)

	if !onSubscribeCalled {
		t.Error("onSubscribe not called")
	}
}
