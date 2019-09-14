package pubsub

import (
	"errors"
	"fmt"
)

var ErrMismatchedDefinition = errors.New("mismatched definition")
var ErrUnknownType = errors.New("unknown type")

var ErrDuplicateClient = ErrDuplicate("owner")
var ErrClientNotFound = ErrNotFound("owner")

var ErrDuplicateSubscription = ErrDuplicate("subscription")
var ErrSubscriptionNotFound = ErrNotFound("subscription")

type ErrDuplicate string

func (e ErrDuplicate) Error() string {
	return fmt.Sprintf("duplicate %s", e)
}

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s not found", e)
}