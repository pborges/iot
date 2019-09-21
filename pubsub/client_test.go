package pubsub

import (
	"fmt"
	"testing"
)

func getTestClient(t *testing.T, broker *Broker, ord int) *Client {
	client, err := broker.CreateClient(fmt.Sprintf("owner%d", ord))
	if err != nil {
		t.Error(err)
	}
	return client
}

func TestClient_PublishAndFanout(t *testing.T) {
	var broker Broker
	c1 := getTestClient(t, &broker, 1)
	c2 := getTestClient(t, &broker, 2)
	c3 := getTestClient(t, &broker, 3)
	c4 := getTestClient(t, &broker, 4)

	var c2Subscribe int64
	var c3Subscribe int64
	var c4Subscribe int64

	if _, err := c2.Subscribe(">", func(name string, value Datum, res Context) error {
		c2Subscribe = value.Value.(int64)
		return nil
	}); err != nil {
		t.Error(err)
	}

	if _, err := c3.Subscribe(">", func(name string, value Datum, res Context) error {
		c3Subscribe = value.Value.(int64)
		return nil
	}); err != nil {
		t.Error(err)
	}

	if _, err := c4.Subscribe("bob", func(name string, value Datum, res Context) error {
		c4Subscribe = value.Value.(int64)
		return nil
	}); err != nil {
		t.Error(err)
	}

	attr, err, reports := c1.CreateAttribute("temp", IntegerDefinition{Default: 3}, nil)
	if err != nil {
		t.Error(err)
	}
	if len(reports) != 2 {
		t.Error("unexpected number of reports, got: ", len(reports))
	}
	if c2Subscribe != 3 || c3Subscribe != 3 {
		t.Errorf("expected c2 and c3 to get the fanout, but their values are still %d and %d", c2Subscribe, c3Subscribe)
	}

	if c4Subscribe != 0 {
		t.Fatalf("did not expect c4 to get the fanout, but its value changed too %d", c4Subscribe)
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
		t.Errorf("expected c2 and c3 to get the fanout, but their values are still %d and %d", c2Subscribe, c3Subscribe)
	}

	if c4Subscribe != 0 {
		t.Error("did not expect c4 to get the fanout")
	}
}

func TestClient_ShouldNotGetOwnMessage(t *testing.T) {
	var broker Broker
	c1 := getTestClient(t, &broker, 1)
	var c1sub Datum
	c2 := getTestClient(t, &broker, 2)
	var c2sub Datum
	c3 := getTestClient(t, &broker, 3)
	var c3sub Datum

	_, err := c1.Subscribe(">", func(name string, value Datum, b Context) error {
		c1sub = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c2.Subscribe(">", func(name string, value Datum, b Context) error {
		c2sub = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c3.Subscribe(">", func(name string, value Datum, b Context) error {
		c3sub = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	def := IntegerDefinition{Default: 55}
	_, err, reports := c1.CreateAttribute("test", def, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(reports) != 2 {
		t.Errorf("expected 2 reports got %d", len(reports))
		for _, r := range reports {
			t.Errorf("\tid:%s err:%s\n", r.String(), r.Error)
		}
	}

	if c1sub.Value != nil {
		t.Fatalf("c1 should not have gotten the fanout, but its value is %d", c1sub.Value)
	}

	if c2sub.Value.(int64) != 55 {
		t.Fatalf("c2 should have gotten the fanout, but its value is %d", c2sub.Value)
	}

	if c3sub.Value.(int64) != 55 {
		t.Fatalf("c3 should have gotten the fanout, but its value is %d", c3sub.Value)
	}

	err, reports = c2.Publish("owner1.test", 60)
	if err != nil {
		t.Fatal(err)
	}

	if len(reports) != 2 {
		t.Fatalf("expected 2 reports got %d", len(reports))
	}

	if c1sub.Value != nil {
		t.Fatalf("c1 should not have gotten the fanout, but its value is %d", c1sub.Value)
	}

	if c2sub.Value.(int64) != 60 {
		t.Fatalf("c2 should have gotten the fanout, but its value is %d", c2sub.Value)
	}

	if c3sub.Value.(int64) != 60 {
		t.Fatalf("c3 should have gotten the fanout, but its value is %d", c3sub.Value)
	}
}
