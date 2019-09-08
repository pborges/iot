package pubsub

import (
	"testing"
)

func TestBasicBroker_cancelSubscription(t *testing.T) {
	b := &CoreBroker{}

	sub := b.Subscribe("*", nil)

	if _, err := b.getSubscription(sub.Id()); err != nil {
		t.Error(err)
	}

	if err := sub.Cancel(); err != nil {
		t.Error(err)
	}

	if _, err := b.getSubscription(sub.Id()); err != ErrorSubscriptionNotFound {
		t.Error("expected error got nothing")
	}
}

func TestBasicBroker_Create(t *testing.T) {
	b := &CoreBroker{}
	onCreateCalled := false
	if _, err := b.Create("test", func(k string, v interface{}) error {
		onCreateCalled = true
		return nil
	}); err != nil {
		t.Error(err)
	}

	if _, ok := b.publications.db["test"]; !ok {
		t.Error("key not found")
	}

	if err, _ := b.Publish("test", true); err != nil {
		t.Error(err)
	}

	if !onCreateCalled {
		t.Error("create onMessage was not called")
	}
}
