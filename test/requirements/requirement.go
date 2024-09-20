package requirements

import "github.com/keystonedb/sdk-go/keystone"

type TestResult struct {
	Name  string
	Error error
}

func NewResult(name string, err error) TestResult {
	return TestResult{
		Name:  name,
		Error: err,
	}
}

type Requirement interface {
	Name() string
	Register(conn *keystone.Connection) error
	Verify(actor *keystone.Actor) []TestResult
}
