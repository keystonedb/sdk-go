package interval

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	createdID keystone.ID
}

func (r *Requirement) Name() string { return "Interval Type CRUD" }

func (r *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.IntervalEntity{})
	return nil
}

func (r *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		r.create(actor),
		r.read(actor),
		r.update(actor),
		r.readUpdated(actor),
	}
}

func (r *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	base := keystone.NewInterval(keystone.IntervalMonth, 3)
	ent := &models.IntervalEntity{
		Name:      "subscription",
		Period:    *base,
		PeriodPtr: base,
	}

	err := actor.Mutate(context.Background(), ent, keystone.WithMutationComment("create interval entity"))
	if err == nil {
		r.createdID = ent.GetKeystoneID()
	}
	return requirements.TestResult{Name: "Create", Error: err}
}

func (r *Requirement) read(actor *keystone.Actor) requirements.TestResult {
	got := &models.IntervalEntity{}
	err := actor.Get(context.Background(), keystone.ByEntityID(got, r.createdID), got, keystone.WithProperties())
	if err == nil {
		if got.Name != "subscription" {
			err = errors.New("name mismatch")
		} else if got.Period.GetType() != keystone.IntervalMonth || got.Period.GetCount() != 3 {
			err = errors.New("period mismatch")
		} else if got.PeriodPtr == nil || got.PeriodPtr.GetType() != keystone.IntervalMonth || got.PeriodPtr.GetCount() != 3 {
			err = errors.New("period ptr mismatch")
		}
	}
	return requirements.TestResult{Name: "Read", Error: err}
}

func (r *Requirement) update(actor *keystone.Actor) requirements.TestResult {
	patch := &models.IntervalEntity{}
	patch.SetKeystoneID(r.createdID)
	patch.Period = *keystone.NewInterval(keystone.IntervalWeek, 2)
	patch.PeriodPtr = keystone.NewInterval(keystone.IntervalWeek, 2)
	err := actor.Mutate(context.Background(), patch, keystone.WithMutationComment("update interval entity"))
	return requirements.TestResult{Name: "Update", Error: err}
}

func (r *Requirement) readUpdated(actor *keystone.Actor) requirements.TestResult {
	got := &models.IntervalEntity{}
	err := actor.Get(context.Background(), keystone.ByEntityID(got, r.createdID), got, keystone.WithProperties())
	if err == nil {
		if got.Period.GetType() != keystone.IntervalWeek || got.Period.GetCount() != 2 {
			err = errors.New("updated period mismatch")
		} else if got.PeriodPtr == nil || got.PeriodPtr.GetType() != keystone.IntervalWeek || got.PeriodPtr.GetCount() != 2 {
			err = errors.New("updated period ptr mismatch")
		}
	}
	return requirements.TestResult{Name: "Read (Updated)", Error: err}
}
