package grandchildren

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
	createdMap         map[keystone.ID]timeRange
	note1ID            string
	note2ID            string
	note3ID            string
}

type timeRange struct {
	start time.Time
	end   time.Time
}

func (d *Requirement) Name() string {
	return "Grandchild Entities"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Subscription{})
	conn.RegisterTypes(models.Renewal{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	d.createdMap = make(map[keystone.ID]timeRange)
	report(d.createSubscription(actor))
	report(d.createRenewals(actor))
	report(d.summaryChildren(actor))
	report(d.removeChild(actor))
	report(d.loadChildren(actor))
	report(d.updateChildren(actor))
	report(d.verifyChildren(actor))
	report(d.truncateChildren(actor))
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
	for i := 0; i < 3; i++ {
		end := start.AddDate(0, 1, 0)
		renewal := &models.Renewal{
			StartDate: start,
			EndDate:   end,
		}
		renewal.SetKeystoneID(d.subscriptionId)
		start = end

		f1 := &models.RenewalNote{
			Date: time.Now(),
			Note: "Note One",
		}
		f2 := &models.RenewalNote{
			Date: time.Now().Add(time.Second * 10),
			Note: "Note Two",
		}
		f3 := &models.RenewalNote{
			Date: time.Now().Add(time.Second * 30),
			Note: "Note Three",
		}

		renewal.Notes = append(renewal.Notes, f1, f2, f3)
		renewal.AddChildren(renewal.Notes)

		r := timeRange{start: time.Now().Truncate(time.Millisecond).Add(-time.Second)}
		createErr := actor.Mutate(context.Background(), renewal, keystone.WithMutationComment("Create renewal "+strconv.Itoa(i)))
		r.end = time.Now().Truncate(time.Millisecond).Add(time.Second)
		d.createdMap[renewal.GetKeystoneID()] = r
		time.Sleep(time.Millisecond)
		if createErr != nil {
			return requirements.TestResult{
				Name:  "Create Renewal",
				Error: createErr,
			}
		}
		if i == 0 {
			d.firstRenewalId = renewal.GetKeystoneID()
			d.note1ID = f1.ChildID()
			d.note2ID = f2.ChildID()
			d.note3ID = f3.ChildID()

			if d.note1ID == "" || d.note2ID == "" || d.note3ID == "" {
				return requirements.TestResult{
					Name:  "Create Renewal",
					Error: fmt.Errorf("failed to create grandchildren, or children did not return a child ID, %s : %s : %s", d.note1ID, d.note2ID, d.note3ID),
				}
			}
		}
	}

	d.renewalCreatedTo = time.Now().Add(time.Second)

	return requirements.TestResult{
		Name: "Create Renewals",
	}
}

func (d *Requirement) summaryChildren(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Load Renewal with Grandchild Summaries",
	}

	renewal := &models.Renewal{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(renewal, d.firstRenewalId), renewal, keystone.WithChildSummary())
	if getErr != nil {
		return res.WithError(getErr)
	}

	if renewal.NumNotes != 3 {
		return res.WithError(fmt.Errorf("expected 3 notes, got %d", renewal.NumNotes))
	}

	return res
}

func (d *Requirement) removeChild(actor *keystone.Actor) requirements.TestResult {
	renewal := &models.Renewal{}
	renewal.SetKeystoneID(d.firstRenewalId)
	renewal.RemoveChild(models.RenewalNote{}, d.note2ID)

	removeErr := actor.Mutate(context.Background(), renewal, keystone.WithMutationComment("Remove a note"))
	return requirements.TestResult{
		Name:  "Remove a Note from the Renewal",
		Error: removeErr,
	}
}

func (d *Requirement) loadChildren(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Load Renewal with Grandchild Notes",
	}

	renewal := &models.Renewal{}
	notes := keystone.WithChildren(keystone.Type(models.RenewalNote{}))
	getErr := actor.GetByID(context.Background(), d.firstRenewalId, renewal, keystone.WithProperties(), keystone.WithSummary(), &notes)
	renewal.Notes = keystone.ChildrenFromLoader[models.RenewalNote](notes)

	if getErr != nil {
		return res.WithError(getErr)
	}
	if len(renewal.Notes) != 2 {
		return res.WithError(fmt.Errorf("expected 2 notes, got %d", len(renewal.Notes)))
	}

	hasNote1 := false
	hasNote3 := false
	for _, n := range renewal.Notes {
		if n.ChildID() == d.note1ID {
			hasNote1 = true
			if n.Note != "Note One" {
				return res.WithError(fmt.Errorf("expected note 1 to be 'Note One', got '%s'", n.Note))
			}
		}
		if n.ChildID() == d.note3ID {
			hasNote3 = true
			if n.Note != "Note Three" {
				return res.WithError(fmt.Errorf("expected note 3 to be 'Note Three', got '%s'", n.Note))
			}
		}
		if n.ChildID() == d.note2ID {
			return res.WithError(errors.New("note 2 was not removed"))
		}
	}
	if !hasNote1 {
		return res.WithError(errors.New("note 1 was not found"))
	}
	if !hasNote3 {
		return res.WithError(errors.New("note 3 was not found"))
	}

	return res
}

func (d *Requirement) updateChildren(actor *keystone.Actor) requirements.TestResult {
	renewal := &models.Renewal{}
	renewal.SetKeystoneID(d.firstRenewalId)
	chd := keystone.NewDynamicChild(models.RenewalNote{})
	chd.SetChildID(d.note1ID)
	chd.Append("Note", []byte(`"Updated Note"`))
	renewal.AddChild(chd)
	writeErr := actor.Mutate(context.Background(), renewal)

	return requirements.TestResult{
		Name:  "Update a Grandchild Note",
		Error: writeErr,
	}
}

func (d *Requirement) verifyChildren(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Verify Updated Grandchild Note",
	}

	renewal := &models.Renewal{}
	notes := keystone.WithChildren(keystone.Type(models.RenewalNote{}), d.note1ID)
	getErr := actor.Get(context.Background(), keystone.ByEntityID(renewal, d.firstRenewalId), renewal, &notes)
	renewal.Notes = keystone.ChildrenFromLoader[models.RenewalNote](notes)

	if getErr != nil {
		return res.WithError(getErr)
	}

	if len(renewal.Notes) != 1 {
		return res.WithError(fmt.Errorf("expected 1 note, got %d", len(renewal.Notes)))
	}

	if renewal.Notes[0].ChildID() != d.note1ID {
		return res.WithError(errors.New("note 1 ID was not set"))
	}

	if renewal.Notes[0].Note != "Updated Note" {
		return res.WithError(fmt.Errorf("expected 'Updated Note', got '%s'", renewal.Notes[0].Note))
	}

	return res
}

func (d *Requirement) truncateChildren(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{
		Name: "Truncate Grandchildren",
	}

	renewal := &models.Renewal{}
	renewal.SetKeystoneID(d.firstRenewalId)

	// Check children exist
	notes := keystone.WithChildren(keystone.Type(models.RenewalNote{}))
	getErr := actor.Get(context.Background(), keystone.ByEntityID(renewal, d.firstRenewalId), renewal, &notes)
	if getErr != nil {
		return res.WithError(getErr)
	}

	renewal.Notes = keystone.ChildrenFromLoader[models.RenewalNote](notes)
	if len(renewal.Notes) == 0 {
		return res.WithError(fmt.Errorf("expected notes, got %d", len(renewal.Notes)))
	}

	renewal.TruncateByType(models.RenewalNote{})

	removeErr := actor.Mutate(context.Background(), renewal, keystone.WithMutationComment("Remove all notes"))
	if removeErr != nil {
		return res.WithError(removeErr)
	}

	renewal = &models.Renewal{}
	loadNotes := keystone.WithChildren(keystone.Type(models.RenewalNote{}))
	getErr = actor.Get(context.Background(), keystone.ByEntityID(renewal, d.firstRenewalId), renewal, &loadNotes)
	if getErr != nil {
		return res.WithError(getErr)
	}

	renewal.Notes = keystone.ChildrenFromLoader[models.RenewalNote](loadNotes)
	if len(renewal.Notes) != 0 {
		return res.WithError(fmt.Errorf("expected 0 notes, got %d", len(renewal.Notes)))
	}

	return res
}
