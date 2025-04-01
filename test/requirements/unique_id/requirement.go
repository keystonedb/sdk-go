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
	validate  string
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
		d.replace(actor),
		d.retrieve(actor),
		d.noUpdate(actor),
		d.retrieve(actor),
	}
}

func (d *Requirement) store(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create"}

	d.uniqueID = "UNQ-" + strconv.Itoa(int(time.Now().Unix()))
	d.validate = "qwertyuiop"
	usr := &models.User{
		ExternalID: d.uniqueID,
		Validate:   d.validate,
	}

	createErr := actor.Mutate(context.Background(), usr, keystone.WithMutationComment("Create a user"))
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}

	return res.WithError(createErr)
}

func (d *Requirement) retrieve(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Read"}
	usr := &models.User{}
	getErr := actor.Get(context.Background(), keystone.ByUniqueProperty(usr, d.uniqueID, "external_id"), usr, keystone.WithProperties("external_id", "validate"))

	if getErr == nil {
		if usr.Validate != d.validate {
			getErr = errors.New("unexpected value for 'validate'")
		}
	}

	return res.WithError(getErr)
}

func (d *Requirement) replace(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Replace Values"}

	d.validate = "secondChangereplace"
	usr := &models.User{
		ExternalID: d.uniqueID,
		Validate:   d.validate,
	}

	createErr := actor.Mutate(context.Background(), usr, keystone.WithMutationComment("Replace user values"), keystone.OnConflictUseID("external_id"))
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}

	return res.WithError(createErr)
}

func (d *Requirement) noUpdate(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Replace Values"}

	usr := &models.User{
		ExternalID: d.uniqueID,
		Validate:   "This should have no impact",
	}

	createErr := actor.Mutate(context.Background(), usr, keystone.WithMutationComment("No mutation happened"), keystone.OnConflictIgnore())
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}

	return res.WithError(createErr)
}
