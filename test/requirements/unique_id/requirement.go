package unique_id

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"strconv"
	"time"
)

type Requirement struct {
	createdID keystone.ID
	uniqueID  string
}

func (d *Requirement) Name() string {
	return "Unique ID"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.User{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.store(actor),
		d.retrieve(actor),
	}
}

func (d *Requirement) store(actor *keystone.Actor) requirements.TestResult {

	d.uniqueID = "UNQ-" + strconv.Itoa(int(time.Now().Unix()))
	usr := &models.User{
		ExternalID: d.uniqueID,
		Validate:   "qwertyuiop",
	}

	createErr := actor.Mutate(context.Background(), usr, keystone.WithMutationComment("Create a user"))
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) retrieve(actor *keystone.Actor) requirements.TestResult {
	usr := &models.User{}
	getErr := actor.Get(context.Background(), keystone.ByUniqueProperty(usr, d.uniqueID, "external_id"), usr, keystone.WithProperties("external_id", "validate"))

	if getErr == nil {
		if usr.Validate != "qwertyuiop" {
			getErr = errors.New("unexpected value for 'validate'")
		}
	}

	return requirements.TestResult{
		Name:  "Read",
		Error: getErr,
	}
}
