package pubsub

type OnMessageFn func(Message) error
type OnCreateFn func(Publication)
type OnPublishFn func(Message, []SubscriptionReport)
type OnSubscribeFn func(CancelableSubscription)

type Broker interface {
	Creator
	Publisher
	Subscriber
}

type Creator interface {
	Create(key string, fn OnMessageFn) (Publication, error)
}

type Publisher interface {
	Publish(key string, value interface{}) (error, []SubscriptionReport)
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
