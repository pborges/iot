package iot

import (
	"time"
)

type Attribute struct {
	Definition
	name   string
	client *Client
	fn     func(interface{}) error
}

func (k Attribute) Cancel() error {
	return k.client.broker.cancelAttribute(k.name)
}

func (k Attribute) Value() Datum {
	return k.client.broker.getAttributeValue(k.name)
}

func (k Attribute) Update(value interface{}) (error, []SubscriptionReport) {
	value, err := k.Definition.Transform(value)
	if err != nil {
		return err, nil
	}
	return k.client.broker.updateAndFanout(k.client, k, value)
}

type Datum struct {
	Name  string
	Def   Definition
	Value interface{}
	By    string
	At    time.Time
}
