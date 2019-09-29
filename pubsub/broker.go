package pubsub

import (
	"errors"
	"fmt"
	"github.com/robfig/cron"
	"time"
)

type Broker struct {
	attributes map[string]Attribute
	values     map[string]Datum
	clients    clients
	cron       *cron.Cron
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
		cronEntries: map[string]cron.EntryID{},
		parent:      parent,
		broker:      b,
		name:        name,
	}
}

func (b Broker) fanout(source Source, attr Attribute) []SubscriptionReport {
	reports := make([]SubscriptionReport, 0)

	b.clients.foreach(func(client *Client) bool {
		// dont fanout to yourself
		if attr.owner.name != client.name {
			client.subs.foreach(func(sub Subscription) bool {
				if KeyMatch(attr.name, sub.filter) {
					//if KeyMatch(attr.name, sub.filter) {
					ctx := Context{
						source: SubscriptionSource{sub: sub},
					}
					report := SubscriptionReport{Source: ctx.source}
					ctx.client = b.createClient(client, sub.filter)
					for _, fn := range sub.fns {
						fmt.Println("⤷ [OnSubscribeEvent]", source, "->", ctx.source, "@"+attr.Value().At.Format(time.Stamp), "ATTR:", attr.name, "VALUE:", attr.Value().Value)
						report.Error = fn(attr.name, attr.Value(), ctx)
						if report.Error != nil {
							fmt.Println("[OnSubscriptionEvent] ERROR:", report.Error)
						}
					}

					reports = append(reports, report)
				}
				return true
			})
		}
		return true
	})

	return reports
}

func (b *Broker) createAttribute(client *Client, name string, def Definition, acceptFns ...OnAcceptFn) (Attribute, error, []SubscriptionReport) {
	attr := Attribute{name: name, owner: client, Definition: def, fns: acceptFns}

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

	err, reports := b.selfUpdateAndFanout(attr, value)
	return attr, err, reports
}

func (b *Broker) selfUpdateAndFanout(attr Attribute, value interface{}) (error, []SubscriptionReport) {
	// update the value
	source := ClientSource{client: attr.owner}
	b.setAttributeValue(source, attr, value)
	fmt.Println("[SelfUpdate        ]", source, "->", source, "@"+attr.Value().At.Format(time.Stamp), "ATTR:", attr.name, "VALUE:", value)

	// fanout
	return nil, b.fanout(source, attr)
}

func (b *Broker) publish(source Source, name string, value interface{}) (error, []SubscriptionReport) {
	if attr, ok := b.attributes[name]; ok {
		fmt.Println("[Publish           ]", source, "->", attr.owner.name, "@"+attr.Value().At.Format(time.Stamp), "ATTR:", name, "VALUE:", value)
		// validate the value
		var err error
		if attr.Definition != nil {
			if value, err = attr.Transform(value); err != nil {
				return err, nil
			}
		}

		// try to run the accept fns
		for _, fn := range attr.fns {
			fmt.Println("[AcceptFN          ]", source, "->", attr.owner.name, "@"+attr.Value().At.Format(time.Stamp), "ATTR:", attr.name, "VALUE:", value)
			err := fn(source, value)
			if err != nil {
				fmt.Println("[AcceptFN          ] ERROR:", err)
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

func (b *Broker) schedule(schedule cron.Schedule, fn func()) cron.EntryID {
	if b.cron == nil {
		b.cron = cron.New(cron.WithSeconds())
		b.cron.Start()
	}
	return b.cron.Schedule(schedule, cron.FuncJob(fn))
}
