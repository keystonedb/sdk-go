package child_entities

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	subscriptionId     keystone.ID
	firstRenewalId     keystone.ID
	renewalStartTime   time.Time
	renewalCreatedFrom time.Time
	renewalCreatedTo   time.Time
}

func (d *Requirement) Name() string {
	return "Child Entities"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Subscription{})
	conn.RegisterTypes(models.Renewal{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.createSubscription(actor),
		d.createRenewals(actor),
		d.getSummary(actor),
		d.getRenewals(actor),
		d.queryRenewals(actor),
	}
}

func (d *Requirement) createSubscription(actor *keystone.Actor) requirements.TestResult {

	sub := &models.Subscription{
		StartDate: time.Now(),
	}

	createErr := actor.Mutate(context.Background(), sub, keystone.WithMutationComment("Create a subscription"))
	if createErr == nil {
		d.subscriptionId = sub.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create Subscription",
		Error: createErr,
	}
}
func (d *Requirement) createRenewals(actor *keystone.Actor) requirements.TestResult {

	start := time.Now().Truncate(time.Millisecond)
	d.renewalStartTime = start
	d.renewalCreatedFrom = time.Now().Truncate(time.Millisecond)
	for i := 0; i < 30; i++ {
		end := start.AddDate(0, 1, 0)
		renewal := &models.Renewal{
			StartDate:    start,
			EndDate:      end,
			CreationDate: start,
		}
		renewal.SetKeystoneID(d.subscriptionId)
		start = end

		createErr := actor.Mutate(context.Background(), renewal, keystone.WithMutationComment("Create renewal "+strconv.Itoa(i)))
		if createErr != nil {
			return requirements.TestResult{
				Name:  "Create Renewal",
				Error: createErr,
			}
		}
		if i == 0 {
			d.firstRenewalId = renewal.GetKeystoneID()
		}
	}

	d.renewalCreatedTo = time.Now().Add(time.Second)

	return requirements.TestResult{
		Name: "Create Renewals",
	}
}

func (d *Requirement) getSummary(actor *keystone.Actor) requirements.TestResult {

	sub := &models.Subscription{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(sub, d.subscriptionId), sub, keystone.WithDescendantCount(keystone.Type(models.Renewal{})))

	if getErr == nil {
		if sub.NumberOfRenewals != 30 {
			getErr = errors.New("number of renewals mismatch")
		}
	}

	return requirements.TestResult{
		Name:  "Get Summary",
		Error: getErr,
	}
}
func (d *Requirement) getRenewals(actor *keystone.Actor) requirements.TestResult {
	req := requirements.TestResult{Name: "Get Renewals"}

	entities, err := actor.Find(context.Background(), keystone.Type(models.Renewal{}),
		keystone.RetrieveOptions(keystone.WithSummary(), keystone.WithProperties()),
		keystone.ChildOf(d.subscriptionId.String()),
	)
	if err != nil {
		return req.WithError(err)
	}

	if len(entities) != 30 {
		return req.WithError(fmt.Errorf("expected 30 renewals, got %d", len(entities)))
	}

	var renewals []models.Renewal
	if err = keystone.UnmarshalToSlice(&renewals, entities...); err != nil {
		return req.WithError(fmt.Errorf("unmarshal error: %w", err))
	}

	for i, r := range renewals {
		if r.DateCreated().Before(d.renewalCreatedFrom) || r.DateCreated().After(d.renewalCreatedTo) {
			return req.WithError(fmt.Errorf("renewal %d: DateCreated %v not within creation window %v - %v", i, r.DateCreated(), d.renewalCreatedFrom, d.renewalCreatedTo))
		}

		if r.CreationDate.Before(d.renewalCreatedFrom) || r.CreationDate.After(d.renewalCreatedTo) {
			return req.WithError(fmt.Errorf("renewal %d: CreationDate %v not within creation window %v - %v", i, r.CreationDate, d.renewalCreatedFrom, d.renewalCreatedTo))
		}
	}

	return req
}

func (d *Requirement) queryRenewals(actor *keystone.Actor) requirements.TestResult {
	req := requirements.TestResult{Name: "QueryIndex Renewals"}

	entities, err := actor.QueryIndex(context.Background(), keystone.Type(models.Renewal{}),
		[]string{"created"},
		keystone.Limit(30, 0),
		keystone.ChildOf(d.subscriptionId.String()),
	)
	if err != nil {
		return req.WithError(err)
	}

	if len(entities) != 30 {
		return req.WithError(fmt.Errorf("expected 30 renewals, got %d", len(entities)))
	}

	var renewals []models.Renewal
	if err = keystone.UnmarshalToSlice(&renewals, entities...); err != nil {
		return req.WithError(fmt.Errorf("unmarshal error: %w", err))
	}

	for i, r := range renewals {
		if r.DateCreated().Before(d.renewalCreatedFrom) || r.DateCreated().After(d.renewalCreatedTo) {
			return req.WithError(fmt.Errorf("renewal %d: DateCreated %v not within creation window %v - %v", i, r.DateCreated(), d.renewalCreatedFrom, d.renewalCreatedTo))
		}

		if r.CreationDate.Before(d.renewalCreatedFrom) || r.CreationDate.After(d.renewalCreatedTo) {
			return req.WithError(fmt.Errorf("renewal %d: CreationDate %v not within creation window %v - %v", i, r.CreationDate, d.renewalCreatedFrom, d.renewalCreatedTo))
		}
	}

	return req
}
