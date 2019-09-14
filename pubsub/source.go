package pubsub

type Source interface {
	Client() string
	String() string
}

type ClientSource struct {
	client *Client
	self   bool
}

func (a ClientSource) IsSelf() bool {
	return a.self
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
	return a.sub.client.name + "[" + a.sub.filter + "]{" + a.sub.id + "}"
}
