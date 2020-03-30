package pubsub

type Node interface {
	NodeId() string
	NodeAttributes() []Attribute
	NodeSubscriptions() []Subscription
}

type BasicNode struct {
	ID            string
	Attributes    []Attribute
	Subscriptions []Subscription
}

func (n BasicNode) NodeId() string {
	return n.ID
}
func (n BasicNode) NodeAttributes() []Attribute {
	return n.Attributes
}
func (n BasicNode) NodeSubscriptions() []Subscription {
	return n.Subscriptions
}
