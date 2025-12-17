package labels

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
	createdID keystone.ID
	labelKey  string
}

func (d *Requirement) Name() string {
	return "Labels"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.User{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.store(actor),
		d.retrieve(actor),
	}
}

func (d *Requirement) store(actor *keystone.Actor) requirements.TestResult {

	d.labelKey = "lbl-" + strconv.Itoa(int(time.Now().Unix()))
	usr := &models.User{
		Validate: "label-write",
	}
	usr.AddLabel(d.labelKey, "label-value")

	createErr := actor.Mutate(context.Background(), usr, keystone.WithMutationComment("Create a user with a label"))
	if createErr == nil {
		d.createdID = usr.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: createErr,
	}
}

func (d *Requirement) retrieve(actor *keystone.Actor) requirements.TestResult {
	usr := &models.User{}
	located, getErr := actor.Find(context.Background(), keystone.Type(usr), keystone.RetrieveOptions(
		keystone.WithProperties("validate"),
		keystone.WithLabels(),
	), keystone.WithLabel(d.labelKey, "label-value"))

	if getErr == nil {
		if len(located) < 0 {
			getErr = errors.New("no user found")
		} else {
			getErr = keystone.Unmarshal(located[0], usr)
			if getErr == nil {
				if usr.Validate != "label-write" {
					getErr = errors.New("unexpected value for 'validate'")
				}
			}
		}
	}

	return requirements.TestResult{
		Name:  "Read",
		Error: getErr,
	}
}
