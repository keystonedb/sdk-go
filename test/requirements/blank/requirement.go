package blank

import (
	"errors"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct{}

func (d *Requirement) Name() string {
	return "Requirement Title"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.action(actor),
	}
}

func (d *Requirement) action(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Action"}
	return res.WithError(errors.New("not Implemented"))
}
