package pubsub

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"time"
)

type Broker struct {
	Log           *log.Logger
	attributes    map[string]*attributeCtx
	subscriptions map[string]*subscriptionCtx
}

type attributeCtx struct {
	Attribute Attribute
	Records   []ValueRecord
}

type subscriptionCtx struct {
	Subscription
}

func (ctx Broker) log() *log.Logger {
	if ctx.Log == nil {
		return log.New(ioutil.Discard, "", 0)
	}
	return ctx.Log
}

func (ctx *Broker) publish(publisher string, attr string, value interface{}) (err error) {
	ctx.log().Printf("publish attribute:'%s' publisher:'%s'", attr, publisher)
	defer func() {
		if err != nil {
			ctx.log().Printf("error publish attribute:'%s' publisher:'%s' err: %s", attr, publisher, err)
		}
	}()
	if ctx.attributes != nil {
		if recCtx, ok := ctx.attributes[attr]; ok {
			value, err = recCtx.Attribute.Definition.ValidateAndTransform(value)
			if err != nil {
				err = fmt.Errorf("validateAndTransform error %w, thrown by '%s'", err, attr)
				return
			}
			if !isOwner(attr, publisher) {
				if err = recCtx.Attribute.Definition.Accept(value); err != nil {
					err = fmt.Errorf("accept error %w, thrown by '%s'", err, attr)
					return
				}
			}

			rec := ValueRecord{
				RecordId: len(recCtx.Records),
				Value: Value{
					AttributeID: attr,
					Value:       value,
					inspected:   recCtx.Attribute.Definition.Inspect(value),
					UpdatedBy:   publisher,
					UpdatedAt:   time.Now(),
				},
			}

			ctx.log().Printf("set attribute:'%s' value:'%s' publisher:'%s'", attr, rec.Value.Inspect(), rec.UpdatedBy)
			if ctx.subscriptions != nil {
				for k, sub := range ctx.subscriptions {
					keyMatch := KeyMatch(attr, sub.Subscription.Filter)
					if keyMatch {
						ctx.log().Printf("fanout subscription:'%s' publisher: '%s' filter: '%s' attribute:'%s' value:'%s'", k, publisher, sub.Subscription.Filter, attr, rec.Value.Inspect())
						execCtx := &executionContext{
							broker:    ctx,
							publisher: k,
						}
						sub.Fn(execCtx, rec.Value)
						res := SubscriptionResponse{
							SubscriptionID: k,
						}
						for _, err := range execCtx.errors {
							ctx.log().Printf("error fanout subscription:'%s' attribute:'%s' value:'%s' err: %s", k, attr, rec.Value.Inspect(), err)
							res.Err = append(res.Err, err)
						}
						rec.SubscriptionResponses = append(rec.SubscriptionResponses, res)
					} else {
						ctx.log().Printf("skip fanout subscription:'%s' publisher: '%s' filter; '%s' attribute:'%s' value:'%s'", k, publisher, sub.Subscription.Filter, attr, rec.Value.Inspect())
					}
				}
			}
			recCtx.Records = append(recCtx.Records, rec)
			return
		}
	}
	err = ErrUnknownAttribute{Attribute: attr}
	return
}

func (ctx attributeCtx) Value(at time.Time) (ValueRecord, error) {
	for i := len(ctx.Records) - 1; i >= 0; i-- {
		if ctx.Records[i].UpdatedAt.Before(at) {
			return ctx.Records[i], nil
		}
	}
	return ValueRecord{}, ErrNoValue{Attribute: ctx.Attribute.Name, Timestamp: at}
}

func (ctx *Broker) Values(filter string, at time.Time) []ValueRecord {
	var recs []ValueRecord
	if ctx.attributes != nil {
		for k, a := range ctx.attributes {
			if KeyMatch(k, filter) {
				if v, err := a.Value(at); err == nil {
					recs = append(recs, v)
				}
			}
		}
	}
	return recs
}

func (ctx *Broker) Value(attr string, at time.Time) (ValueRecord, error) {
	if ctx.attributes != nil {
		if rec, ok := ctx.attributes[attr]; ok {
			return rec.Value(at)
		}
	}
	return ValueRecord{}, ErrUnknownAttribute{Attribute: attr}
}

func (ctx *Broker) Publish(publisher Node, attr string, value interface{}) error {
	return ctx.publish(publisher.NodeId(), attr, value)
}

func (ctx *Broker) Register(n Node) error {
	ctx.log().Println("register node", n.NodeId())
	if ctx.attributes == nil {
		ctx.attributes = make(map[string]*attributeCtx)
	}
	if ctx.subscriptions == nil {
		ctx.subscriptions = make(map[string]*subscriptionCtx)
	}

	for _, attr := range n.NodeAttributes() {
		// TODO validate attr.Name
		// TODO duplicate attr
		id := fmt.Sprintf("%s.%s", n.NodeId(), attr.Name)
		if attr.Definition == nil {
			return fmt.Errorf("definition cannot be nil for attribute:'%s'", id)
		}
		ctx.log().Printf("register attribute: '%s' def: '%s'", id, reflect.TypeOf(attr.Definition).Name())
		ctx.attributes[id] = &attributeCtx{
			Attribute: attr,
		}
		if err := ctx.Publish(n, id, attr.Definition.DefaultValue()); err != nil {
			return err
		}
	}

	for _, sub := range n.NodeSubscriptions() {
		// TODO validate sub.Name
		// TODO duplicate sub
		id := fmt.Sprintf("%s@%s", n.NodeId(), sub.Name)
		ctx.log().Printf("register subscription: '%s' filter: '%s'", id, sub.Filter)
		ctx.subscriptions[id] = &subscriptionCtx{
			Subscription: sub,
		}
	}

	return nil
}
