package pubsub

import "testing"

func TestCancelableSubscription_Cancel(t *testing.T) {
	broker := &CoreBroker{}

	s := broker.Subscribe("*", nil)

	if err := s.Cancel(); err != nil {
		t.Error(err)
	}

	if err := s.Cancel(); err != ErrorSubscriptionAlreadyCanceled {
		t.Error("did not get the expected error")
	}
}
