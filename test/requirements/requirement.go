package requirements

import "github.com/keystonedb/sdk-go/keystone"

type TestResult struct {
	Name  string
	Error error
}

func (t TestResult) WithError(err error) TestResult {
	t.Error = err
	return t
}

func NewResult(name string, err error) TestResult {
	return TestResult{
		Name:  name,
		Error: err,
	}
}

// Reporter is invoked by a Requirement for each TestResult as soon as it is
// produced, so callers can stream output instead of waiting for the full set.
type Reporter func(TestResult)

type Requirement interface {
	Name() string
	Register(conn *keystone.Connection) error
	Verify(actor *keystone.Actor, report Reporter)
}
