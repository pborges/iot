package iot

import "testing"

func TestBroker_CreateClient(t *testing.T) {
	b := Broker{}

	t.Run("should create client", func(t *testing.T) {
		client, err := b.CreateClient("test")
		if err != nil {
			t.Error(err)
		}
		if client.name != "test" {
			t.Error("unexpected name")
		}

		if b.clients == nil {
			t.Error("clients map should not be nil")
		}

		if len(b.clients) != 1 {
			t.Error("invalid number of clients")
		}

		if b.clients["test"] == nil {
			t.Error("client not present in map")
		}

		if b.clients["test"].name != "test" {
			t.Error("client did not get the correct name")
		}
	})

	t.Run("should error on duplicate name", func(t *testing.T) {
		_, err := b.CreateClient("test")
		if err != ErrDuplicateName {
			t.Error("expected error on duplicate name didnt ge tone")
		}
	})
}

func TestBroker_List(t *testing.T) {
	b := Broker{}

	client, err := b.CreateClient("test")
	if err != nil {
		t.Error(err)
	}
	t.Run("test create undefined attribute", func(t *testing.T) {
		attr, err, reports := client.Create("one", nil, nil)
		if err != nil {
			t.Error(err)
		}
		if len(reports) != 0 {
			t.Error("expected no reports")
		}
		if attr.name != "test.one" {
			t.Error("unexpected name")
		}

		attrs := b.List(">")
		if len(attrs) != 1 {
			t.Error("unexpected number of results")
		}

		if attrs[0].Name != "test.one" {
			t.Error("unexpected name")
		}

		if attrs[0].Value != nil {
			t.Error("unexpected value")
		}
	})
	t.Run("test create defined attribute", func(t *testing.T) {
		attr, err, reports := client.Create("one", IntegerDefinition{Default: 1234}, nil)
		if err != nil {
			t.Error(err)
		}
		if len(reports) != 0 {
			t.Error("expected no reports")
		}
		if attr.name != "test.one" {
			t.Error("unexpected name")
		}

		attrs := b.List(">")
		if len(attrs) != 1 {
			t.Error("unexpected number of results")
		}
		if i, ok := attrs[0].Value.(int64); !ok || i != 1234 {
			t.Error("unexpected value")
		}
	})
}
