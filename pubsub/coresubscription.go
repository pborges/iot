package pubsub

type CoreSubscription struct {
	filter string
	id     string
	fn     OnMessageFn
}

func (s CoreSubscription) Id() string {
	return s.id
}

func (s CoreSubscription) Filter() string {
	return s.filter
}

type CancelableCoreSubscription struct {
	CoreSubscription
	broker *CoreBroker
}

func (s CancelableCoreSubscription) Cancel() error {
	return s.broker.cancelSubscription(s.Id())
}

type SubscriptionReport struct {
	CoreSubscription
	err error
}

func (s SubscriptionReport) Error() error {
	return s.err
}
