package pubsub

import (
	"testing"
)

func TestKeyMatch(t *testing.T) {
	type KF struct {
		Key    string
		Filter string
		Result bool
	}

	tests := []KF{
		{
			Key:    "debug.n1.gpio.0",
			Filter: "debug.n1.gpio.*",
			Result: true,
		},
		{
			Key:    "debug.n1.gpio.0",
			Filter: "debug.n1.gpio",
			Result: false,
		},
		{
			Key:    "debug.n1.gpio.0",
			Filter: "debug.n1.>",
			Result: true,
		},
		{
			Key:    "debug.n1.gpio.0",
			Filter: "debug.n1.*",
			Result: false,
		},
		{
			Key:    "debug.n1.gpio.0",
			Filter: "debug.n1.gpio.0",
			Result: true,
		},
		{
			Key:    "debug.n1.gpio.1",
			Filter: "debug.n1.gpio.0",
			Result: false,
		},
		{
			Key:    "debug.n1.gpio.1",
			Filter: "*",
			Result: false,
		},
		{
			Key:    "debug.n1.gpio.1",
			Filter: ">",
			Result: true,
		},
		{
			Key:    "debug.n1.data.1",
			Filter: "debug.*.gpio.*",
			Result: false,
		},
		{
			Key:    "debug.n1.gpio.1",
			Filter: "debug.*.gpio.1",
			Result: true,
		},
		{
			Key:    "debug.n1.gpio.1",
			Filter: "debug.n1.data.temp.1",
			Result: false,
		},
	}
	for _, kf := range tests {
		if KeyMatch(kf.Key, kf.Filter) != kf.Result {
			t.Errorf("key:%s filter:%s expected:%t", kf.Key, kf.Filter, kf.Result)
		}
	}
}
