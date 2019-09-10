package pubsub

type LambdaBroker struct {
	CreateFn    func(string, AcceptFn) (Publication, error)
	PublishFn   func(key string, value interface{}) (error, []SubscriptionReport)
	SubscribeFn func(filter string, fn OnMessageFn) CancelableSubscription
}

func (b LambdaBroker) Create(key string, fn AcceptFn) (Publication, error) {
	if b.CreateFn != nil {
		return b.CreateFn(key, fn)
	}
	return nil, nil
}
func (b LambdaBroker) Publish(key string, value interface{}) (error, []SubscriptionReport) {
	if b.PublishFn != nil {
		return b.PublishFn(key, value)
	}
	return nil, nil
}
func (b LambdaBroker) Subscribe(filter string, fn OnMessageFn) CancelableSubscription {
	if b.SubscribeFn != nil {
		return b.SubscribeFn(filter, fn)
	}
	return nil
}
