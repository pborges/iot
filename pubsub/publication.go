package pubsub

type Publication struct {
	key    string
	broker *Broker
}

func (p Publication) Cancel() {
	p.broker.cancelPublication(p.key)
}

func (p Publication) Key() string {
	return p.key
}

func (p Publication) Update(value interface{}) {
	p.broker.publish(p.key, value)
}
