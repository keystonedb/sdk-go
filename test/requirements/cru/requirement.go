package cru

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"time"
)

var (
	Name            = "A"
	Name2           = "B"
	HeightInCm      = int64(190)
	DOB             = time.Date(1985, time.June, 24, 0, 0, 0, 0, time.UTC)
	BalanceCurrency = "GBP"
	BalanceAmount   = int64(1000)
)

type Requirement struct {
	createdID keystone.ID
}

func (d *Requirement) Name() string {
	return "Create Read Update"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Person{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.read(actor),
		d.readVerifyFail(actor),
		d.update(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {

	psn := &models.Person{
		BaseEntity:   keystone.BaseEntity{},
		Name:         Name,
		HeightInCm:   HeightInCm,
		DOB:          DOB,
		BankBalance:  keystone.NewAmount(BalanceCurrency, BalanceAmount),
		FullName:     keystone.NewSecretString("John Doe", "Jo*** D***"),
		AccountPin:   keystone.NewVerifyString("1234"),
		SecretAnswer: keystone.NewSecretString("Pet Name", "Pe*******"),
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
	getErr := actor.Get(context.Background(), keystone.ByEntityID(psn, d.createdID), psn, keystone.WithDecryptedProperties(), keystone.WithVerifiedProperty("account_pin", "1234"))

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
		} else if !psn.AccountPin.WasChecked() {
			getErr = errors.New("account pin was not checked")
		} else if !psn.AccountPin.Verified() {
			getErr = errors.New("account pin was not verified")
		} else if psn.SecretAnswer.String() != "Pet Name" {
			getErr = errors.New("secret answer mismatch")
		}
	}

	return requirements.TestResult{
		Name:  "Read",
		Error: getErr,
	}
}
func (d *Requirement) readVerifyFail(actor *keystone.Actor) requirements.TestResult {

	psn := &models.Person{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(psn, d.createdID), psn, keystone.WithVerifiedProperty("account_pin", "4312"))

	if getErr == nil {
		if !psn.AccountPin.WasChecked() {
			getErr = errors.New("account pin was not checked")
		} else if psn.AccountPin.Verified() {
			getErr = errors.New("account pin was incorrectly verified")
		}
	}

	return requirements.TestResult{
		Name:  "Read Verify Failure",
		Error: getErr,
	}
}

func (d *Requirement) update(actor *keystone.Actor) requirements.TestResult {
	psn := &models.Person{}
	psn.SetKeystoneID(d.createdID)
	psn.Name = Name2
	updateErr := actor.Mutate(context.Background(), psn, keystone.WithMutationComment("Update a person"), keystone.MutateProperties("name"))

	if updateErr == nil {
		updateErr = actor.Get(context.Background(), keystone.ByEntityID(psn, d.createdID), psn, keystone.WithProperties())
		if updateErr != nil {
			// Return this
		} else if psn.Name != Name2 {
			updateErr = errors.New("name mismatch")
		}
	}

	return requirements.TestResult{
		Name:  "Update",
		Error: updateErr,
	}
}
