package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

type mockEntityCard struct {
	Brand  string `keystone:",query"`
	Issuer string
}

type mockEntityToken struct {
	BaseEntity
	Status  string `keystone:",query"`
	Card    *mockEntityCard
	Comment string
}

func TestMockEntityRetrievesAsAStoredEntity(t *testing.T) {
	conn, mock, listener, server := MockConnection()
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	mock.DefineFunc = func(context.Context, *proto.SchemaRequest) (*proto.Schema, error) { return &proto.Schema{}, nil }
	actor := conn.Actor("ws", "127.0.0.1", "user", "go-test")

	stored := &mockEntityToken{Status: "active", Card: &mockEntityCard{Brand: "VISA", Issuer: "Bank"}, Comment: "untagged"}
	stored.SetKeystoneID("inst1-tok1")
	mock.RetrieveFunc = func(context.Context, *proto.EntityRequest) (*proto.EntityResponse, error) {
		return MockEntity(stored)
	}

	got := &mockEntityToken{}
	if err := actor.GetByID(context.Background(), "inst1-tok1", got, WithDecryptedProperties()); err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !got.ExistsInKeystone() {
		t.Error("ExistsInKeystone = false")
	}
	if !got.IsStateActive() {
		t.Error("IsStateActive = false")
	}
	if got.GetKeystoneID() != "inst1-tok1" {
		t.Errorf("ID = %q", got.GetKeystoneID())
	}
	if got.Status != "active" || got.Comment != "untagged" || got.Card == nil || got.Card.Brand != "VISA" || got.Card.Issuer != "Bank" {
		t.Errorf("entity = %+v, card %+v; want the stored values", got, got.Card)
	}
}

func TestMockEntityFindsAsAStoredEntity(t *testing.T) {
	conn, mock, listener, server := MockConnection()
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()
	mock.DefineFunc = func(context.Context, *proto.SchemaRequest) (*proto.Schema, error) { return &proto.Schema{}, nil }
	actor := conn.Actor("ws", "127.0.0.1", "user", "go-test")

	stored := &mockEntityToken{Status: "active"}
	stored.SetKeystoneID("inst1-tok1")
	mock.FindFunc = func(context.Context, *proto.FindRequest) (*proto.FindResponse, error) {
		e, err := MockEntity(stored)
		return &proto.FindResponse{Entities: []*proto.EntityResponse{e}}, err
	}

	results, err := actor.Find(context.Background(), Type(mockEntityToken{}), WithProperties("status"))
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	found, err := AsSlice[*mockEntityToken](results...)
	if err != nil || len(found) != 1 || found[0].Status != "active" || !found[0].IsStateActive() {
		t.Errorf("found = %+v, err = %v; want the stored entity, active", found, err)
	}
}
