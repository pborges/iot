package pubsub

type AcceptFn func(value interface{}) error

type Publication interface {
	Key() string
	Update(value interface{}) (error, []SubscriptionReport)
}

type Subscription interface {
	Filter() string
}

type CancelableSubscription interface {
	Subscription
	Cancel() error
}

type SubscriptionReport interface {
	Subscription
	Error() error
}

type Broker interface {
	Create(key string, fn AcceptFn) (Publication, error)
	Publish(key string, value interface{}) (error, []SubscriptionReport)
	Subscribe(filter string, fn OnMessageFn) CancelableSubscription
}
