package lookup_actor

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID keystone.ID
	lookupID  string
}

func (d *Requirement) Name() string {
	return "Actor Lookup Methods"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Transaction{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.lookup(actor),
		d.lookupOne(actor),
		d.lookupNotFound(actor),
		d.lookupOneNotFound(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	d.lookupID = "lookup-actor-" + strconv.Itoa(int(time.Now().UnixMilli()))
	trans := &models.Transaction{
		Amount:      *keystone.NewAmount("GBP", 2500),
		ID:          d.lookupID,
		PaymentType: "wire",
	}

	createErr := actor.Mutate(context.Background(), trans, keystone.WithMutationComment("Create transaction for Actor.Lookup test"))
	if createErr == nil {
		d.createdID = trans.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create Entity",
		Error: createErr,
	}
}

func (d *Requirement) lookup(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Actor.Lookup",
	}

	results, err := actor.Lookup(context.Background(), "id", d.lookupID)
	if err != nil {
		res.Error = err
		return res
	}

	if results == nil || len(results) == 0 {
		res.Error = errors.New("no results found from Lookup")
		return res
	}

	found := false
	for _, result := range results {
		if d.createdID.Matches(result.GetEntityId()) {
			found = true
			break
		}
	}

	if !found {
		res.Error = errors.New("created entity not found in Lookup results")
	}

	return res
}

func (d *Requirement) lookupOne(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Actor.LookupOne",
	}

	result, err := actor.LookupOne(context.Background(), "id", d.lookupID)
	if err != nil {
		res.Error = err
		return res
	}

	if result == nil {
		res.Error = errors.New("no result found from LookupOne")
		return res
	}

	if !d.createdID.Matches(result.GetEntityId()) {
		res.Error = errors.New("LookupOne returned wrong entity")
	}

	return res
}

func (d *Requirement) lookupNotFound(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Actor.Lookup Not Found",
	}

	results, err := actor.Lookup(context.Background(), "id", "nonexistent-lookup-id-"+strconv.Itoa(int(time.Now().UnixNano())))
	if err != nil {
		res.Error = err
		return res
	}

	if len(results) != 0 {
		res.Error = errors.New("expected no results for nonexistent lookup value")
	}

	return res
}

func (d *Requirement) lookupOneNotFound(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Actor.LookupOne Not Found",
	}

	result, err := actor.LookupOne(context.Background(), "id", "nonexistent-lookup-id-"+strconv.Itoa(int(time.Now().UnixNano())))
	if err != nil {
		res.Error = err
		return res
	}

	if result != nil {
		res.Error = errors.New("expected nil result for nonexistent lookup value")
	}

	return res
}
