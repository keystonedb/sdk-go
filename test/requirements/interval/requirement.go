package interval

import (
	"context"
	"errors"
	"strconv"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

type Requirement struct {
	runID     string
	createdID keystone.ID
}

func (r *Requirement) Name() string { return "Interval Type CRUD + Query" }

func (r *Requirement) Register(conn *keystone.Connection) error {
	r.runID = k4id.New().String()
	conn.RegisterTypes(models.IntervalEntity{})
	return nil
}

func (r *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		r.create(actor),
		r.read(actor),
		r.update(actor),
		r.readUpdated(actor),
		r.createQueryEntities(actor),
		r.queryEqual(actor),
		r.queryNotEqual(actor),
		r.queryGreaterThan(actor),
		r.queryGreaterThanOrEqual(actor),
		r.queryLessThan(actor),
		r.queryLessThanOrEqual(actor),
	}
}

func (r *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	base := keystone.NewInterval(keystone.IntervalMonth, 3)
	ent := &models.IntervalEntity{
		Name:      "subscription-" + r.runID,
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
		if got.Name != "subscription-"+r.runID {
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

// createQueryEntities creates entities with varying intervals for query tests.
// After this, we have:
//   - "day-1-<runID>"   → Day/1
//   - "month-3-<runID>" → Month/3
//   - "month-6-<runID>" → Month/6
//   - "year-1-<runID>"  → Year/1
func (r *Requirement) createQueryEntities(actor *keystone.Actor) requirements.TestResult {
	entities := []struct {
		name     string
		interval *keystone.Interval
	}{
		{"day-1-" + r.runID, keystone.NewInterval(keystone.IntervalDay, 1)},
		{"month-3-" + r.runID, keystone.NewInterval(keystone.IntervalMonth, 3)},
		{"month-6-" + r.runID, keystone.NewInterval(keystone.IntervalMonth, 6)},
		{"year-1-" + r.runID, keystone.NewInterval(keystone.IntervalYear, 1)},
	}

	for _, e := range entities {
		ent := &models.IntervalEntity{
			Name:   e.name,
			Period: *e.interval,
		}
		err := actor.Mutate(context.Background(), ent, keystone.WithMutationComment("create query entity "+e.name))
		if err != nil {
			return requirements.TestResult{Name: "Create Query Entities", Error: err}
		}
	}
	return requirements.TestResult{Name: "Create Query Entities"}
}

// queryEqual: period == Month/3 should return exactly 1 entity (the one created in createQueryEntities)
func (r *Requirement) queryEqual(actor *keystone.Actor) requirements.TestResult {
	target := *keystone.NewInterval(keystone.IntervalMonth, 3)
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.IntervalEntity{}),
		[]string{"name", "period"}, keystone.Limit(10, 0),
		keystone.WhereEquals("period", target),
		keystone.WhereContains("name", r.runID),
	)
	if err == nil {
		if len(entities) != 1 {
			err = errors.New("expected 1 entity with period=Month/3, got " + strconv.Itoa(len(entities)))
		} else {
			got := &models.IntervalEntity{}
			if unErr := keystone.Unmarshal(entities[0], got); unErr != nil {
				err = unErr
			} else if got.Name != "month-3-"+r.runID {
				err = errors.New("expected name month-3-" + r.runID + ", got " + got.Name)
			}
		}
	}
	return requirements.TestResult{Name: "Query Equal", Error: err}
}

// queryNotEqual: period != Month/3 should return 3 entities (day-1, month-6, year-1)
func (r *Requirement) queryNotEqual(actor *keystone.Actor) requirements.TestResult {
	target := *keystone.NewInterval(keystone.IntervalMonth, 3)
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.IntervalEntity{}),
		[]string{"name"}, keystone.Limit(10, 0),
		keystone.WhereNotEquals("period", target),
		keystone.WhereContains("name", r.runID),
	)
	if err == nil {
		if len(entities) < 3 {
			err = errors.New("expected at least 3 entities with period!=Month/3, got " + strconv.Itoa(len(entities)))
		}
	}
	return requirements.TestResult{Name: "Query NotEqual", Error: err}
}

// queryGreaterThan: period > Month/3 should return month-6 and year-1
func (r *Requirement) queryGreaterThan(actor *keystone.Actor) requirements.TestResult {
	target := *keystone.NewInterval(keystone.IntervalMonth, 3)
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.IntervalEntity{}),
		[]string{"name"}, keystone.Limit(10, 0),
		keystone.WhereGreaterThan("period", target),
		keystone.WhereContains("name", r.runID),
	)
	if err == nil {
		if len(entities) < 2 {
			err = errors.New("expected at least 2 entities with period>Month/3, got " + strconv.Itoa(len(entities)))
		}
	}
	return requirements.TestResult{Name: "Query GreaterThan", Error: err}
}

// queryGreaterThanOrEqual: period >= Month/3 should return month-3, month-6, and year-1
func (r *Requirement) queryGreaterThanOrEqual(actor *keystone.Actor) requirements.TestResult {
	target := *keystone.NewInterval(keystone.IntervalMonth, 3)
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.IntervalEntity{}),
		[]string{"name"}, keystone.Limit(10, 0),
		keystone.WhereGreaterThanOrEquals("period", target),
		keystone.WhereContains("name", r.runID),
	)
	if err == nil {
		if len(entities) < 3 {
			err = errors.New("expected at least 3 entities with period>=Month/3, got " + strconv.Itoa(len(entities)))
		}
	}
	return requirements.TestResult{Name: "Query GreaterThanOrEqual", Error: err}
}

// queryLessThan: period < Month/3 should return day-1
func (r *Requirement) queryLessThan(actor *keystone.Actor) requirements.TestResult {
	target := *keystone.NewInterval(keystone.IntervalMonth, 3)
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.IntervalEntity{}),
		[]string{"name"}, keystone.Limit(10, 0),
		keystone.WhereLessThan("period", target),
		keystone.WhereContains("name", r.runID),
	)
	if err == nil {
		if len(entities) < 1 {
			err = errors.New("expected at least 1 entity with period<Month/3, got " + strconv.Itoa(len(entities)))
		}
	}
	return requirements.TestResult{Name: "Query LessThan", Error: err}
}

// queryLessThanOrEqual: period <= Month/3 should return day-1 and month-3
func (r *Requirement) queryLessThanOrEqual(actor *keystone.Actor) requirements.TestResult {
	target := *keystone.NewInterval(keystone.IntervalMonth, 3)
	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.IntervalEntity{}),
		[]string{"name"}, keystone.Limit(10, 0),
		keystone.WhereLessThanOrEquals("period", target),
		keystone.WhereContains("name", r.runID),
	)
	if err == nil {
		if len(entities) < 2 {
			err = errors.New("expected at least 2 entities with period<=Month/3, got " + strconv.Itoa(len(entities)))
		}
	}
	return requirements.TestResult{Name: "Query LessThanOrEqual", Error: err}
}
