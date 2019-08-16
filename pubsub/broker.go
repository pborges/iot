package pubsub

import (
	uuid "github.com/satori/go.uuid"
	"sync"
)

type Broker interface {
	Create(key string, fn OnMessageFn) (Publication, error)
	Publish(key string, value interface{}) (error, []SubscriptionReport)
	Subscribe(filter string, fn OnMessageFn) CancelableSubscription

	publish(key string, value interface{}) []SubscriptionReport
	cancelPublication(id string)
	cancelSubscription(id string)
}

type OnMessageFn func(Message) error

type BasicBroker struct {
	publications struct {
		db map[string]OnMessageFn
		*sync.RWMutex
	}
	subscriptions struct {
		db map[string][]Subscription
		*sync.RWMutex
	}
}

func (b *BasicBroker) initPublications() {
	if b.publications.RWMutex == nil {
		b.publications.RWMutex = new(sync.RWMutex)
		b.publications.Lock()
		b.publications.db = make(map[string]OnMessageFn)
		b.publications.Unlock()
	}
}

func (b *BasicBroker) initSubscriptions() {
	if b.subscriptions.RWMutex == nil {
		b.subscriptions.RWMutex = new(sync.RWMutex)
		b.subscriptions.Lock()
		b.subscriptions.db = make(map[string][]Subscription)
		b.subscriptions.Unlock()
	}
}

func (b *BasicBroker) getKey(key string) (OnMessageFn, error) {
	b.publications.RLock()
	defer b.publications.RUnlock()

	if fn, ok := b.publications.db[key]; ok {
		return fn, nil
	}

	return nil, ErrorKeyNotFound
}

func (b *BasicBroker) getSubscription(id string) (Subscription, error) {
	b.subscriptions.RLock()
	defer b.subscriptions.RUnlock()

	for _, subs := range b.subscriptions.db {
		for _, sub := range subs {
			if sub.Id() == id {
				return sub, nil
			}
		}
	}

	return Subscription{}, ErrorSubscriptionNotFound
}

func (b *BasicBroker) keyMatch(key, filter string) bool {
	return KeyMatch(key, filter)
}

func (b *BasicBroker) generateId() string {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return id.String()
}

func (b *BasicBroker) publish(key string, value interface{}) []SubscriptionReport {
	b.initPublications()
	b.publications.Lock()
	defer b.publications.Unlock()

	reports := make([]SubscriptionReport, 0)
	b.subscriptions.RLock()
	defer b.subscriptions.RUnlock()
	for f, subs := range b.subscriptions.db {
		if b.keyMatch(key, f) {
			for _, sub := range subs {
				report := SubscriptionReport{
					Subscription: sub,
					err:          nil,
				}
				if sub.fn != nil {
					report.err = sub.fn(Message{
						MessageMetadata: MessageMetadata{
							Id:  b.generateId(),
							Key: key,
						},
						Value: value,
					})
				}
				reports = append(reports, report)
			}
		}
	}
	return reports
}

func (b *BasicBroker) cancelPublication(key string) {
	b.initPublications()
	b.publications.Lock()
	delete(b.publications.db, key)
	b.publications.Unlock()
}

func (b *BasicBroker) cancelSubscription(id string) {
	b.initSubscriptions()
	b.subscriptions.Lock()
	defer b.subscriptions.Unlock()
	for f, subs := range b.subscriptions.db {
		for i, sub := range subs {
			if sub.Id() == id {
				subs = append(subs[:i], subs[i+1:]...)
				b.subscriptions.db[f] = subs
				return
			}
		}
	}
}

func (b *BasicBroker) Create(key string, fn OnMessageFn) (Publication, error) {
	b.initPublications()
	b.publications.Lock()
	defer b.publications.Unlock()
	if _, ok := b.publications.db[key]; ok {
		return Publication{}, ErrorKeyAlreadyDefined
	}
	b.publications.db[key] = fn
	publication := Publication{
		broker: b,
		key:    key,
	}
	return publication, nil
}

func (b *BasicBroker) Publish(key string, value interface{}) (error, []SubscriptionReport) {
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

func (b *BasicBroker) Subscribe(filter string, fn OnMessageFn) CancelableSubscription {
	b.initSubscriptions()

	sub := CancelableSubscription{
		Subscription: Subscription{
			filter: filter,
			id:     b.generateId(),
			fn:     fn,
		},
		broker: b,
	}

	b.subscriptions.Lock()
	defer b.subscriptions.Unlock()
	if _, ok := b.subscriptions.db[filter]; !ok {
		b.subscriptions.db[filter] = make([]Subscription, 0)
	}
	b.subscriptions.db[filter] = append(b.subscriptions.db[filter], sub.Subscription)
	return sub
}
