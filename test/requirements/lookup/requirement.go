package lookup

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
	createdID string
	lookupID  string
}

func (d *Requirement) Name() string {
	return "Lookup Entity"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Transaction{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.lookup(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {

	d.lookupID = "lutid-" + strconv.Itoa(int(time.Now().UnixMilli()))
	trans := &models.Transaction{
		Amount:      keystone.NewAmount("GBP", 1023),
		ID:          d.lookupID,
		PaymentType: "card",
	}

	createErr := actor.Mutate(context.Background(), trans, "Lookup Transaction")
	if createErr == nil {
		d.createdID = trans.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) lookup(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Lookup",
	}

	trans := &models.Transaction{}
	results, findErr := actor.Find(context.Background(), keystone.Type(trans), keystone.WithProperties("ID", "PaymentType"), keystone.WhereEquals("id", d.lookupID))
	if findErr != nil {
		res.Error = findErr
		return res
	}

	if results == nil {
		res.Error = errors.New("no results found")
		return res
	}

	located := false
	for _, result := range results {
		if result.GetEntity().GetEntityId() == d.createdID {
			located = true
			break
		}
	}

	if !located {
		res.Error = errors.New("lookup failed")
	}

	return res
}
