package embedded

import (
	"context"
	"errors"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

type Requirement struct {
	createdID keystone.ID
	uid       string
	lookup    string
}

func (d *Requirement) Name() string {
	return "Embedded Data"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Embedded{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.unique(actor),
		d.find(actor),
		d.update(actor),
		d.reFind(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {

	mdl := &models.Embedded{
		Name: "Values",
		Extended: models.ExtendedData{
			LookupValue: "verify-value-" + time.Now().String(),
			UniqueID:    "UNIQUE-" + k4id.New().UUID(),
			BoolValue:   true,
			Price:       *keystone.NewAmount("USD", 100),
		},
	}

	mdl.ExtendedRef = &mdl.Extended

	d.uid = mdl.Extended.UniqueID
	d.lookup = mdl.Extended.LookupValue

	createErr := actor.Mutate(context.Background(), mdl)
	if createErr == nil {
		d.createdID = mdl.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) unique(actor *keystone.Actor) requirements.TestResult {

	mdl := &models.Embedded{}
	getErr := actor.Get(context.Background(), keystone.ByUniqueProperty(mdl, d.uid, "extended.unique_id"), mdl)

	if getErr == nil {
		if mdl.GetKeystoneID() != d.createdID {
			getErr = errors.New("did not find the correct entity")
		}
	}

	return requirements.TestResult{
		Name:  "Get By Unique",
		Error: getErr,
	}
}

func (d *Requirement) find(actor *keystone.Actor) requirements.TestResult {

	mdl := &models.Embedded{}
	entities, getErr := actor.Find(context.Background(), keystone.Type(mdl), keystone.WithProperties("extended.unique_id", "extended_ref.unique_id", "extended.bool_value"), keystone.WhereEquals("extended.lookup_value", d.lookup))

	if getErr == nil {
		switch len(entities) {
		case 0:
			getErr = errors.New("no entities found")
		case 1:
			_ = keystone.Unmarshal(entities[0], mdl)
			if mdl.GetKeystoneID() != d.createdID {
				getErr = errors.New("did not find the correct entity")
			} else if mdl.Extended.UniqueID != d.uid {
				getErr = errors.New("did not load the extended data")
			} else if mdl.ExtendedRef.UniqueID != d.uid {
				getErr = errors.New("did not load the extended ref data")
			} else if mdl.Extended.BoolValue != true {
				getErr = errors.New("did not find the correct entity (bool)")
			}
		default:
			getErr = errors.New("found too many entities")
		}
	}

	return requirements.TestResult{
		Name:  "Find By Lookup Value",
		Error: getErr,
	}
}

func (d *Requirement) update(actor *keystone.Actor) requirements.TestResult {
	ret := requirements.TestResult{Name: "Update embedded"}

	mdl := &models.Embedded{
		Extended: models.ExtendedData{},
	}
	mdl.SetKeystoneID(d.createdID)

	mdl.Extended.StringValue = "abc"

	updateErr := actor.Mutate(context.Background(), mdl, keystone.MutateProperties("extended.string_value", "extended.bool_value"))
	if updateErr == nil {
		return ret.WithError(updateErr)
	}

	return ret
}

func (d *Requirement) reFind(actor *keystone.Actor) requirements.TestResult {
	ret := requirements.TestResult{Name: "ReLoad Embedded"}

	mdl := &models.Embedded{}
	getErr := actor.GetByID(context.Background(), d.createdID, mdl, keystone.WithProperties("extended.string_value", "extended.bool_value"))
	if getErr != nil {
		return ret.WithError(getErr)
	}

	if mdl.GetKeystoneID() != d.createdID {
		return ret.WithError(errors.New("did not find the correct entity"))
	} else if mdl.Extended.StringValue != "abc" {
		return ret.WithError(errors.New("incorrect string value"))
	} else if mdl.Extended.BoolValue != false {
		return ret.WithError(errors.New("incorrect boolean value"))
	}

	return ret
}
