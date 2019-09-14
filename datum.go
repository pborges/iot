package iot

import "time"

type Datum struct {
	Owner string
	Name  string
	Def   Definition
	Value interface{}
	By    Source
	At    time.Time
}

type Source interface {
	Client() string
	String() string
}

type UpdateSource struct {
	client *Client
}

func (a UpdateSource) Client() string {
	return a.client.name
}
func (a UpdateSource) String() string {
	return a.client.name
}

type PublishSource struct {
	client *Client
}

func (a PublishSource) Client() string {
	return a.client.name
}
func (a PublishSource) String() string {
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
	return a.sub.client.name + "[" + a.sub.filter + "]{" + a.sub.id + "}"
}
