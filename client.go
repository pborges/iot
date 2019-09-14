package iot

import (
	"sync"
)

//TODO should this func sig have BrokerAccess and or an Attribute in the parms?
type OnAcceptFn func(interface{}) error

type Client struct {
	parent *Client
	broker *Broker
	name   string
	subs   subscriptions
}

type subscriptions struct {
	db   map[string]Subscription
	lock sync.RWMutex
}

func (c *subscriptions) cancel(sub Subscription) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.db != nil {
		if _, ok := c.db[sub.filter]; ok {
			delete(c.db, sub.filter)
			return nil
		}
	}
	return ErrNotFound("subscription")
}

func (c *subscriptions) foreach(fn func(sub Subscription) bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for _, s := range c.db {
		if c := fn(s); c {
			return
		}
	}
}

func (c *subscriptions) subscribe(sub Subscription) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.db == nil {
		c.db = make(map[string]Subscription)
	}

	if _, ok := c.db[sub.filter]; ok {
		return ErrDuplicateName
	}
	c.db[sub.filter] = sub
	return nil
}

func (c *Client) createSubClient(name string) *Client {
	return &Client{
		parent: c,
		broker: c.broker,
		name:   c.name + "[" + name + "]",
	}
}

func (c *Client) Create(name string, def Definition, acceptFn OnAcceptFn) (Attribute, error, []SubscriptionReport) {
	// prefix the attribute name with the Client name
	return c.broker.createAttribute(c, c.name+"."+name, def, acceptFn)
}

func (c *Client) Publish(name string, value interface{}) (error, []SubscriptionReport) {
	return c.broker.publish(c, name, value)
}

func (c *Client) Subscribe(filter string, fn OnMessageFn) (Subscription, error) {
	sub := Subscription{
		client: c,
		filter: filter,
		fn:     fn,
	}
	if err := c.subs.subscribe(sub); err != nil {
		return Subscription{}, err
	}
	return sub, nil
}
