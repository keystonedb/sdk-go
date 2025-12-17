package squid

import (
	"fmt"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

type Requirement struct {
	sqkey string
	squat string
	squid uint32
}

func (d *Requirement) Name() string {
	return "SeQUence IDs"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.recover(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "SQUID Creation"}

	d.sqkey = k4id.New().String() + "-test"
	for i := 1; i <= 200; i++ {
		squid, err := actor.Squid(d.sqkey)
		if err != nil {
			return res.WithError(err)
		}
		d.squat = squid.GetSquat()
		d.squid = squid.GetSquid()
		if d.squid != uint32(i) {
			return res.WithError(fmt.Errorf("squid does not match expected value %d != %d", d.squid, i))
		}
	}

	return res
}

func (d *Requirement) recover(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "SQUID Recovery"}

	recovered, err := actor.SquidRetrieve(d.sqkey, d.squat)
	if err != nil {
		return res.WithError(err)
	}
	if recovered.GetSquid() != d.squid {
		return res.WithError(fmt.Errorf("squid does not match expected value %d != %d", recovered.GetSquid(), d.squid))
	}

	return res
}
