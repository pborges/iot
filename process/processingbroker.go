package process

import (
	"errors"
	"github.com/pborges/iot/pubsub"
)

type ProcessingBroker struct {
	name          string
	broker        pubsub.Broker
	subscriptions []pubsub.CancelableSubscription
}

func (b *ProcessingBroker) Wrap(next pubsub.Broker) {
	b.broker = next
}

func (b *ProcessingBroker) Create(key string, fn pubsub.OnMessageFn) (Publication, error) {
	pub, err := b.broker.Create(b.name+"."+key, fn)
	return Publication{
		broker:      b,
		Publication: pub,
	}, err
}

func (b *ProcessingBroker) Subscribe(filter string, fn OnMessageFn) {
	sub := b.broker.Subscribe(filter, func(m pubsub.Message) error {
		wrap, ok := m.Value.(Message)
		if !ok {
			return errors.New("unable to wrap message")
		}
		return fn(wrap)
	})
	b.subscriptions = append(b.subscriptions, sub)
}

type Publication struct {
	broker *ProcessingBroker
	pubsub.Publication
}

func (p Publication) Publish(value interface{}) {
	p.Publication.Publish(Message{
		Process: p.broker.name,
		Value:   value,
	})
}
