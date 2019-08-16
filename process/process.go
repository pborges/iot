package process

import (
	"errors"
	"github.com/pborges/iot/pubsub"
)

type Process struct {
	Name          string
	broker        pubsub.Broker
	subscriptions []pubsub.CancelableSubscription
}

func (p *Process) Subscribe(filter string, fn OnMessageFn) {
	p.subscriptions = append(p.subscriptions, p.broker.Subscribe(filter, func(m pubsub.Message) error {
		wrap, ok := m.Value.(Message)
		if !ok {
			return errors.New("unable to wrap message")
		}
		wrap.MessageMetadata = m.MessageMetadata
		return fn(wrap)
	}))
}

func (p *Process) Publish(key string, value interface{}) (error, []pubsub.SubscriptionReport) {
	return p.broker.Publish(key, Message{Process: p.Name, Value: value})
}
