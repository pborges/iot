package definition

import "github.com/pborges/iot/pubsub"

type Definition interface {
	OnAccept(fn pubsub.AcceptFn)
	Accept(interface{}) error
}

type BaseDefinition struct {
	AcceptFN pubsub.AcceptFn
}

func (d *BaseDefinition) OnAccept(fn pubsub.AcceptFn) {
	if d.AcceptFN == nil {
		d.AcceptFN = fn
		return
	}
	acceptFn := d.AcceptFN
	d.AcceptFN = func(value interface{}) error {
		if err := acceptFn(value); err != nil {
			return err
		}
		return fn(value)
	}
}
