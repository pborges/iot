package pubsub

type Subscription struct {
	filter string
	id     string
	fn     OnMessageFn
}

func (s Subscription) Id() string {
	return s.id
}

func (s Subscription) Filter() string {
	return s.filter
}

type CancelableSubscription struct {
	Subscription
	broker *BasicBroker
}

func (s *CancelableSubscription) Cancel() error {
	if s.broker == nil {
		return ErrorSubscriptionAlreadyCanceled
	}
	s.broker.cancelSubscription(s.Id())
	s.broker = nil
	return nil
}

type SubscriptionReport struct {
	Subscription
	err error
}

func (s SubscriptionReport) Error() error {
	return s.err
}
