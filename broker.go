package iot

import (
	"errors"
	"fmt"
	"time"
)

type Broker struct {
	attributes map[string]Attribute
	values     map[string]Datum
	clients    map[string]*Client
}

// delete the attribute from the map, but leave the value
func (b Broker) cancelAttribute(name string) error {
	if b.attributes != nil {
		delete(b.attributes, name)
	}
	return nil
}

func (b *Broker) setAttributeValue(by *Client, attr Attribute, value interface{}) {
	if b.values == nil {
		b.values = make(map[string]Datum)
	}
	b.values[attr.name] = Datum{
		Name:  attr.name,
		Def:   attr.Definition,
		Value: value,
		By:    by.name,
		At:    time.Now(),
	}
}

func (b Broker) getAttributeValue(name string) Datum {
	if b.values == nil {
		return Datum{}
	}
	return b.values[name]
}

func (b Broker) fanout(attr Attribute) []SubscriptionReport {
	reports := make([]SubscriptionReport, 0)
	if b.clients != nil {
		for _, client := range b.clients {
			for _, sub := range client.subs {
				if KeyMatch(attr.name, sub.filter) {
					report := SubscriptionReport{Subscription: sub}
					responder := BrokerAccess{
						sub:    sub,
						client: client.subClient(sub.filter),
					}
					report.Error = sub.fn(attr.name, attr.Value(), responder)
					fmt.Println("[OnSubscribeEvent]", "TO:", responder.client.name, "ATTR:", attr.name, "VALUE:", attr.Value().Value, "BY", attr.Value().By+"@"+attr.Value().At.Format(time.RFC822))
					reports = append(reports, report)
				}
			}
		}
	}
	return reports
}

func (b *Broker) cancelSubscription(client *Client, name string) error {
	return nil
}

func (b *Broker) createAttribute(client *Client, name string, def Definition, acceptFn func(interface{}) error) (Attribute, error, []SubscriptionReport) {
	attr := Attribute{name: name, client: client, Definition: def, fn: acceptFn}

	// validate the default value if we have a definition
	var value interface{}
	if def != nil {
		var err error
		if value, err = def.Transform(def.DefaultValue()); err != nil {
			return Attribute{}, err, nil
		}
		value = def.DefaultValue()
	}

	if b.attributes == nil {
		b.attributes = make(map[string]Attribute)
	}
	b.attributes[attr.name] = attr

	err, reports := b.updateAndFanout(client, attr, value)
	return attr, err, reports
}

func (b *Broker) updateAndFanout(by *Client, attr Attribute, value interface{}) (error, []SubscriptionReport) {
	fmt.Println("[SelfUpdate      ] ATTR:", attr.name, "VALUE:", value)
	// update the value
	b.setAttributeValue(by, attr, value)

	// fanout
	return nil, b.fanout(attr)
}

func (b *Broker) publish(by *Client, name string, value interface{}) (error, []SubscriptionReport) {
	fmt.Println("[Publish         ] ATTR:", name, "VALUE:", value, "BY:", by.name)
	if attr, ok := b.attributes[name]; ok {
		// validate the value
		var err error
		if value, err = attr.Transform(value); err != nil {
			return err, nil
		}

		// try to run the accept fn
		if attr.fn != nil {
			if err := attr.fn(value); err != nil {
				return err, nil
			}
		}

		// update the attribute value
		b.setAttributeValue(by, attr, value)

		return nil, b.fanout(attr)
	}
	return errors.New("unknown attribute"), nil
}

func (b *Broker) subscribe(client *Client, filter string, fn OnMessageFn) (Subscription, error) {
	if _, ok := client.subs[filter]; ok {
		return Subscription{}, ErrDuplicateName
	}
	sub := Subscription{
		client: client,
		filter: filter,
		fn:     fn,
	}
	client.subs[filter] = sub
	return sub, nil
}

func (b *Broker) CreateClient(name string) (*Client, error) {
	if b.clients == nil {
		b.clients = make(map[string]*Client)
	}
	if _, ok := b.clients[name]; ok {
		return nil, ErrDuplicateName
	}
	client := &Client{
		broker: b,
		name:   name,
		subs:   map[string]Subscription{},
	}
	b.clients[name] = client
	return client, nil
}

func (b Broker) List(filter string) []Datum {
	var data []Datum
	// loop through the values, the attributes may be long gone
	if b.values != nil {
		for _, datum := range b.values {
			if KeyMatch(datum.Name, filter) {
				data = append(data, datum)
			}
		}
	}
	return data
}
