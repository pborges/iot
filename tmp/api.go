package tmp

import "time"

// This package contains some code "doodles" not intended to be production code

type Broker struct {
}

func (b *Broker) Publish(interface{}) {
	//do publish
}

func (b *Broker) Subscribe(fn func(string, interface{})) Subscription {
	//do subscriptions
	return Subscription{}
}

type Subscription struct {
}

// Audited Broker

type AuditedBroker struct {
	OwnerName string
	next      Broker
}

func (b *AuditedBroker) Publish(value interface{}) {
	b.next.Publish(AuditedValue{
		OwnerName: b.OwnerName,
		Value:     value,
	})
}

func (b *AuditedBroker) Subscribe(fn func(AuditedMetadata, interface{})) AuditedSubscription {
	sub := b.next.Subscribe(func(key string, value interface{}) {
		fn(AuditedMetadata{
			Key:       key,
			Timestamp: time.Now(),
			OwnerName: b.OwnerName,
		}, value)
	})
	return AuditedSubscription{
		Subscription: sub,
	}
}

type AuditedSubscription struct {
	Subscription
}

type AuditedValue struct {
	OwnerName string
	Value     interface{}
}

type AuditedMetadata struct {
	Key       string
	Timestamp time.Time
	OwnerName string
}
