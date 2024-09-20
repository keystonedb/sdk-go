package stats

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"strconv"
	"time"
)

type Requirement struct {
	conn      *keystone.Connection
	createdID string
}

func (d *Requirement) Name() string {
	return "schema Statistics"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	d.conn = conn
	conn.RegisterTypes(models.Person{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.allTimeCount(actor),
		d.todayCount(actor),
		d.lastWeekBreakdown(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	psn := &models.Person{
		BaseEntity: keystone.BaseEntity{},
		Name:       "Sally Stat",
	}

	createErr := actor.Mutate(context.Background(), psn, "Known entity")
	if createErr == nil {
		d.createdID = psn.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) allTimeCount(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "All Time Count",
	}

	if d.conn == nil {
		res.Error = errors.New("no connection available")
		return res
	}

	schema := &proto.Key{Key: keystone.Type(models.Person{}), Source: actor.VendorApp()}

	ks := d.conn.DirectClient()
	stats, err := ks.SchemaStatistics(context.Background(), &proto.SchemaStatisticsRequest{
		Authorization: actor.Authorization(),
		Schema:        schema,
	})

	if err != nil {
		res.Error = err
		return res
	}

	res.Name += " (" + strconv.Itoa(int(stats.GetInRangeCount())) + " entries)"

	if stats.GetInRangeCount() < 1 {
		res.Error = errors.New("at least one entry should have been created")
	}

	return res
}

func (d *Requirement) todayCount(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Today Count",
	}

	if d.conn == nil {
		res.Error = errors.New("no connection available")
		return res
	}

	schema := &proto.Key{Key: keystone.Type(models.Person{}), Source: actor.VendorApp()}

	ks := d.conn.DirectClient()
	stats, err := ks.SchemaStatistics(context.Background(), &proto.SchemaStatisticsRequest{
		Authorization: actor.Authorization(),
		Schema:        schema,
		CreatedFrom:   proto.CreateDate(time.Now()),
	})

	if err != nil {
		res.Error = err
		return res
	}

	res.Name += " (" + strconv.Itoa(int(stats.GetInRangeCount())) + " entries)"

	if stats.GetInRangeCount() < 1 {
		res.Error = errors.New("at least one entry should have been created")
	}

	return res
}

func (d *Requirement) lastWeekBreakdown(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Weekly Breakdown",
	}

	if d.conn == nil {
		res.Error = errors.New("no connection available")
		return res
	}

	schema := &proto.Key{Key: keystone.Type(models.Person{}), Source: actor.VendorApp()}

	ks := d.conn.DirectClient()
	stats, err := ks.SchemaStatistics(context.Background(), &proto.SchemaStatisticsRequest{
		Authorization:    actor.Authorization(),
		Schema:           schema,
		CreatedFrom:      proto.CreateDate(time.Now().Add(-7 * 24 * time.Hour)),
		CreatedUntil:     proto.CreateDate(time.Now()),
		IncludeBreakdown: true,
	})

	if err != nil {
		res.Error = err
		return res
	}

	res.Name += " (" + strconv.Itoa(int(stats.GetInRangeCount())) + " entries / " + strconv.Itoa(len(stats.GetDailyCount())) + " days)"

	if len(stats.GetDailyCount()) < 1 {
		res.Error = errors.New("at least one day should have been returned")
		return res
	}

	if stats.GetInRangeCount() < 1 {
		res.Error = errors.New("at least one entry should have been created")
	}

	return res
}
