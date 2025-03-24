package iid

import (
	"context"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

type Requirement struct {
	entityID keystone.ID
}

func (d *Requirement) Name() string {
	return "Incrementing IDs"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.prepare(actor),
		d.firstID(actor),
		d.secondID(actor),
		d.thirdID(actor),
	}
}

func (d *Requirement) prepare(actor *keystone.Actor) requirements.TestResult {
	usr := &models.User{
		ExternalID: k4id.New().String(),
	}
	mutateErr := actor.Mutate(context.Background(), usr)
	d.entityID = usr.GetKeystoneID()

	return requirements.TestResult{
		Name:  "Prepare ID Entity",
		Error: mutateErr,
	}
}

func (d *Requirement) firstID(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "First ID"}

	id1 := keystone.NewIncrementingID(d.entityID, "trans", "auth")
	idRes, err := id1.Commit(actor)
	if err != nil {
		return res.WithError(err)
	}

	if idRes.IDCount("trans") != 1 {
		return res.WithError(fmt.Errorf("trans count is %d not 1", idRes.IDCount("trans")))
	}
	if idRes.IDCount("auth") != 1 {
		return res.WithError(fmt.Errorf("auth count is %d not 1", idRes.IDCount("auth")))
	}

	return res
}

func (d *Requirement) secondID(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Second ID"}

	id1 := keystone.NewIncrementingID(d.entityID, "trans", "auth")
	idRes, err := id1.Commit(actor)
	if err != nil {
		return res.WithError(err)
	}

	if idRes.IDCount("trans") != 2 {
		return res.WithError(fmt.Errorf("trans count is %d not2", idRes.IDCount("trans")))
	}
	if idRes.IDCount("auth") != 2 {
		return res.WithError(fmt.Errorf("auth count is %d not 2", idRes.IDCount("auth")))
	}

	return res
}

func (d *Requirement) thirdID(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Third ID"}

	id1 := keystone.NewIncrementingID(d.entityID, "trans").WithRead("auth")
	idRes, err := id1.Commit(actor)
	if err != nil {
		return res.WithError(err)
	}

	if idRes.IDCount("trans") != 3 {
		return res.WithError(fmt.Errorf("trans count is %d not 3", idRes.IDCount("trans")))
	}
	if idRes.IDCount("auth") != 2 {
		// Auth ID not increased
		return res.WithError(fmt.Errorf("auth count is %d not 2", idRes.IDCount("auth")))
	}

	return res
}
