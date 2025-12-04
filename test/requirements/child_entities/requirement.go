package child_entities

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	subscriptionId keystone.ID
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

	start := time.Now()
	for i := 0; i < 30; i++ {
		end := start.AddDate(0, 1, 0)
		renewal := &models.Renewal{
			StartDate:    start,
			EndDate:      end,
			CreationDate: time.Now(),
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
	}

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

	entities, err := actor.Find(context.Background(), keystone.Type(models.Renewal{}), keystone.WithProperties(), keystone.ChildOf(d.subscriptionId.String()))
	if err != nil {
		return req.WithError(err)
	}

	if len(entities) != 30 {
		return req.WithError(errors.New("number of renewals mismatch"))
	}

	return req
}
