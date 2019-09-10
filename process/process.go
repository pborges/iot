package process

import (
	"github.com/pborges/iot/pubsub"
)

type Process struct {
	pubsub.Broker
	name          string
	publications  []pubsub.Publication
	subscriptions []pubsub.CancelableSubscription
}

func (b Process) Name() string {
	return b.name
}

func (b Process) Create(key string, fn pubsub.AcceptFn) (pubsub.Publication, error) {
	if pub, err := b.Broker.Create(b.name+"."+key, fn); err == nil {
		b.publications = append(b.publications, pub)
		return pub, nil
	} else {
		return nil, err
	}
}

func (b Process) Subscribe(filter string, fn pubsub.OnMessageFn) {
	b.subscriptions = append(b.subscriptions, b.Broker.Subscribe(filter, fn))
}
