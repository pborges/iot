package pubsub

import (
	"strings"
)

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
