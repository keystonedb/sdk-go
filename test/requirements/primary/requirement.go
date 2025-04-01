package primary

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
	"time"
)

type Requirement struct {
	createdID   keystone.ID
	johnDoeHash string
	lastName    string
}

func (d *Requirement) Name() string {
	return "Primary Key"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.WithPrimary{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	d.lastName = "Doe-" + time.Now().Format(time.DateTime)
	return []requirements.TestResult{
		d.storeNoPrimary(actor),
		d.store(actor),
		d.retrieve(actor),
		d.mustNotChange(actor),
		d.retrieve(actor),
		d.overwriteNoOp(actor),
		d.retrieve(actor),
		d.bench(actor),
	}
}

func (d *Requirement) storeNoPrimary(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create no primary"}

	usr := &models.WithPrimary{}
	usr.FirstName = "John"
	usr.LastName = d.lastName

	createErr := actor.Mutate(context.Background(), usr)
	if createErr == nil {
		return res.WithError(errors.New("primary key not set expects error"))
	}

	return res
}

func (d *Requirement) store(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create"}

	usr := &models.WithPrimary{}
	usr.FirstName = "John"
	usr.LastName = d.lastName
	usr.SetHash()
	d.johnDoeHash = usr.NameHash

	createErr := actor.Mutate(context.Background(), usr)
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}

	return res.WithError(createErr)
}

func (d *Requirement) retrieve(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Read"}
	usr := &models.WithPrimary{}
	getErr := actor.Get(context.Background(), keystone.ByUniqueProperty(usr, d.johnDoeHash, "name_hash"), usr, keystone.WithProperties())

	if getErr == nil {
		if usr.FirstName != "John" {
			getErr = errors.New("unexpected value for 'first name'")
		} else if usr.LastName != d.lastName {
			getErr = errors.New("unexpected value for 'last name'")
		}
	}

	return res.WithError(getErr)
}

func (d *Requirement) overwriteNoOp(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Overwrite, noOp"}

	usr := &models.WithPrimary{}
	usr.FirstName = "John"
	usr.LastName = d.lastName + "Replaced"
	usr.SetHash()
	usr.NameHash = d.johnDoeHash

	createErr := actor.Mutate(context.Background(), usr, keystone.OnConflictIgnore())
	if createErr == nil {
		if d.createdID != usr.GetKeystoneID() {
			createErr = errors.New("expected original created ID to be returned, got " + usr.GetKeystoneID().String())
		}
	}
	return res.WithError(createErr)
}

func (d *Requirement) mustNotChange(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Primary Key must not change"}

	usr := &models.WithPrimary{}
	usr.SetKeystoneID(d.createdID)
	usr.FirstName = "John"
	usr.LastName = d.lastName
	usr.NameHash = k4id.New().String() // Random ID

	createErr := actor.Mutate(context.Background(), usr)
	if createErr == nil {
		return res.WithError(errors.New("no error returned"))
	}
	return res
}

func (d *Requirement) bench(actor *keystone.Actor) requirements.TestResult {
	start := time.Now()
	for i := 0; i < 1000; i++ {
		d.overwriteNoOp(actor)
	}
	end := time.Since(start)
	return requirements.TestResult{Name: "Bench - " + end.String()}
}
