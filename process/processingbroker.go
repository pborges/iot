package process

import (
	"errors"
	"github.com/pborges/iot/pubsub"
)

type ProcessingBroker struct {
	pubsub.Broker
	Name          string
	subscriptions []pubsub.CancelableSubscription
}

func (p *ProcessingBroker) Subscribe(filter string, fn OnMessageFn) {
	p.subscriptions = append(p.subscriptions, p.Broker.Subscribe(filter, func(m pubsub.Message) error {
		wrap, ok := m.Value.(Message)
		if !ok {
			return errors.New("unable to wrap message")
		}
		wrap.MessageMetadata = m.MessageMetadata
		return fn(wrap)
	}))
}

func (p *ProcessingBroker) Publish(key string, value interface{}) (error, []pubsub.SubscriptionReport) {
	return p.Broker.Publish(key, Message{Process: p.Name, Value: value})
}

func (p *ProcessingBroker) Create(key string, fn OnMessageFn) (pubsub.Publication, error) {
	return p.Broker.Create(p.Name+"."+key, func(m pubsub.Message) error {
		return fn(Message{
			MessageMetadata: m.MessageMetadata,
			Process:         p.Name,
			Value:           m.Value,
		})
	})
}
