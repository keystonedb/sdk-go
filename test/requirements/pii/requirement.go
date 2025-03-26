package pii

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
	"time"
)

var (
	PersonName = "John Smith"
	Email      = "john.smith@example.com"
	Phone      = "1234567890"
	Phone2     = "87654567867"
)

type Requirement struct {
	createdID    keystone.ID
	createdRefID keystone.ID
	piiToken     string
	referenceID  string
}

func (d *Requirement) Name() string {
	return "PII Data"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.PiiPerson{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	gen := k4id.DefaultGenerator()
	d.referenceID = gen.New().UUID()
	return []requirements.TestResult{
		d.updateWithoutPii(actor),
		d.createToken(actor),
		d.createReuseToken(actor),
		d.create(actor),
		d.createReference(actor),
		d.read(actor, true, " After Create"),
		d.read(actor, false, " After Create - Ref"),
		d.updateWithoutPiiWrite(actor),
		d.update(actor),
		d.read(actor, true, " After Update"),
		d.read(actor, false, " After Update - Ref"),
		d.createWithoutToken(actor),
		d.anonymize(actor),
		d.readAnonymized(actor),
		d.restore(actor),
		d.read(actor, true, " After Restore"),
		d.read(actor, false, " After Restore - Ref"),
	}
}

func (d *Requirement) createWithoutToken(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Create Without Token"}
	psn := &models.PiiPerson{
		Name:   keystone.NewPersonName(PersonName),
		Email:  keystone.NewEmail(Email),
		Phone:  keystone.NewPhone(Phone),
		NonPii: "Random Value",
	}

	createErr := actor.Mutate(context.Background(), psn)
	if createErr == nil {
		return result.WithError(errors.New("expect failure writing pii with no token"))
	}

	return result
}

func (d *Requirement) createToken(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Create Token"}
	token, err := actor.NewGDPRToken(d.referenceID, "GB")
	if err != nil {
		return result.WithError(err)
	}
	d.piiToken = token
	return result
}

func (d *Requirement) createReuseToken(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "ReUse PII Token"}
	token, err := actor.NewGDPRToken(d.referenceID, "GB")
	if err != nil {
		return result.WithError(err)
	}
	if d.piiToken != token {
		return result.WithError(errors.New("pii token not reused"))
	}
	return result
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Create"}
	psn := &models.PiiPerson{
		Name:   keystone.NewPersonName(PersonName),
		Email:  keystone.NewEmail(Email),
		Phone:  keystone.NewPhone(Phone),
		NonPii: "Random Value",
	}

	createErr := actor.Mutate(context.Background(), psn, keystone.WithPiiToken(d.piiToken))
	if createErr != nil {
		return result.WithError(createErr)
	}

	d.createdID = psn.GetKeystoneID()
	return result
}
func (d *Requirement) createReference(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Create With Reference"}
	psn := &models.PiiPerson{
		Name:   keystone.NewPersonName(PersonName),
		Email:  keystone.NewEmail(Email),
		Phone:  keystone.NewPhone(Phone),
		NonPii: "Random Value",
	}

	createErr := actor.Mutate(context.Background(), psn, keystone.WithPiiReference(actor.VendorID(), actor.AppID(), d.referenceID))
	if createErr != nil {
		return result.WithError(createErr)
	}

	d.createdRefID = psn.GetKeystoneID()
	return result
}

func (d *Requirement) read(actor *keystone.Actor, primaryID bool, reason string) requirements.TestResult {
	result := requirements.TestResult{Name: "Read " + reason}

	readID := d.createdID
	if !primaryID {
		readID = d.createdRefID
	}

	psn := &models.PiiPerson{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(psn, readID), psn, keystone.WithDecryptedProperties())

	if getErr != nil {
		return result.WithError(getErr)
	}

	if psn.Name.String() != PersonName {
		getErr = errors.New("name mismatch")
	} else if psn.Email.String() != Email {
		getErr = errors.New("email mismatch")
	} else if psn.Phone.String() != Phone && psn.Phone.String() != Phone2 {
		getErr = errors.New("phone mismatch")
	}

	return result
}

func (d *Requirement) update(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Pii Info Update"}

	psn := &models.PiiPerson{}
	psn.Phone = keystone.NewPhone(Phone2)
	psn.NonPii = "Pii Updated Alongside"
	psn.SetKeystoneID(d.createdID)

	createErr := actor.Mutate(context.Background(), psn)
	if createErr != nil {
		return result.WithError(createErr)
	}

	return result
}

func (d *Requirement) updateWithoutPiiWrite(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Pii Info Update - No historical pii"}

	psn := &models.PiiPerson{}
	psn.NonPii = "Pii Updated Alongside"
	createErr := actor.Mutate(context.Background(), psn)
	if createErr != nil {
		return result.WithError(createErr)
	}

	psn.Phone = keystone.NewPhone(Phone2)

	updateErr := actor.Mutate(context.Background(), psn, keystone.WithPiiToken(d.piiToken))
	if updateErr != nil {
		return result.WithError(createErr)
	}

	return result
}

func (d *Requirement) updateWithoutPii(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Pii Info Update"}

	psn := &models.PiiPerson{
		NonPii: "Random Value Without Pii",
	}
	psn.SetKeystoneID(d.createdID)

	createErr := actor.Mutate(context.Background(), psn)
	if createErr != nil {
		return result.WithError(createErr)
	}

	return result
}
func (d *Requirement) anonymize(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Anonymize"}
	resp, err := actor.Anonymize(d.piiToken)
	if err != nil {
		return result.WithError(err)
	}
	if resp.GetRecoveryUntil() == nil {
		return result.WithError(errors.New("recovery until is nil"))
	}

	if resp.GetSuccess() != true {
		return result.WithError(errors.New("anonymize failed"))
	}

	recoTime := resp.GetRecoveryUntil().AsTime()
	if recoTime.Before(time.Now()) {
		return result.WithError(errors.New("recovery until is before now"))
	}
	if recoTime.After(time.Now().Add(time.Hour * 24 * 7)) {
		return result.WithError(errors.New("recovery until is after 7 days"))
	}

	return result
}
func (d *Requirement) readAnonymized(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Read Anonymized"}

	psn := &models.PiiPerson{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(psn, d.createdID), psn, keystone.WithDecryptedProperties())

	if getErr != nil {
		return result.WithError(getErr)
	}

	if psn.Name.String() != PersonName {
		getErr = errors.New("name mismatch")
	} else if psn.Email.String() != Email {
		getErr = errors.New("email mismatch")
	} else if psn.Phone.String() != Phone {
		getErr = errors.New("phone mismatch")
	}

	return result
}

func (d *Requirement) restore(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Anonymize Rollback"}
	resp, err := actor.AnonymizeRollback(d.piiToken)
	if err != nil {
		return result.WithError(err)
	}
	if resp.GetRecoveryUntil() != nil {
		return result.WithError(errors.New("recovery until is not nil"))
	}
	if resp.GetSuccess() != true {
		return result.WithError(errors.New("anonymize rollback failed"))
	}
	return result
}
