package akv

import (
	"context"
	"errors"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct{}

func (d *Requirement) Name() string {
	return "App Key Value"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.put(actor),
		d.get(actor),
	}
}

func (d *Requirement) put(actor *keystone.Actor) requirements.TestResult {

	putResp, putErr := actor.AKVPut(context.Background(), keystone.AKV("val1", 123), keystone.AKV("val2", "abc"))

	if putErr == nil {
		if !putResp.Success {
			putErr = fmt.Errorf("%d - %s", putResp.GetErrorCode(), putResp.ErrorMessage)
		}
	}

	return requirements.TestResult{
		Name:  "Put",
		Error: putErr,
	}
}

func (d *Requirement) get(actor *keystone.Actor) requirements.TestResult {

	resp, getErr := actor.AKVGet(context.Background(), "val1", "val2", "val3")

	if getErr == nil {
		if val, hasVal := resp["val1"]; !hasVal {
			getErr = errors.New("val1 not found")
		} else if val.GetInt() != 123 {
			getErr = errors.New("val1 has wrong value")
		}

		if val, hasVal := resp["val2"]; !hasVal {
			getErr = errors.New("val2 not found")
		} else if val.GetText() != "abc" {
			getErr = errors.New("val2 has wrong value")
		}

		if _, hasVal := resp["val3"]; hasVal {
			getErr = errors.New("val3 should not be found")
		}
	}

	return requirements.TestResult{
		Name:  "Get",
		Error: getErr,
	}
}
