package status

import (
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct{}

func (d *Requirement) Name() string {
	return "Status Check"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.check(actor),
	}
}

func (d *Requirement) check(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Status Check"}

	status, err := actor.ServerStatus()
	if err != nil {
		return res.WithError(err)
	}

	if !status.GetAuthenticated() {
		return res.WithError(errors.New("expected authenticated status, got unauthenticated"))
	}

	if status.GetAuthenticatedVendor() != actor.VendorID() {
		return res.WithError(errors.New("expected authenticated vendor to match actor's vendor ID"))
	}

	if status.GetAuthenticatedApp() != actor.AppID() {
		return res.WithError(errors.New("expected authenticated app to match actor's app ID"))
	}

	if status.GetUptimeSeconds() < 1 {
		return res.WithError(errors.New("expected uptime to be greater than 0 seconds"))
	}

	if status.GetVersion() == "" {
		return res.WithError(errors.New("expected version to be set, got empty string"))
	}

	return res
}
