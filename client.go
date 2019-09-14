package iot

//TODO should this func sig have BrokerAccess and or an Attribute in the parms?
type OnAcceptFn func(interface{}) error

type Client struct {
	broker *Broker
	name   string
	subs   map[string]Subscription
}

func (c *Client) subClient(name string) *Client {
	return &Client{
		broker: c.broker,
		name:   c.name + "[" + name + "]",
		subs:   map[string]Subscription{},
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
	if _, ok := c.subs[filter]; ok {
		return Subscription{}, ErrDuplicateName
	}
	sub := Subscription{
		client: c,
		filter: filter,
		fn:     fn,
	}
	c.subs[filter] = sub
	return sub, nil
}
