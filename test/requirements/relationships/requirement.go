package relationships

import (
	"context"
	"errors"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"strconv"
	"time"
)

type Requirement struct {
	PersonID       string
	TransactionID  string
	Transaction2ID string
}

func (d *Requirement) Name() string {
	return "Relationships"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Person{})
	conn.RegisterTypes(models.Transaction{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.count(actor),
		d.lookup(actor),
		d.loadRelations(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Create Entities",
	}

	psn := &models.Person{
		Name: "Mr Transaction",
	}

	psnErr := actor.Mutate(context.Background(), psn, keystone.WithMutationComment("Create Person"))
	if psnErr != nil {
		res.Error = psnErr
		return res
	}
	d.PersonID = psn.GetKeystoneID()

	trans := &models.Transaction{
		Amount:      keystone.NewAmount("GBP", 1023),
		ID:          "relid-" + strconv.Itoa(int(time.Now().UnixMilli())),
		PaymentType: "card",
	}
	trans.AddRelationship("payee", psn.GetKeystoneID(), nil, time.Now())
	t1Err := actor.Mutate(context.Background(), trans, keystone.WithMutationComment("Create Transaction 1"))
	if t1Err != nil {
		res.Error = t1Err
		return res
	}
	d.TransactionID = trans.GetKeystoneID()

	trans2 := &models.Transaction{
		Amount:      keystone.NewAmount("GBP", 1023),
		ID:          "relid2-" + strconv.Itoa(int(time.Now().UnixMilli())),
		PaymentType: "card",
	}
	trans2.AddRelationship("payee", psn.GetKeystoneID(), nil, time.Now())
	t2Err := actor.Mutate(context.Background(), trans2, keystone.WithMutationComment("Create Transaction 2"))
	if t2Err != nil {
		res.Error = t2Err
		return res
	}
	d.Transaction2ID = trans2.GetKeystoneID()

	psn.AddRelationship("payment", trans.GetKeystoneID(), map[string]string{"initial": "true"}, time.Now())
	psn.AddRelationship("payment", trans2.GetKeystoneID(), nil, time.Now())
	updatePsn := actor.Mutate(context.Background(), psn, keystone.WithMutationComment("Create Person"))
	if updatePsn != nil {
		res.Error = updatePsn
		return res
	}

	return res
}

func (d *Requirement) count(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Count",
	}

	psn := &models.Person{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(psn, d.PersonID), psn, keystone.RetrieveOptions(
		keystone.WithProperties(),
		keystone.WithSiblingRelationshipCount("payment"),
	))
	if getErr != nil {
		res.Error = getErr
		return res
	}

	if psn.Name != "Mr Transaction" {
		res.Error = errors.New("person name is not correct")
		return res
	}

	if psn.PaymentCount != 2 {
		res.Error = errors.New("transaction count is not 2")
		return res
	}

	return res
}

func (d *Requirement) lookup(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Lookup By Relationship",
	}

	transactions, getErr := actor.Find(context.Background(), keystone.Type(&models.Transaction{}), keystone.WithProperties(), keystone.RelationToSibling(d.PersonID, "payee"))
	if getErr != nil {
		res.Error = getErr
		return res
	}

	if transactions == nil {
		res.Error = errors.New("no transactions found")
		return res
	}

	if len(transactions) != 2 {
		res.Error = errors.New("transaction count is not 2")
		return res
	}

	l1, l2 := false, false
	for _, t := range transactions {
		if t.GetEntity().GetEntityId() == d.TransactionID {
			l1 = true
		} else if t.GetEntity().GetEntityId() == d.Transaction2ID {
			l2 = true
		}
	}

	if !l1 {
		res.Error = fmt.Errorf("transaction 1 (%s) not found", d.TransactionID)
		return res
	}

	if !l2 {
		res.Error = fmt.Errorf("transaction 2 (%s) not found", d.Transaction2ID)
		return res
	}

	return res
}

func (d *Requirement) loadRelations(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Load Relations with Data",
	}

	psn := &models.Person{}
	getErr := actor.GetByID(context.Background(), d.PersonID, psn, keystone.RetrieveOptions(keystone.WithRelationships("payment")))
	if getErr != nil {
		res.Error = getErr
		return res
	}

	if len(psn.GetRelationships()) != 2 {
		res.Error = errors.New("transaction count is not 2")
		return res
	}

	for _, rel := range psn.GetRelationships() {
		if rel.GetRelationship().GetKey() == "payment" {
			if rel.GetTargetId() == d.TransactionID {
				if rel.GetData()["initial"] != "true" {
					res.Error = errors.New("initial payment data not found")
					return res
				}
			}
		}
	}

	return res
}
