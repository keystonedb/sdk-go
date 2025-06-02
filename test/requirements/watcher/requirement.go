package watcher

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	eid     keystone.ID
	updated bool
}

func (d *Requirement) Name() string {
	return "Watchers"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.read(actor),
		d.readAndUpdate(actor),
		d.read(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create"}

	dt := &models.DataTypes{
		AmountPt: keystone.NewAmount("USD", 100),
		Secret:   keystone.NewSecureString("Original", "MASKED"),
		Boolean:  true,
		Float:    10.99,
	}
	err := actor.Mutate(context.Background(), dt, keystone.WithMutationComment("Create"))
	if err != nil {
		return res.WithError(err)
	}

	d.eid = dt.GetKeystoneID()

	return res
}

func (d *Requirement) read(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Read"}
	dt := &models.DataTypes{}
	err := actor.GetByID(context.Background(), d.eid, dt, keystone.WithDecryptedProperties())
	if err != nil {
		return res.WithError(err)
	}

	if dt.AmountPt == nil {
		return res.WithError(errors.New("AmountPt is nil"))
	}
	if dt.AmountPt.GetCurrency() != "USD" {
		return res.WithError(errors.New("AmountPt currency is not USD"))
	}
	if dt.AmountPt.GetUnits() != 100 {
		return res.WithError(errors.New("AmountPt amount is not 100"))
	}

	if dt.Boolean != true {
		return res.WithError(errors.New("boolean is not true"))
	}

	if dt.Float != 10.99 {
		return res.WithError(errors.New("float is not 10.99"))
	}

	if dt.Secret.Masked != "MASKED" {
		return res.WithError(errors.New("Secret.Masked is not MASKED"))
	}

	if dt.Secret.Original != "Original" {
		return res.WithError(errors.New("Secret.Original is not Original"))
	}

	if d.updated && dt.String != "Updated" {
		return res.WithError(errors.New("string is not Updated"))
	}

	return res
}

func (d *Requirement) readAndUpdate(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Read and Update"}

	dt := &models.DataTypes{}
	err := actor.GetByID(context.Background(), d.eid, dt, keystone.WithProperties("string"))
	if err != nil {
		return res.WithError(err)
	}

	dt.String = "Updated"

	mutateErr := actor.Mutate(context.Background(), dt, keystone.WithMutationComment("Update"))
	if mutateErr != nil {
		return res.WithError(mutateErr)
	}
	d.updated = true
	return res
}
