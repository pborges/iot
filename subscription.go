package iot

import (
	"github.com/satori/go.uuid"
	"sync"
)

type OnMessageFn func(name string, value Datum, b Context) error

type SubscriptionReport struct {
	Subscription
	Error error
}

type subscriptions struct {
	db   map[string]Subscription
	lock sync.RWMutex
}

func (c *subscriptions) delete(sub Subscription) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.db != nil {
		if _, ok := c.db[sub.id]; ok {
			delete(c.db, sub.id)
			return nil
		}
	}
	return ErrSubscriptionNotFound
}

func (c *subscriptions) foreach(fn func(sub Subscription) bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, s := range c.db {
		// continue as long as we get true back
		if c := fn(s); !c {
			return
		}
	}
}

func (c *subscriptions) store(sub Subscription) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.db == nil {
		c.db = make(map[string]Subscription)
	}

	if id, err := uuid.NewV4(); err == nil {
		sub.id = id.String()
	} else {
		return err
	}

	if _, ok := c.db[sub.id]; ok {
		return ErrDuplicateSubscription
	}

	c.db[sub.id] = sub
	return nil
}

type Subscription struct {
	id     string
	filter string
	client *Client
	fn     OnMessageFn
}

func (s Subscription) Id() string {
	return s.id
}

func (s Subscription) Filter() string {
	return s.filter
}

func (s Subscription) Cancel() error {
	return s.client.subs.delete(s)
}

// Context provides the subscription callback a way to interact with the broker in an accountable manner
type Context struct {
	source Source
	client *Client
}

func (ctx Context) Source() Source {
	return ctx.source
}

func (ctx Context) Publish(name string, value interface{}) (error, []SubscriptionReport) {
	//make a owner that represents the specific subscription
	return ctx.client.broker.publish(ctx.source, name, value)
}

func (ctx Context) List(filter string) []Datum {
	return ctx.client.broker.List(filter)
}
