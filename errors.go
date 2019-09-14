package iot

import (
	"fmt"
)

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
