package hashed_id

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"time"
)

var (
	Name  = "A"
	Name2 = "B"
)

const Last = "-Last"

type Requirement struct {
	hashedID string
}

func (d *Requirement) Name() string {
	return "Create Read Update - Hashed ID"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Hashi{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	d.hashedID = "userid123"
	return []requirements.TestResult{
		d.create(actor),
		d.read(actor),
		d.update(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create"}
	hashi := &models.Hashi{
		FirstName: Name,
		LastName:  Name + Last,
	}
	idErr := hashi.SetHashID(d.hashedID)
	if idErr != nil {
		return res.WithError(idErr)
	}

	createErr := actor.Mutate(context.Background(), hashi, keystone.WithMutationComment("Create a person"))
	time.Sleep(time.Second * 3)
	return res.WithError(createErr)
}

func (d *Requirement) read(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Read"}
	hashi := &models.Hashi{}
	getErr := actor.Get(context.Background(), keystone.ByHashID(hashi, d.hashedID), hashi, keystone.WithProperties())
	if getErr != nil {
		return res.WithError(getErr)
	}

	if hashi.FirstName != Name {
		return res.WithError(errors.New("first name mismatch"))
	}
	if hashi.LastName != Name+Last {
		return res.WithError(errors.New("last name mismatch"))
	}

	return res
}

func (d *Requirement) update(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Update"}

	hashi := &models.Hashi{}
	_ = hashi.SetHashID(d.hashedID)
	hashi.FirstName = Name2
	hashi.LastName = Name2 + Last
	updateErr := actor.Mutate(context.Background(), hashi, keystone.WithMutationComment("Update a person"), keystone.MutateProperties("first_name", "last_name"))

	if updateErr != nil {
		return res.WithError(updateErr)
	}

	if getErr := actor.Get(context.Background(), keystone.ByHashID(hashi, d.hashedID), hashi, keystone.WithProperties()); getErr != nil {
		return res.WithError(getErr)
	}

	if hashi.FirstName != Name2 {
		return res.WithError(errors.New("first name mismatch"))
	}
	if hashi.LastName != Name2+Last {
		return res.WithError(errors.New("last name mismatch"))
	}

	return res
}
