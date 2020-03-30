package pubsub

import "time"

type Context interface {
	Publish(attr string, value interface{}) error
	Value(attr string, at time.Time) (ValueRecord, error)
	Error(error)
}
type executionContext struct {
	broker    *Broker
	publisher string
	errors    []error
}

func (ctx *executionContext) Value(attr string, at time.Time) (ValueRecord, error) {
	return ctx.broker.Value(attr, at)
}

func (ctx *executionContext) Error(err error) {
	if err == nil {
		return
	}
	ctx.errors = append(ctx.errors, err)
}

func (ctx *executionContext) Publish(attr string, value interface{}) error {
	return ctx.broker.publish(ctx.publisher, attr, value)
}
