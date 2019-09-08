package pubsub

import "testing"

func TestCancelableSubscription_Cancel(t *testing.T) {
	b := &Broker{}

	sub := b.Subscribe("*", nil)

	if len(b.subscriptions.db["*"]) != 1 {
		t.Error("expected one subscription, got ", len(b.subscriptions.db))
	}

	if err := sub.Cancel(); err != nil {
		t.Error(err)
	}

	if len(b.subscriptions.db["*"]) != 0 {
		t.Error("expected zero subscriptions, got ", len(b.subscriptions.db))
	}
}