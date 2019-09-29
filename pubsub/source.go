package pubsub

import "github.com/robfig/cron"

type Source interface {
	Client() string
	String() string
}

type ClientSource struct {
	client *Client
}

func (a ClientSource) Client() string {
	return a.client.name
}
func (a ClientSource) String() string {
	return a.client.name
}

type SubscriptionSource struct {
	sub Subscription
}

func (a SubscriptionSource) Id() string {
	return a.sub.Id()
}
func (a SubscriptionSource) Filter() string {
	return a.sub.Filter()
}

func (a SubscriptionSource) Client() string {
	return a.sub.client.name
}

func (a SubscriptionSource) String() string {
	return a.sub.client.name + "[" + a.sub.id + ":" + a.sub.filter + "]"
}

type CronSource struct {
	client *Client
	id     string
}

func (a CronSource) entry() cron.EntryID {
	return a.client.cronEntries[a.id]
}

func (a CronSource) Client() string {
	return a.client.name
}
func (a CronSource) String() string {
	return a.client.name + "[cron:" + a.id + "]"
}
