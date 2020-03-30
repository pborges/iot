package pubsub

type Subscription struct {
	Name   string
	Filter string
	Fn     func(ctx Context, v Value)
}
