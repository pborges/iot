package iot

import (
	"fmt"
)

var ErrDuplicateClient = ErrDuplicate("client")
var ErrClientNotFound = ErrNotFound("client")

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
