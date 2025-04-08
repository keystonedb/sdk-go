package destroy

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

type Requirement struct {
	entityID keystone.ID
}

func (d *Requirement) Name() string {
	return "Destroy"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.createEntity(actor),
		d.destroyEntity(actor),
	}
}

func (d *Requirement) createEntity(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create Entity"}
	usr := &models.User{
		ExternalID: k4id.New().String(),
	}
	mutateErr := actor.Mutate(context.Background(), usr)
	d.entityID = usr.GetKeystoneID()
	return res.WithError(mutateErr)
}

func (d *Requirement) destroyEntity(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Destroy " + d.entityID.String()}
	destroyed, err := actor.PermanentlyDestroyEntity(context.Background(), keystone.Type(models.User{}), d.entityID, keystone.EidHash(d.entityID), "Destroy user test")
	if err != nil {
		return res.WithError(err)
	}
	if !destroyed {
		return res.WithError(errors.New("entity not destroyed"))
	}
	return res.WithError(err)
}
