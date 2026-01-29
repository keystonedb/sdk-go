package keystone

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/proto"
)

// Events retrieves event entries for an entity
func (a *Actor) Events(ctx context.Context, entityID string, opts ...EventsOption) ([]*proto.EntityEvent, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	options := &eventsOptions{}
	for _, opt := range opts {
		opt(options)
	}

	req := &proto.EventRequest{
		Authorization:  a.Authorization(),
		EntityId:       entityID,
		EventByType:    options.eventTypes,
		EventsInWindow: options.window,
	}

	resp, err := a.connection.Events(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetEvents(), nil
}

type eventsOptions struct {
	eventTypes []*proto.Key
	window     *proto.Window
}

// EventsOption is a functional option for the Events method
type EventsOption func(*eventsOptions)

// WithEventTypes sets the event types to filter by
func WithEventTypes(types ...*proto.Key) EventsOption {
	return func(o *eventsOptions) {
		o.eventTypes = types
	}
}

// WithEventsWindow sets the time window for event retrieval
func WithEventsWindow(window *proto.Window) EventsOption {
	return func(o *eventsOptions) {
		o.window = window
	}
}
