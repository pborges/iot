package pubsub

type OnMessageFn func(key string, value interface{}) error

type BasicSubscription struct {
	filter string
	fn     OnMessageFn
}

func (s BasicSubscription) Filter() string {
	return s.filter
}

type BasicCancelableSubscription struct {
	BasicSubscription
	broker *BasicBroker
}

func (s *BasicCancelableSubscription) Cancel() error {
	return s.broker.cancelSubscription(s)
}

type BasicSubscriptionReport struct {
	Subscription
	err error
}

func (s BasicSubscriptionReport) Error() error {
	return s.err
}
