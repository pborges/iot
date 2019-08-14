package pubsub

import "testing"

func TestCancelableSubscription_Cancel(t *testing.T) {
	s := CancelableSubscription{
		Subscription: Subscription{},
		broker:       &BasicBroker{},
	}

	if s.broker == nil {
		t.Fail()
	}

	if err := s.Cancel(); err != nil {
		t.Error(err)
	}

	if s.broker != nil {
		t.Fail()
	}

	if err := s.Cancel(); err != ErrorSubscriptionAlreadyCanceled {
		t.Error("did not get the expected error")
	}
}
