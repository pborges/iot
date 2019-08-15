package pubsub

import (
	uuid "github.com/satori/go.uuid"
	"sync"
)

type Broker interface {
	Create(key string, fn OnMessageFn) error
	Publish(key string, value interface{}) (error, []SubscriptionReport)
	Subscribe(filter string, fn OnMessageFn) CancelableSubscription
}

type OnMessageFn func(Message) error

type BasicBroker struct {
	keys struct {
		db map[string]OnMessageFn
		*sync.RWMutex
	}
	subscriptions struct {
		db map[string][]Subscription
		*sync.RWMutex
	}
}

func (b *BasicBroker) initKeys() {
	if b.keys.RWMutex == nil {
		b.keys.RWMutex = new(sync.RWMutex)
		b.keys.Lock()
		b.keys.db = make(map[string]OnMessageFn)
		b.keys.Unlock()
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
	b.keys.RLock()
	defer b.keys.RUnlock()

	if fn, ok := b.keys.db[key]; ok {
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

func (b *BasicBroker) Create(key string, fn OnMessageFn) error {
	b.initKeys()
	b.keys.Lock()
	defer b.keys.Unlock()
	if _, ok := b.keys.db[key]; ok {
		return ErrorKeyAlreadyDefined
	}
	b.keys.db[key] = fn
	return nil
}

func (b *BasicBroker) Publish(key string, value interface{}) (error, []SubscriptionReport) {
	b.initKeys()
	b.initSubscriptions()
	if fn, err := b.getKey(key); err != nil {
		return err, nil
	} else if fn != nil {
		err := fn(Message{
			Id:    b.generateId(),
			Key:   key,
			Value: value,
		})
		if err != nil {
			return err, nil
		}
	}

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
						Id:    b.generateId(),
						Key:   key,
						Value: value,
					})
				}
				reports = append(reports, report)
			}
		}
	}
	return nil, reports
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
