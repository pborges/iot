package pubsub

import (
	"fmt"
	"testing"
)

func TestBroker_CreateClient(t *testing.T) {
	b := Broker{}

	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("owner%d", i)
		t.Run(fmt.Sprintf("should create owner %s", name), func(t *testing.T) {
			client, err := b.CreateClient(name)
			if err != nil {
				t.Error(err)
			}

			if client.name != name {
				t.Error("unexpected name")
			}

			if b.clients.db == nil {
				t.Fatal("clients map should not be nil")
			}

			if len(b.clients.db) != i {
				t.Fatal("invalid number of clients")
			}

			if b.clients.db[client.name] == nil {
				t.Fatal("owner not present in map")
			}

			if b.clients.db[client.name].name != name {
				t.Fatal("owner did not get the correct name")
			}
			t.Run("should error on duplicate name", func(t *testing.T) {
				_, err := b.CreateClient(name)
				if err != ErrDuplicateClient {
					t.Fatal("expected error on duplicate name didnt get one")
				}
			})
		})
	}
}

func TestBroker_List(t *testing.T) {
	b := Broker{}

	client, err := b.CreateClient("test")
	if err != nil {
		t.Error(err)
	}
	t.Run("test create undefined attribute", func(t *testing.T) {
		attr, err, reports := client.CreateAttribute("one", nil, nil)
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
		attr, err, reports := client.CreateAttribute("one", IntegerDefinition{Default: 1234}, nil)
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
