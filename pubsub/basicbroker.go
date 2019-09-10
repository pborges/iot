package pubsub

import (
	"sync"
)

type BasicBroker struct {
	publications struct {
		db map[string]AcceptFn
		*sync.RWMutex
	}
	subscriptions struct {
		db map[string][]*BasicCancelableSubscription
		*sync.RWMutex
	}
}

func (b *BasicBroker) initPublications() {
	if b.publications.RWMutex == nil {
		b.publications.RWMutex = new(sync.RWMutex)
		b.publications.Lock()
		b.publications.db = make(map[string]AcceptFn)
		b.publications.Unlock()
	}
}

func (b *BasicBroker) initSubscriptions() {
	if b.subscriptions.RWMutex == nil {
		b.subscriptions.RWMutex = new(sync.RWMutex)
		b.subscriptions.Lock()
		b.subscriptions.db = make(map[string][]*BasicCancelableSubscription)
		b.subscriptions.Unlock()
	}
}

func (b *BasicBroker) getAcceptFn(key string) (AcceptFn, error) {
	b.publications.RLock()
	defer b.publications.RUnlock()

	if fn, ok := b.publications.db[key]; ok {
		return fn, nil
	}

	return nil, ErrorKeyNotFound
}

func (b *BasicBroker) keyMatch(key, filter string) bool {
	return KeyMatch(key, filter)
}

func (b *BasicBroker) publish(key string, value interface{}) []SubscriptionReport {
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
				report := BasicSubscriptionReport{
					Subscription: sub.BasicSubscription,
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

func (b *BasicBroker) cancelPublication(key string) {
	b.initPublications()
	b.publications.Lock()
	delete(b.publications.db, key)
	b.publications.Unlock()
}

func (b *BasicBroker) cancelSubscription(s *BasicCancelableSubscription) error {
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
func (b *BasicBroker) Publish(key string, value interface{}) (error, []SubscriptionReport) {
	b.initPublications()
	b.initSubscriptions()

	if transformFn, err := b.getAcceptFn(key); err == nil {
		err = transformFn(value)
		if err != nil {
			return err, nil
		}
	} else {
		return err, nil
	}

	return nil, b.publish(key, value)
}

func (b *BasicBroker) Create(key string, fn AcceptFn) (Publication, error) {
	b.initPublications()
	b.publications.Lock()
	if _, ok := b.publications.db[key]; ok {
		return nil, ErrorKeyAlreadyDefined
	}
	b.publications.db[key] = fn
	b.publications.Unlock()

	publication := &BasicPublication{
		broker: b,
		key:    key,
	}

	return publication, nil
}

func (b *BasicBroker) Subscribe(filter string, fn OnMessageFn) CancelableSubscription {
	b.initSubscriptions()

	sub := &BasicCancelableSubscription{
		BasicSubscription: BasicSubscription{
			filter: filter,
			fn:     fn,
		},
		broker: b,
	}

	b.subscriptions.Lock()
	if _, ok := b.subscriptions.db[filter]; !ok {
		b.subscriptions.db[filter] = make([]*BasicCancelableSubscription, 0)
	}
	b.subscriptions.db[filter] = append(b.subscriptions.db[filter], sub)
	b.subscriptions.Unlock()
	return sub
}
