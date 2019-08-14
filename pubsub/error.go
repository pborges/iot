package pubsub

import "errors"

var ErrorKeyAlreadyDefined = errors.New("key already exists by this name")
var ErrorKeyNotFound = errors.New("no key exists by this name")
var ErrorSubscriptionNotFound = errors.New("no subscription exists by this id")
var ErrorSubscriptionAlreadyCanceled = errors.New("subscription has already been canceled")
