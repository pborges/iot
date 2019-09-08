package pubsub

import (
	"sync"
)

type TransformFn func(in interface{}) (out interface{}, err error)

type Broker struct {
	publications struct {
		db map[string]TransformFn
		*sync.RWMutex
	}
	subscriptions struct {
		db map[string][]*CancelableSubscription
		*sync.RWMutex
	}
}

func (b *Broker) initPublications() {
	if b.publications.RWMutex == nil {
		b.publications.RWMutex = new(sync.RWMutex)
		b.publications.Lock()
		b.publications.db = make(map[string]TransformFn)
		b.publications.Unlock()
	}
}

func (b *Broker) initSubscriptions() {
	if b.subscriptions.RWMutex == nil {
		b.subscriptions.RWMutex = new(sync.RWMutex)
		b.subscriptions.Lock()
		b.subscriptions.db = make(map[string][]*CancelableSubscription)
		b.subscriptions.Unlock()
	}
}

func (b *Broker) getKeyTransformFunc(key string) (TransformFn, error) {
	b.publications.RLock()
	defer b.publications.RUnlock()

	if fn, ok := b.publications.db[key]; ok {
		return fn, nil
	}

	return nil, ErrorKeyNotFound
}

func (b *Broker) keyMatch(key, filter string) bool {
	return KeyMatch(key, filter)
}

func (b *Broker) publish(key string, value interface{}) []SubscriptionReport {
	b.initPublications()
	b.initSubscriptions()

	b.publications.Lock()
	defer b.publications.Unlock()

	reports := make([]SubscriptionReport, 0)
	b.subscriptions.RLock()
	for f, subs := range b.subscriptions.db {
		if b.keyMatch(key, f) {
			// fanout the value and collect the reports
			for _, sub := range subs {
				report := SubscriptionReport{
					Subscription: sub.Subscription,
					err:          nil,
				}
				if sub.fn != nil {
					report.err = sub.fn(key, value)
				}
				reports = append(reports, report)
			}
		}
	}
	b.subscriptions.RUnlock()
	return reports
}

func (b *Broker) cancelPublication(key string) {
	b.initPublications()
	b.publications.Lock()
	delete(b.publications.db, key)
	b.publications.Unlock()
}

func (b *Broker) cancelSubscription(s *CancelableSubscription) error {
	b.initSubscriptions()
	b.subscriptions.Lock()
	defer b.subscriptions.Unlock()
	for f, subs := range b.subscriptions.db {
		for i, sub := range subs {
			if s == sub {
				subs = append(subs[:i], subs[i+1:]...)
				b.subscriptions.db[f] = subs
				return nil
			}
		}
	}
	return ErrorSubscriptionAlreadyCanceled
}
func (b *Broker) Publish(key string, value interface{}) (error, []SubscriptionReport) {
	b.initPublications()
	b.initSubscriptions()

	if transformFn, err := b.getKeyTransformFunc(key); err == nil {
		// transform the value
		value, err = transformFn(value)
		if err != nil {
			return err, nil
		}
	} else {
		return err, nil
	}

	return nil, b.publish(key, value)
}

func (b *Broker) Create(key string, fn TransformFn) (Publication, error) {
	b.initPublications()
	b.publications.Lock()
	if _, ok := b.publications.db[key]; ok {
		return Publication{}, ErrorKeyAlreadyDefined
	}
	b.publications.db[key] = fn
	b.publications.Unlock()
	publication := Publication{
		broker: b,
		key:    key,
	}

	return publication, nil
}

func (b *Broker) Subscribe(filter string, fn OnMessageFn) *CancelableSubscription {
	b.initSubscriptions()

	sub := &CancelableSubscription{
		Subscription: Subscription{
			filter: filter,
			fn:     fn,
		},
		broker: b,
	}

	b.subscriptions.Lock()
	if _, ok := b.subscriptions.db[filter]; !ok {
		b.subscriptions.db[filter] = make([]*CancelableSubscription, 0)
	}
	b.subscriptions.db[filter] = append(b.subscriptions.db[filter], sub)
	b.subscriptions.Unlock()
	return sub
}
