package pubsub

type OnAcceptFn func(Source, interface{}) error

type clients struct {
	db map[string]*Client
}

func (c *clients) delete(client *Client) (error, []SubscriptionReport) {
	var reports []SubscriptionReport
	if c.db != nil {
		if _, ok := c.db[client.name]; ok {
			client.subs.foreach(func(sub Subscription) bool {
				report := SubscriptionReport{
					Source: SubscriptionSource{sub: sub},
					Error:  sub.Cancel(),
				}
				reports = append(reports, report)
				return true
			})
			client.broker = nil
			delete(c.db, client.name)
			return nil, reports
		}
	}
	return ErrClientNotFound, reports
}

func (c *clients) foreach(fn func(client *Client) bool) {
	for _, s := range c.db {
		if c := fn(s); !c {
			return
		}
	}
}

func (c *clients) store(client *Client) error {
	if c.db == nil {
		c.db = make(map[string]*Client)
	}

	if _, ok := c.db[client.name]; ok {
		return ErrDuplicateClient
	}

	c.db[client.name] = client
	return nil
}

type Client struct {
	parent *Client
	broker *Broker
	name   string
	subs   subscriptions
}

func (c Client) Name() string {
	return c.name
}

func (c *Client) CreateAttribute(name string, def Definition, acceptFn OnAcceptFn) (Attribute, error, []SubscriptionReport) {
	// prefix the attribute name with the Client name
	return c.broker.createAttribute(c, c.name+"."+name, def, acceptFn)
}

func (c *Client) Publish(name string, value interface{}) (error, []SubscriptionReport) {
	return c.broker.publish(ClientSource{client: c}, name, value)
}

func (c *Client) Subscribe(filter string, fn OnMessageFn) (Subscription, error) {
	sub := Subscription{
		client: c,
		filter: filter,
		fn:     fn,
	}
	if err := c.subs.store(sub); err != nil {
		return Subscription{}, err
	}
	return sub, nil
}
