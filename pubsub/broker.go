package pubsub

type OnMessageFn func(Message) error
type OnCreateFn func(Publication)
type OnPublishFn func(Message, []SubscriptionReport)
type OnSubscribeFn func(CancelableSubscription)

type EventableBroker interface {
	Broker
	OnCreate(OnCreateFn)
	OnPublish(OnPublishFn)
	OnSubscribe(OnSubscribeFn)
}

type Broker interface {
	Create(key string, fn OnMessageFn) (Publication, error)
	Publish(key string, value interface{}) (error, []SubscriptionReport)
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
