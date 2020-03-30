package pubsub

import (
	"strings"
)

func isOwner(key, publisher string) bool {
	segPublisher := strings.Split(publisher, ".")
	segKey := strings.Split(key, ".")

	if len(segPublisher) <= 0 || len(segKey) <= 0 {
		return false
	}

	return segKey[0] == segPublisher[0]
}

func KeyMatch(attr string, filter string) bool {
	segAttr := strings.Split(attr, ".")
	segFilter := strings.Split(filter, ".")

	if len(segFilter) <= 0 || len(segAttr) <= 0 {
		return false
	}

	if len(segAttr) > len(segFilter) {
		segFilter = append(segFilter, make([]string, len(segAttr)-len(segFilter))...)
	} else {
		segAttr = append(segAttr, make([]string, len(segFilter)-len(segAttr))...)
	}
	for i, f := range segFilter {
		if f == ">" {
			return true
		}
		if f == "*" {
			continue
		}
		if f != segAttr[i] {
			return false
		}
	}
	return true
}
