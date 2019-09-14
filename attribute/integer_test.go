package attribute

import (
	"github.com/pborges/iot/pubsub"
	"testing"
)

func TestIntegerDefinition_Extract(t *testing.T) {
	i := IntegerDefinition{
		Min: 3,
		Max: 10,
	}

	i2 := IntegerDefinition{
		Min: 3,
		Max: 10,
	}

	i3 := IntegerDefinition{}

	var val int64 = 6

	d := pubsub.Datum{
		Def:   i,
		Value: val,
	}

	if v, err := i.Extract(d); v != val || err != nil {
		t.Fatalf("unexpected value err: %s", err)
	}

	if v, err := i2.Extract(d); v != val || err != nil {
		t.Fatalf("unexpected value err: %s", err)
	}

	if v, err := i3.Extract(d); v != i3.DefaultValue() || err == nil {
		t.Fatalf("unexpected value err: %s", err)
	}

}
