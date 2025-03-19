package prewrite

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"net/http"
	"strings"
	"time"
)

var (
	Name            = "A"
	Name2           = "B"
	Name3           = "C"
	HeightInCm      = int64(190)
	DOB             = time.Date(1985, time.June, 24, 0, 0, 0, 0, time.UTC)
	BalanceCurrency = "GBP"
	BalanceAmount   = int64(1000)
)

type Requirement struct {
	createdID keystone.ID
}

func (d *Requirement) Name() string {
	return "Pre-Write Value Validation"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Person{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.read(actor),
		d.updateFail(actor),
		d.update(actor),
		d.updateAgain(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {

	psn := &models.Person{
		BaseEntity:   keystone.BaseEntity{},
		Name:         Name,
		HeightInCm:   HeightInCm,
		DOB:          DOB,
		BankBalance:  *keystone.NewAmount(BalanceCurrency, BalanceAmount),
		FullName:     keystone.NewSecureString("John Doe", "Jo*** D***"),
		AccountPin:   keystone.NewVerifyString("4321"),
		SecretAnswer: keystone.NewSecureString("Pet Name", "Pe*******"),
	}

	createErr := actor.Mutate(context.Background(), psn, keystone.WithMutationComment("Create a person"))
	if createErr == nil {
		d.createdID = psn.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) read(actor *keystone.Actor) requirements.TestResult {

	psn := &models.Person{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(psn, d.createdID), psn, keystone.WithDecryptedProperties())

	if getErr == nil {
		if psn.Name != Name {
			getErr = errors.New("name mismatch")
		} else if psn.HeightInCm != HeightInCm {
			getErr = errors.New("height mismatch")
		} else if psn.DOB.UTC() != DOB.UTC() {
			getErr = errors.New("dob mismatch")
		} else if psn.BankBalance.Currency != BalanceCurrency {
			getErr = errors.New("balance currency mismatch, got " + psn.BankBalance.Currency)
		} else if psn.BankBalance.Units != BalanceAmount {
			getErr = errors.New("balance amount mismatch")
		} else if psn.FullName.Original != "John Doe" {
			getErr = errors.New("full name mismatch, got " + psn.FullName.Original)
		} else if psn.SecretAnswer.String() != "Pet Name" {
			getErr = errors.New("secret answer mismatch")
		}
	}

	return requirements.TestResult{
		Name:  "Read",
		Error: getErr,
	}
}

func (d *Requirement) updateFail(actor *keystone.Actor) requirements.TestResult {
	psn := &models.Person{}
	psn.SetKeystoneID(d.createdID)
	psn.Name = Name2
	updateErr := actor.Mutate(context.Background(), psn, keystone.WithMutationComment("Update a person"),
		keystone.MutateProperties("name"), keystone.MatchExisting(keystone.WhereEquals("name", Name+"Mismatch")))

	if updateErr == nil {
		updateErr = errors.New("expected name mismatch error")
	} else {
		var ksErr *keystone.Error
		if ok := errors.As(updateErr, &ksErr); ok {
			if ksErr.Extended[0] != "name : value mismatch" {
				updateErr = errors.New("missing name mismatch error")
			} else if ksErr.ErrorCode != http.StatusConflict {
				updateErr = errors.New("expected status conflict error")
			} else if ksErr.ErrorMessage != "Validation conditions not met" {
				updateErr = errors.New("expected validation conditions not met error")
			} else {
				updateErr = nil
			}
		} else if strings.Contains(updateErr.Error(), "Validation conditions not met") {
			updateErr = nil
		}
	}

	return requirements.TestResult{
		Name:  "Mismatch PreFetch",
		Error: updateErr,
	}
}

func (d *Requirement) update(actor *keystone.Actor) requirements.TestResult {
	psn := &models.Person{}
	psn.SetKeystoneID(d.createdID)
	psn.Name = Name2
	updateErr := actor.Mutate(context.Background(), psn, keystone.WithMutationComment("Update a person"),
		keystone.MutateProperties("name"), keystone.MatchExisting(keystone.WhereEquals("name", Name), keystone.WhereEquals("height_in_cm", HeightInCm)))

	if updateErr == nil {
		updateErr = actor.Get(context.Background(), keystone.ByEntityID(psn, d.createdID), psn, keystone.WithProperties())
		if updateErr != nil {
			// Return this
		} else if psn.Name != Name2 {
			updateErr = errors.New("name mismatch")
		}
	}

	return requirements.TestResult{
		Name:  "Update Conditional",
		Error: updateErr,
	}
}

func (d *Requirement) updateAgain(actor *keystone.Actor) requirements.TestResult {
	psn := &models.Person{}
	psn.SetKeystoneID(d.createdID)
	psn.Name = Name3
	updateErr := actor.Mutate(context.Background(), psn,
		keystone.WithMutationComment("Update a person"),
		keystone.MutateProperties("name"),
		keystone.MatchExisting(keystone.WhereEquals("name", Name2), keystone.WhereGreaterThan("height_in_cm", 189)))

	if updateErr == nil {
		updateErr = actor.Get(context.Background(), keystone.ByEntityID(psn, d.createdID), psn, keystone.WithProperties())
		if updateErr != nil {
			// Return this
		} else if psn.Name != Name3 {
			updateErr = errors.New("name mismatch")
		}
	}

	return requirements.TestResult{
		Name:  "Update Conditional",
		Error: updateErr,
	}
}
