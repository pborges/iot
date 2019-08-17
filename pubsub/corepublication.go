package pubsub

type CorePublication struct {
	key    string
	broker *CoreBroker
}

func (p CorePublication) Cancel() {
	p.broker.cancelPublication(p.key)
}

func (p CorePublication) Key() string {
	return p.key
}

func (p CorePublication) Publish(value interface{}) {
	p.broker.publish(p.key, value)
}
