package embedded

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
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {

	mdl := &models.Embedded{
		Name: "Values",
		Extended: models.ExtendedData{
			LookupValue: "verify-value-" + time.Now().String(),
			UniqueID:    "UNIQUE-" + k4id.New().UUID(),
			Price:       keystone.NewAmount("USD", 100),
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
	entities, getErr := actor.Find(context.Background(), keystone.Type(mdl), keystone.WithProperties("extended.unique_id", "extended_ref.unique_id"), keystone.WhereEquals("extended.lookup_value", d.lookup))

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
