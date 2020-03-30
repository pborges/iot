package pubsub

import "time"

type Value struct {
	AttributeID string
	Value       interface{}
	UpdatedBy   string
	UpdatedAt   time.Time
	inspected   string
}

func (v Value) Inspect() string {
	return v.inspected
}

type ValueRecord struct {
	RecordId int
	Value
	SubscriptionResponses []SubscriptionResponse
}

type SubscriptionResponse struct {
	SubscriptionID string
	Err            []error
}
