package exists

import (
	"context"
	"errors"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID keystone.ID
}

func (d *Requirement) Name() string {
	return "Exists Check"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.write(actor),
		d.exists(actor),
		d.notExists(actor),
	}
}

func (d *Requirement) write(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Write"}
	u := models.Subscription{
		StartDate: time.Now(),
	}
	err := actor.Mutate(context.Background(), &u)
	d.createdID = u.GetKeystoneID()
	return res.WithError(err)
}

func (d *Requirement) exists(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Exists"}

	sub := models.Subscription{}
	err := actor.GetByID(context.Background(), d.createdID, &sub, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}

	if !sub.ExistsInKeystone() {
		return res.WithError(errors.New("entity not stored in keystone"))
	}

	return res
}

func (d *Requirement) notExists(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Not Exists"}
	sub := models.Subscription{}
	err := actor.GetByID(context.Background(), "RAKHDON23", &sub, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}

	if sub.ExistsInKeystone() {
		return res.WithError(errors.New("entity stored in keystone"))
	}

	return res
}
