package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type IntValue struct {
	Value int
}

func (v IntValue) Inspect() string {
	return strconv.Itoa(v.Value)
}
func (v IntValue) Interface() interface{} {
	return v
}

type IntValidator struct {
	Min      int
	Max      int
	Unsigned bool
}

func (v IntValidator) Validate(val Value) error {
	if casted, ok := val.(IntValue); ok {
		if v.Min != 0 || v.Max != 0 {
			if casted.Value < v.Min {
				return errors.New("value is too small")
			}
			if casted.Value > v.Max {
				return errors.New("value is too small")
			}
		}
		if v.Unsigned && casted.Value < 0 {
			return errors.New("value cannot be signed")
		}
		return nil
	}
	return errors.New("unknown value type")
}

type Value interface {
	Interface() interface{}
	Inspect() string
}

type ValueValidator interface {
	Validate(Value) error
}

type ValueRecorder interface {
	RecordAt(id FullyQualifiedEntity, ts time.Time, v Value) (ValueRecord, error)
}

type ValueRetriever interface {
	ValueAt(id FullyQualifiedEntity, ts time.Time) (rec ValueRecord, err error)
	ValuesBetween(id FullyQualifiedEntity, start, end time.Time) (recs []ValueRecord, err error)
}

type FullyQualifiedEntity interface {
	FullyQualifiedIdentifier() string
}

type Attribute struct {
	Owner     string
	Name      string
	Validator ValueValidator
}

func (a Attribute) FullyQualifiedIdentifier() string {
	return fmt.Sprintf("%s.%s", a.Owner, a.Name)
}

type ValueRecord struct {
	ID string
	Value
	Timestamp time.Time
}

func (r ValueRecord) FullyQualifiedIdentifier() string {
	return r.ID
}

type Server struct {
	ValueRecorder
	entities      []FullyQualifiedEntity
	subscriptions map[string]Subscription
}
type Subscription struct {
	Filter string
	Fn     SubscriptionFn
}
type SubscriptionFn func(rec ValueRecord) error

func (s *Server) Subscribe(id FullyQualifiedEntity, filter string, fn SubscriptionFn) {
	if s.subscriptions == nil {
		s.subscriptions = make(map[string]Subscription)
	}
	s.subscriptions[id.FullyQualifiedIdentifier()] = Subscription{
		Filter: filter,
		Fn:     fn,
	}
}

func (s *Server) Publish(id FullyQualifiedEntity, value Value) (error, []error) {
	if rec, err := s.RecordAt(id, time.Now(), value); err == nil {
		var pubErrors []error
		for _, sub := range s.subscriptions {
			if KeyMatch(sub.Filter, id.FullyQualifiedIdentifier()) {
				if err := sub.Fn(rec); err != nil {
					pubErrors = append(pubErrors, err)
				}
			}
		}
		return nil, pubErrors
	} else {
		return err, nil
	}
}

func (s *Server) RecordAt(id FullyQualifiedEntity, ts time.Time, v Value) (ValueRecord, error) {
	for _, e := range s.entities {
		if e.FullyQualifiedIdentifier() == id.FullyQualifiedIdentifier() {
			switch entity := e.(type) {
			case Attribute:
				if err := entity.Validator.Validate(v); err != nil {
					return ValueRecord{}, err
				}
				return s.ValueRecorder.RecordAt(id, ts, v)
			}
			return ValueRecord{}, errors.New("unknown entity type")
		}
	}
	return ValueRecord{}, errors.New("entity does not exist")
}

func (s *Server) RegisterEntity(id FullyQualifiedEntity) error {
	for _, e := range s.entities {
		if e.FullyQualifiedIdentifier() == id.FullyQualifiedIdentifier() {
			return errors.New("entity already exists")
		}
	}
	s.entities = append(s.entities, id)
	return nil
}

func KeyMatch(key, filter string) bool {
	segFilter := strings.Split(filter, ".")
	segKey := strings.Split(key, ".")

	if len(segKey) > len(segFilter) {
		segFilter = append(segFilter, make([]string, len(segKey)-len(segFilter))...)
	} else {
		segKey = append(segKey, make([]string, len(segFilter)-len(segKey))...)
	}

	for i, f := range segFilter {
		if f == ">" {
			return true
		}
		if f != "*" && f != segKey[i] {
			return false
		}
	}
	return true
}

type DumbRecorder struct {
}

func (r DumbRecorder) RecordAt(id FullyQualifiedEntity, ts time.Time, v Value) (ValueRecord, error) {
	return ValueRecord{}, nil
}

func main() {
	s := Server{
		ValueRecorder: DumbRecorder{},
	}

	a1 := Attribute{
		Owner: "jim",
		Name:  "temperature",
		Validator: IntValidator{
			Min:      0,
			Max:      0,
			Unsigned: false,
		},
	}
	fmt.Println(s.RegisterEntity(a1))

	s.Subscribe(a1, ">", func(rec ValueRecord) error {
		fmt.Println("SUB", rec.Inspect())
		return nil
	})

	fmt.Println(s.Publish(a1, IntValue{Value: 8}))
}
