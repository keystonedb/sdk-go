package requirements

import (
	"errors"

	"github.com/keystonedb/sdk-go/keystone"
)

type DummyRequirement struct {
}

func (d DummyRequirement) Register(conn *keystone.Connection) error {
	return nil
}
func (d DummyRequirement) Name() string { return "Dummy" }

func (d DummyRequirement) Verify(actor *keystone.Actor) []TestResult {
	return []TestResult{
		NewResult("Failure", errors.New("fail")),
		NewResult("Success", nil),
	}
}
