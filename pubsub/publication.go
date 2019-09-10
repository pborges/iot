package pubsub

type BasicPublication struct {
	key    string
	broker *BasicBroker
}

func (p BasicPublication) Cancel() {
	p.broker.cancelPublication(p.key)
}

func (p BasicPublication) Key() string {
	return p.key
}

func (p BasicPublication) Update(value interface{}) (error, []SubscriptionReport) {
	return nil, p.broker.publish(p.key, value)
}
