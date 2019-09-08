package pubsub

import (
	"testing"
)

// this is not a valid test yet
func TestRecordingBroker_Create(t *testing.T) {
	b := RecordingBroker{
		Next: &CoreBroker{},
	}

	b.Create("hi", nil)
}
