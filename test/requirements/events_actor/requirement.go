package events_actor

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID keystone.ID
}

func (d *Requirement) Name() string {
	return "Events Actor"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.User{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.storeEvent(actor),
		d.retrieveEvents(actor),
		d.retrieveEventsByType(actor),
	}
}

func (d *Requirement) storeEvent(actor *keystone.Actor) requirements.TestResult {
	usr := &models.User{
		Validate: "events-actor-test",
	}
	usr.AddEvent("test-event-1", map[string]string{"key1": "value1", "key2": "value2"})
	usr.AddEvent("test-event-2", map[string]string{"key3": "value3"})

	createErr := actor.Mutate(context.Background(), usr, keystone.WithMutationComment("Create user with events for Actor.Events test"))
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}
	return requirements.TestResult{
		Name:  "Store Events",
		Error: createErr,
	}
}

func (d *Requirement) retrieveEvents(actor *keystone.Actor) requirements.TestResult {
	if d.createdID == "" {
		return requirements.TestResult{
			Name:  "Retrieve Events",
			Error: errors.New("no entity created"),
		}
	}

	events, err := actor.Events(context.Background(), string(d.createdID))
	if err != nil {
		return requirements.TestResult{
			Name:  "Retrieve Events",
			Error: err,
		}
	}

	if len(events) < 2 {
		return requirements.TestResult{
			Name:  "Retrieve Events",
			Error: errors.New("expected at least 2 events, got fewer"),
		}
	}

	// Verify we got our expected event types
	foundEvent1 := false
	foundEvent2 := false
	for _, event := range events {
		if event.GetType().GetKey() == "test-event-1" {
			foundEvent1 = true
			if event.GetData()["key1"] != "value1" {
				return requirements.TestResult{
					Name:  "Retrieve Events",
					Error: errors.New("event data mismatch for test-event-1"),
				}
			}
		}
		if event.GetType().GetKey() == "test-event-2" {
			foundEvent2 = true
		}
	}

	if !foundEvent1 || !foundEvent2 {
		return requirements.TestResult{
			Name:  "Retrieve Events",
			Error: errors.New("expected events not found"),
		}
	}

	return requirements.TestResult{
		Name:  "Retrieve Events",
		Error: nil,
	}
}

func (d *Requirement) retrieveEventsByType(actor *keystone.Actor) requirements.TestResult {
	if d.createdID == "" {
		return requirements.TestResult{
			Name:  "Retrieve Events By Type",
			Error: errors.New("no entity created"),
		}
	}

	// Filter by event type
	events, err := actor.Events(
		context.Background(),
		string(d.createdID),
		keystone.WithEventTypes(&proto.Key{Key: "test-event-1"}),
	)
	if err != nil {
		return requirements.TestResult{
			Name:  "Retrieve Events By Type",
			Error: err,
		}
	}

	if len(events) == 0 {
		return requirements.TestResult{
			Name:  "Retrieve Events By Type",
			Error: errors.New("expected at least 1 event when filtering by type"),
		}
	}

	// Verify all returned events are of the expected type
	for _, event := range events {
		if event.GetType().GetKey() != "test-event-1" {
			return requirements.TestResult{
				Name:  "Retrieve Events By Type",
				Error: errors.New("received event with unexpected type"),
			}
		}
	}

	return requirements.TestResult{
		Name:  "Retrieve Events By Type",
		Error: nil,
	}
}
