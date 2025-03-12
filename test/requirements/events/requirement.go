package events

import (
	"context"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID keystone.ID
}

func (d *Requirement) Name() string {
	return "Events"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.User{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.storeEvent(actor),
	}
}

func (d *Requirement) storeEvent(actor *keystone.Actor) requirements.TestResult {

	usr := &models.User{
		Validate: "event-store",
	}
	usr.AddEvent("tst1", map[string]string{"key": "value", "xx": "yy"})

	createErr := actor.Mutate(context.Background(), usr, keystone.WithMutationComment("Create user, with an event"))
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}
	return requirements.TestResult{
		Name:  "Store Event",
		Error: createErr,
	}
}
