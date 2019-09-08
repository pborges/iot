package pubsub

import "fmt"

type RecordingBroker struct {
	Next Broker
}

func (b RecordingBroker) Create(key string, fn OnMessageFn) (Publication, error) {
	fmt.Println("create:", key)
	pub, err := b.Next.Create(key, fn)
	fmt.Println("create res:", pub.Key(), err)
	return pub, err
}

func (b RecordingBroker) Publish(key string, value interface{}) (error, []SubscriptionReport) {
	return b.Next.Publish(key, value)
}

func (b RecordingBroker) Subscribe(filter string, fn OnMessageFn) CancelableSubscription {
	return b.Next.Subscribe(filter, fn)
}
