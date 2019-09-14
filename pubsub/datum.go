package pubsub

import (
	"time"
)

type Datum struct {
	Owner string
	Name  string
	Def   Definition
	Value interface{}
	By    Source
	At    time.Time
}
