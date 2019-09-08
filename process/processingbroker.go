package process

import (
	"github.com/pborges/iot/pubsub"
)

type Message struct {
	Process string
	Key     string
	Value   interface{}
}

type ProcessingBroker struct {
	name          string
	broker        *pubsub.Broker
	subscriptions []*pubsub.CancelableSubscription
}

func (b *ProcessingBroker) Wrap(next *pubsub.Broker) {
	b.broker = next
}

func (b *ProcessingBroker) Create(key string, fn pubsub.TransformFn) (Publication, error) {
	// prefix the key
	key = b.name + "." + key
	// wrap the transformation with another transformation into Message
	pub, err := b.broker.Create(key, func(in interface{}) (interface{}, error) {
		// do the provided transformation
		out, err := fn(in)
		// wrap the transformed value and add Process specific metadata
		return Message{
			Key:     key,
			Process: b.name,
			Value:   out,
		}, err
	})

	return Publication{
		broker:      b,
		Publication: pub,
	}, err
}

func (b *ProcessingBroker) Subscribe(filter string, fn func(metadata Message) error) {
	sub := b.broker.Subscribe(filter, func(key string, value interface{}) error {
		return fn(value.(Message))
	})
	b.subscriptions = append(b.subscriptions, sub)
}

type Publication struct {
	broker *ProcessingBroker
	pubsub.Publication
}
