package pubsub

type OnMessageFn func(key string, value interface{}) error
type OnCreateFn func(pub Publication)
type OnSubscribeFn func(CancelableSubscription)

type Broker interface {
	Creator
	Subscriber
}

type Creator interface {
	Create(key string, fn OnMessageFn) (Publication, error)
}

type Subscriber interface {
	Subscribe(filter string, fn OnMessageFn) CancelableSubscription
}

type Subscription interface {
	Id() string
	Filter() string
}

type CancelableSubscription interface {
	Subscription
	Cancel() error
}

type Publication interface {
	Key() string
	Publish(value interface{})
	Cancel()
}
