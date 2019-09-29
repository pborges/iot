package pubsub

import (
	"github.com/robfig/cron"
	uuid "github.com/satori/go.uuid"
	"time"
)

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
	parent      *Client
	broker      *Broker
	name        string
	subs        subscriptions
	cronEntries map[string]cron.EntryID
}

func (c Client) Name() string {
	return c.name
}

func (c *Client) CreateAttribute(name string, def Definition, acceptFns ...OnAcceptFn) (Attribute, error, []SubscriptionReport) {
	// prefix the attribute name with the Client name
	return c.broker.createAttribute(c, c.name+"."+name, def, acceptFns...)
}

func (c *Client) Publish(name string, value interface{}) (error, []SubscriptionReport) {
	return c.broker.publish(ClientSource{client: c}, name, value)
}

func (c *Client) Subscribe(id string, filter string, fns ...OnMessageFn) (Subscription, error) {
	sub := Subscription{
		client: c,
		id:     id,
		filter: filter,
		fns:    fns,
	}
	if err := c.subs.store(sub); err != nil {
		return Subscription{}, err
	}
	return sub, nil
}

func (c *Client) schedule(schedule cron.Schedule, fn func(ctx Context)) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	entry := c.broker.schedule(schedule, func() {
		ctx := Context{
			source: CronSource{
				client: c,
				id:     id.String(),
			},
			client: c,
		}
		fn(ctx)
	})
	c.cronEntries[id.String()] = entry
	return err
}

func (c *Client) Schedule(at time.Time, fn func(ctx Context)) error {
	return c.schedule(atSpecificTime(at), func(ctx Context) {
		fn(ctx)
		c.cancelCron(ctx.source.(CronSource).id)
	})
}

func (c *Client) ScheduleEvery(every time.Duration, fn func(ctx Context)) error {
	return c.schedule(cron.Every(every), fn)
}

func (c *Client) cancelCron(id string) {
	c.broker.cron.Remove(c.cronEntries[id])
}

type atSpecificTime time.Time

func (a atSpecificTime) Next(t time.Time) time.Time {
	if t.Before(time.Time(a)) {
		return time.Time(a)
	}
	return time.Time{}
}
