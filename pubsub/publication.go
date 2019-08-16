package pubsub

type Publication struct {
	key    string
	broker Broker
}

func (p *Publication) Cancel() {
	if p.broker != nil {
		p.broker.cancelPublication(p.key)
		p.broker = nil
	}
}

func (p Publication) Key() string {
	return p.key
}

func (p Publication) Publish(value interface{}) {
	p.broker.publish(p.key, value)
}
