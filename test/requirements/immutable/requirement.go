package immutable

import (
	"context"
	"errors"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"strconv"
	"strings"
	"time"
)

type Requirement struct {
	createdID string
}

func (d *Requirement) Name() string {
	return "Immutable Entity"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Transaction{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.update(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {

	trans := &models.Transaction{
		Amount:      keystone.NewAmount("GBP", 1023),
		ID:          "trans-ids-" + strconv.Itoa(int(time.Now().UnixMilli())),
		PaymentType: "card",
	}

	createErr := actor.Mutate(context.Background(), trans, keystone.WithMutationComment("New Transaction"))
	if createErr == nil {
		d.createdID = trans.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) update(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Update",
	}

	trans := &models.Transaction{}
	trans.SetKeystoneID(d.createdID)
	trans.PaymentType = "cash"
	updateErr := actor.Mutate(context.Background(), trans, keystone.WithMutationComment("Change payment type"))
	if updateErr == nil || !strings.Contains(updateErr.Error(), "Updates are not permitted") {
		// Allow this error
		res.Error = fmt.Errorf("expected updates to not be permitted, got %s", updateErr)
		return res
	}

	res.Error = actor.Get(context.Background(), keystone.ByEntityID(trans, d.createdID), trans, keystone.WithProperties())
	if res.Error != nil {
		// Return this
	} else if trans.PaymentType != "card" {
		res.Error = errors.New("PaymentType was updated")
	}

	return res
}
