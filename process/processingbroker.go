package process

import (
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

func (b *ProcessingBroker) Create(key string, fn func(MetaData) error) (Publication, error) {
	pub, err := b.broker.Create(b.name+"."+key, func(key string, value interface{}) error {
		data := MetaData{
			Process: b.name,
			Key:     key,
			Value:   value,
		}
		return fn(data)
	})

	return Publication{
		broker:      b,
		Publication: pub,
	}, err
}

func (b *ProcessingBroker) Subscribe(filter string, fn OnMessageFn) {
	sub := b.broker.Subscribe(filter, func(key string, value interface{}) error {
		data := MetaData{
			Process: b.name,
			Key:     key,
			Value:   value,
		}
		return fn(data)
	})
	b.subscriptions = append(b.subscriptions, sub)
}

type Publication struct {
	broker *ProcessingBroker
	pubsub.Publication
}

func (p Publication) Publish(value interface{}) {
	p.Publication.Publish(MetaData{
		Process: p.broker.name,
		Value:   value,
	})
}
