package keystone

import (
	"context"
	"testing"
)

func TestActor_Lookup_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.Lookup(context.Background(), "email", "test@example.com")
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
}

func TestActor_Lookup_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.Lookup(context.Background(), "email", "test@example.com")
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
}

func TestActor_LookupOne_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.LookupOne(context.Background(), "email", "test@example.com")
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
}

func TestActor_LookupOne_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.LookupOne(context.Background(), "email", "test@example.com")
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
}

func TestLookupOptions(t *testing.T) {
	// Test WithLookupSchemeID
	opts := &lookupOptions{}
	WithLookupSchemeID("user")(opts)
	if opts.schemeID != "user" {
		t.Errorf("WithLookupSchemeID: expected 'user', got %s", opts.schemeID)
	}
}
