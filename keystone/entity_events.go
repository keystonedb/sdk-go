package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

// EventProvider is an interface for entities that can have events
type EventProvider interface {
	ClearEvents() error
	GetEvents() []*proto.EntityEvent
}

// EmbeddedEvents is a struct that implements EventProvider
type EmbeddedEvents struct {
	ksEntityEvents []*proto.EntityEvent
}

// ClearEvents clears the events
func (e *EmbeddedEvents) ClearEvents() error {
	e.ksEntityEvents = []*proto.EntityEvent{}
	return nil
}

// GetEvents returns the events
func (e *EmbeddedEvents) GetEvents() []*proto.EntityEvent {
	return e.ksEntityEvents
}

// AddEvent adds an event
func (e *EmbeddedEvents) AddEvent(eventType string, properties map[string]string) {
	e.ksEntityEvents = append(e.ksEntityEvents, &proto.EntityEvent{
		Type: &proto.Key{Key: eventType},
		Time: timestamppb.New(time.Now()),
		Data: properties,
	})
}
