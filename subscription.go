package iot

type OnMessageFn func(name string, value Datum, b BrokerAccess) error

type SubscriptionReport struct {
	Subscription
	Error error
}

type Subscription struct {
	filter string
	client *Client
	fn     OnMessageFn
}

func (s Subscription) Name() string {
	return s.client.name + "[" + s.filter + "]"
}

func (s Subscription) Filter() string {
	return s.filter
}

func (s Subscription) Cancel() error {
	return s.client.subs.cancel(s)
}

// BrokerAccess provides the subscription callback a way to interact with the broker in an accountable manner
type BrokerAccess struct {
	sub    Subscription
	client *Client
}

func (b BrokerAccess) Publish(name string, value interface{}) (error, []SubscriptionReport) {
	//make a client that represents the specific subscription
	return b.client.broker.publish(b.client, name, value)
}

func (b BrokerAccess) List(filter string) []Datum {
	return b.client.broker.List(filter)
}
