package pubsub

type OnMessageFn func(key string, value interface{}) error

type Subscription struct {
	filter string
	fn     OnMessageFn
}

func (s Subscription) Filter() string {
	return s.filter
}

type CancelableSubscription struct {
	Subscription
	broker *Broker
}

func (s *CancelableSubscription) Cancel() error {
	return s.broker.cancelSubscription(s)
}

type SubscriptionReport struct {
	Subscription
	err error
}

func (s SubscriptionReport) Error() error {
	return s.err
}
