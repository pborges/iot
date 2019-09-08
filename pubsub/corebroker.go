package pubsub

import (
	uuid "github.com/satori/go.uuid"
	"sync"
)

type CoreBroker struct {
	publications struct {
		db map[string]OnMessageFn
		*sync.RWMutex
	}
	subscriptions struct {
		db map[string][]CoreSubscription
		*sync.RWMutex
	}
	onCreate    OnCreateFn
	onPublish   OnPublishFn
	onSubscribe OnSubscribeFn
}

func (b *CoreBroker) initPublications() {
	if b.publications.RWMutex == nil {
		b.publications.RWMutex = new(sync.RWMutex)
		b.publications.Lock()
		b.publications.db = make(map[string]OnMessageFn)
		b.publications.Unlock()
	}
}

func (b *CoreBroker) initSubscriptions() {
	if b.subscriptions.RWMutex == nil {
		b.subscriptions.RWMutex = new(sync.RWMutex)
		b.subscriptions.Lock()
		b.subscriptions.db = make(map[string][]CoreSubscription)
		b.subscriptions.Unlock()
	}
}

func (b *CoreBroker) getKey(key string) (OnMessageFn, error) {
	b.publications.RLock()
	defer b.publications.RUnlock()

	if fn, ok := b.publications.db[key]; ok {
		return fn, nil
	}

	return nil, ErrorKeyNotFound
}

func (b *CoreBroker) getSubscription(id string) (CoreSubscription, error) {
	b.subscriptions.RLock()
	defer b.subscriptions.RUnlock()

	for _, subs := range b.subscriptions.db {
		for _, sub := range subs {
			if sub.Id() == id {
				return sub, nil
			}
		}
	}

	return CoreSubscription{}, ErrorSubscriptionNotFound
}

func (b *CoreBroker) keyMatch(key, filter string) bool {
	return KeyMatch(key, filter)
}

func (b *CoreBroker) generateId() string {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return id.String()
}

func (b *CoreBroker) publish(key string, value interface{}) []SubscriptionReport {
	b.initPublications()
	b.publications.Lock()
	defer b.publications.Unlock()

	msg := Message{
		MessageMetadata: MessageMetadata{
			Id:  b.generateId(),
			Key: key,
		},
		Value: value,
	}
	reports := make([]SubscriptionReport, 0)
	b.subscriptions.RLock()
	for f, subs := range b.subscriptions.db {
		if b.keyMatch(key, f) {
			for _, sub := range subs {
				report := SubscriptionReport{
					CoreSubscription: sub,
					err:              nil,
				}
				if sub.fn != nil {
					report.err = sub.fn(msg)
				}
				reports = append(reports, report)
			}
		}
	}
	b.subscriptions.RUnlock()
	if b.onPublish != nil {
		b.onPublish(msg, reports)
	}
	return reports
}

func (b *CoreBroker) cancelPublication(key string) {
	b.initPublications()
	b.publications.Lock()
	delete(b.publications.db, key)
	b.publications.Unlock()
}

func (b *CoreBroker) cancelSubscription(id string) error {
	b.initSubscriptions()
	b.subscriptions.Lock()
	defer b.subscriptions.Unlock()
	for f, subs := range b.subscriptions.db {
		for i, sub := range subs {
			if sub.Id() == id {
				subs = append(subs[:i], subs[i+1:]...)
				b.subscriptions.db[f] = subs
				return nil
			}
		}
	}
	return ErrorSubscriptionAlreadyCanceled
}

func (b *CoreBroker) Create(key string, fn OnMessageFn) (Publication, error) {
	b.initPublications()
	b.publications.Lock()
	if _, ok := b.publications.db[key]; ok {
		return CorePublication{}, ErrorKeyAlreadyDefined
	}
	b.publications.db[key] = fn
	b.publications.Unlock()
	publication := CorePublication{
		broker: b,
		key:    key,
	}

	if b.onCreate != nil {
		b.onCreate(publication)
	}

	return publication, nil
}

func (b *CoreBroker) Publish(key string, value interface{}) (error, []SubscriptionReport) {
	b.initPublications()
	b.initSubscriptions()

	if fn, err := b.getKey(key); err != nil {
		return err, nil
	} else if fn != nil {
		err := fn(Message{
			MessageMetadata: MessageMetadata{
				Id:  b.generateId(),
				Key: key,
			},
			Value: value,
		})
		if err != nil {
			return err, nil
		}
	}

	return nil, b.publish(key, value)
}

func (b *CoreBroker) Subscribe(filter string, fn OnMessageFn) CancelableSubscription {
	b.initSubscriptions()

	sub := CancelableCoreSubscription{
		CoreSubscription: CoreSubscription{
			filter: filter,
			id:     b.generateId(),
			fn:     fn,
		},
		broker: b,
	}

	b.subscriptions.Lock()
	if _, ok := b.subscriptions.db[filter]; !ok {
		b.subscriptions.db[filter] = make([]CoreSubscription, 0)
	}
	b.subscriptions.db[filter] = append(b.subscriptions.db[filter], sub.CoreSubscription)
	b.subscriptions.Unlock()
	if b.onSubscribe != nil {
		b.onSubscribe(sub)
	}
	return sub
}
