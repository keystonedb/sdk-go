package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func TestActor_Events_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.Events(context.Background(), "entity-123")
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
}

func TestActor_Events_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.Events(context.Background(), "entity-123")
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
}

func TestEventsOptions(t *testing.T) {
	// Test WithEventsWindow
	opts := &eventsOptions{}
	window := &proto.Window{}
	WithEventsWindow(window)(opts)
	if opts.window != window {
		t.Errorf("WithEventsWindow: expected window to be set")
	}

	// Test WithEventTypes with single key
	opts = &eventsOptions{}
	key1 := &proto.Key{Key: "event-type-1"}
	WithEventTypes(key1)(opts)
	if len(opts.eventTypes) != 1 {
		t.Errorf("WithEventTypes: expected 1 type, got %d", len(opts.eventTypes))
	}
	if opts.eventTypes[0].Key != "event-type-1" {
		t.Errorf("WithEventTypes: expected 'event-type-1', got %s", opts.eventTypes[0].Key)
	}

	// Test WithEventTypes with multiple keys
	opts = &eventsOptions{}
	key2 := &proto.Key{Key: "event-type-2"}
	key3 := &proto.Key{Key: "event-type-3"}
	WithEventTypes(key2, key3)(opts)
	if len(opts.eventTypes) != 2 {
		t.Errorf("WithEventTypes: expected 2 types, got %d", len(opts.eventTypes))
	}
	if opts.eventTypes[0].Key != "event-type-2" || opts.eventTypes[1].Key != "event-type-3" {
		t.Errorf("WithEventTypes: expected [event-type-2, event-type-3], got %v", opts.eventTypes)
	}

	// Test combining options
	opts = &eventsOptions{}
	window = &proto.Window{}
	key := &proto.Key{Key: "combined-type"}
	WithEventsWindow(window)(opts)
	WithEventTypes(key)(opts)
	if opts.window != window {
		t.Errorf("Combined options: expected window to be set")
	}
	if len(opts.eventTypes) != 1 || opts.eventTypes[0].Key != "combined-type" {
		t.Errorf("Combined options: expected event types to be set")
	}
}
