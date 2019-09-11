package iot

import (
	"fmt"
	"testing"
)

var broker Broker
var clientOrd int

func getTestClient(t *testing.T) *Client {
	clientOrd++
	client, err := broker.CreateClient(fmt.Sprintf("client%d", clientOrd))
	if err != nil {
		t.Error(err)
	}
	return client
}

func TestClient_PublishAndFanout(t *testing.T) {
	c1 := getTestClient(t)
	c2 := getTestClient(t)
	c3 := getTestClient(t)
	c4 := getTestClient(t)

	var c2Subscribe int64
	var c3Subscribe int64
	var c4Subscribe int64

	if _, err := c2.Subscribe(">", func(name string, value Datum, res BrokerAccess) error {
		c2Subscribe = value.Value.(int64)
		return nil
	}); err != nil {
		t.Error(err)
	}

	if _, err := c3.Subscribe(">", func(name string, value Datum, res BrokerAccess) error {
		c3Subscribe = value.Value.(int64)
		return nil
	}); err != nil {
		t.Error(err)
	}

	if _, err := c4.Subscribe("bob", func(name string, value Datum, res BrokerAccess) error {
		c4Subscribe = value.Value.(int64)
		return nil
	}); err != nil {
		t.Error(err)
	}

	attr, err, reports := c1.Create("temp", IntegerDefinition{Default: 3}, nil)
	if err != nil {
		t.Error(err)
	}
	if len(reports) != 2 {
		t.Error("unexpected number of reports, got: ", len(reports))
	}
	if c2Subscribe != 3 || c3Subscribe != 3 {
		t.Error("expected c2 and c3 to get the fanout")
	}

	if c4Subscribe != 0 {
		t.Error("did not expect c4 to get the fanout")
	}

	c2Subscribe = 0
	c3Subscribe = 0
	c4Subscribe = 0

	err, reports = attr.Update(55)
	if err != nil {
		t.Error(err)
	}
	if len(reports) != 2 {
		t.Error("unexpected number of reports, got: ", len(reports))
	}
	if c2Subscribe != 55 || c3Subscribe != 55 {
		t.Error("expected c2 and c3 to get the fanout")
	}

	if c4Subscribe != 0 {
		t.Error("did not expect c4 to get the fanout")
	}
}
