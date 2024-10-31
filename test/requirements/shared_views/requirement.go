package shared_views

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

const vendor2ID = "ven2"
const app2ID = "ap2"

type Requirement struct {
	entityID         string
	sharedViewToken  string
	psn              *models.Person
	initialView      *proto.SharedViewResponse
	secondConnection *keystone.Connection
	secondActor      *keystone.Actor
}

func (d *Requirement) Name() string {
	return "Shared Views"
}

func (d *Requirement) Register(conn *keystone.Connection) error { return nil }

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	d.secondConnection = keystone.NewConnection(actor.Connection().DirectClient(), vendor2ID, app2ID, "test-access-token")
	act2 := d.secondConnection.Actor("tt", "127.0.0.2", "random-userid", "UserAgent")
	d.secondActor = &act2
	return []requirements.TestResult{
		d.share(actor),
		d.read(actor),
		d.verify(actor),
	}
}

func (d *Requirement) share(actor *keystone.Actor) requirements.TestResult {
	d.psn = &models.Person{
		BaseEntity:   keystone.BaseEntity{},
		Name:         "John",
		HeightInCm:   123,
		DOB:          time.Now(),
		BankBalance:  keystone.NewAmount("USD", 345),
		FullName:     keystone.NewSecretString("John Doe", "Jo*** D***"),
		AccountPin:   "1234",
		SecretAnswer: keystone.NewSecretString("Pet Name", "Pe*******"),
	}
	actor.Mutate(context.Background(), d.psn)
	d.entityID = d.psn.GetKeystoneID()

	resp, err := actor.ShareView(context.Background(), proto.NewVendorApp(vendor2ID, app2ID), keystone.NewSharedView("height_in_cm", "name", "full_name").ForEntity(d.entityID))

	if resp != nil {
		d.initialView = resp
	} else if err == nil {
		err = errors.New("no response available")
	}

	return requirements.TestResult{
		Name:  "Share View",
		Error: err,
	}
}

func (d *Requirement) read(actor *keystone.Actor) requirements.TestResult {

	allViews, err := actor.SharedViews(context.Background(), nil, d.entityID, "", false)
	if err == nil {
		if allViews == nil {
			err = errors.New("no shared views response given")
		} else if len(allViews.GetViews()) == 0 {
			err = errors.New("no shared views available")
		} else {
			viewCount := len(allViews.GetViews())
			switch viewCount {
			case 1:
				view := allViews.GetViews()[0]
				if view.GetToken() != d.initialView.GetToken() {
					err = errors.New("token mismatch")
				}
				break
			default:
				err = errors.New("unexpected 1 view, got " + strconv.Itoa(viewCount))
			}
		}
	}

	return requirements.TestResult{
		Name:  "Read Shared Views",
		Error: err,
	}
}

func (d *Requirement) verify(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{
		Name: "Verify Shared View Access",
	}

	loadProps := keystone.WithDecryptedProperties("height_in_cm", "name", "bank_balance", "full_name")

	psn1 := &models.Person{}
	psn2 := &models.Person{}
	a1Err := actor.GetByID(context.Background(), d.entityID, psn1, loadProps)
	if a1Err != nil {
		result.Error = a1Err
		return result
	}

	if psn1.HeightInCm != d.psn.HeightInCm {
		result.Error = errors.New("owner: height in cm mismatch")
		return result
	}

	a2Err := d.secondActor.GetSharedByID(context.Background(), actor.VendorApp(), d.entityID, psn2, loadProps)
	if a2Err != nil {
		result.Error = a2Err
		return result
	}

	if psn2.HeightInCm != d.psn.HeightInCm {
		result.Error = errors.New("viewer: height in cm mismatch")
		return result
	}

	if psn1.FullName.String() != d.psn.FullName.Original {
		result.Error = errors.New("owner: full name mismatch")
		return result
	}

	if psn2.FullName.String() != d.psn.FullName.Masked {
		result.Error = errors.New("viewer: full name not masked")
		return result
	}

	if psn1.BankBalance.Currency != d.psn.BankBalance.Currency {
		result.Error = errors.New("owner: bank balance should be loaded")
		return result
	}

	if psn2.BankBalance.Currency != "" {
		result.Error = errors.New("viewer: bank balance should not be loaded")
		return result
	}

	return result
}
