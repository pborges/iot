package pubsub

type Attribute struct {
	Definition
	name  string
	owner *Client
	fns   []OnAcceptFn
}

func (k Attribute) Cancel() error {
	return k.owner.broker.cancelAttribute(k.name)
}

func (k Attribute) Value() Datum {
	return k.owner.broker.getAttributeValue(k.name)
}

func (k Attribute) Update(value interface{}) (error, []SubscriptionReport) {
	value, err := k.Transform(value)
	if err != nil {
		return err, nil
	}
	return k.owner.broker.selfUpdateAndFanout(k, value)
}
