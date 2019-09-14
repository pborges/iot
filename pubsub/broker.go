package pubsub

import (
	"errors"
	"fmt"
	"time"
)

type Broker struct {
	attributes map[string]Attribute
	values     map[string]Datum
	clients    clients
}

// delete the attribute from the map, but leave the value
func (b Broker) cancelAttribute(name string) error {
	if b.attributes != nil {
		delete(b.attributes, name)
	}
	return nil
}

func (b *Broker) setAttributeValue(source Source, attr Attribute, value interface{}) {
	if b.values == nil {
		b.values = make(map[string]Datum)
	}
	b.values[attr.name] = Datum{
		Owner: attr.owner.name,
		Name:  attr.name,
		Def:   attr.Definition,
		Value: value,
		By:    source,
		At:    time.Now(),
	}
}

func (b Broker) getAttributeValue(name string) Datum {
	if b.values == nil {
		return Datum{}
	}
	return b.values[name]
}

func (b *Broker) createClient(parent *Client, name string) *Client {
	if parent != nil {
		name = parent.name + "[" + name + "]"
	}
	return &Client{
		parent: parent,
		broker: b,
		name:   name,
	}
}

func (b Broker) fanout(source Source, attr Attribute) []SubscriptionReport {
	reports := make([]SubscriptionReport, 0)

	b.clients.foreach(func(client *Client) bool {
		client.subs.foreach(func(sub Subscription) bool {
			if KeyMatch(attr.name, sub.filter) {
				report := SubscriptionReport{Subscription: sub}
				ctx := Context{
					source: SubscriptionSource{sub: sub},
				}
				ctx.client = b.createClient(client, sub.filter)

				report.Error = sub.fn(attr.name, attr.Value(), ctx)
				fmt.Println("⤷ [OnSubscribeEvent]", "TO:", ctx.Source(), "ATTR:", attr.name, "VALUE:", attr.Value().Value, "SOURCE:", source.String()+"@"+attr.Value().At.Format(time.RFC822))

				reports = append(reports, report)
			}
			return true
		})
		return true
	})

	return reports
}

func (b *Broker) cancelSubscription(client *Client, name string) error {
	return nil
}

func (b *Broker) createAttribute(client *Client, name string, def Definition, acceptFn func(interface{}) error) (Attribute, error, []SubscriptionReport) {
	attr := Attribute{name: name, owner: client, Definition: def, fn: acceptFn}

	// validate the default value if we have a definition
	var value interface{}
	if def != nil {
		var err error
		if value, err = attr.Transform(attr.DefaultValue()); err != nil {
			return Attribute{}, err, nil
		}
		value = attr.DefaultValue()
	}

	if b.attributes == nil {
		b.attributes = make(map[string]Attribute)
	}
	b.attributes[attr.name] = attr

	err, reports := b.selfUpdateAndFanout(client, attr, value)
	return attr, err, reports
}

func (b *Broker) selfUpdateAndFanout(by *Client, attr Attribute, value interface{}) (error, []SubscriptionReport) {
	fmt.Println("[SelfUpdate        ] ATTR:", attr.name, "VALUE:", value)
	// update the value
	source := UpdateSource{client: by}
	b.setAttributeValue(source, attr, value)

	// fanout
	return nil, b.fanout(source, attr)
}

func (b *Broker) publish(source Source, name string, value interface{}) (error, []SubscriptionReport) {
	fmt.Println("[Publish           ] ATTR:", name, "VALUE:", value, "SOURCE:", source)
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
		b.setAttributeValue(source, attr, value)

		return nil, b.fanout(source, attr)
	}
	return errors.New("unknown attribute"), nil
}

func (b *Broker) CreateClient(name string) (*Client, error) {
	client := b.createClient(nil, name)
	if err := b.clients.store(client); err != nil {
		return nil, err
	}
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
