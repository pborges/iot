package pubsub

import (
	"testing"
)

func TestBasicBroker_cancelSubscription(t *testing.T) {
	b := &BasicBroker{}

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
