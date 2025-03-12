package stringset

import (
	"context"
	"errors"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"strings"
)

type Requirement struct {
	createdID keystone.ID
}

type Settable struct {
	keystone.BaseEntity
	TheList keystone.StringSet
}

func (d *Requirement) Name() string {
	return "String Sets"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Person{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	resp := requirements.TestResult{
		Name: "Full Cycle",
	}

	item := &Settable{
		TheList: keystone.NewStringSet(),
	}

	mutateErr := actor.Mutate(context.Background(), item)
	if mutateErr == nil {
		d.createdID = item.GetKeystoneID()
	} else {
		return resp.WithError(mutateErr)
	}

	item.TheList.Add("a")
	mutateErr = actor.Mutate(context.Background(), item)
	if mutateErr != nil {
		return resp.WithError(mutateErr)
	}

	if len(item.TheList.ToAdd()) > 0 {
		//return resp.WithError(errors.New("ToAdd not cleared after mutate"))
	}

	item.TheList.Add("b")
	mutateErr = actor.Mutate(context.Background(), item)
	if mutateErr != nil {
		return resp.WithError(mutateErr)
	}

	if len(item.TheList.Values()) != 2 {
		return resp.WithError(fmt.Errorf("values not correct after mutate: %s", strings.Join(item.TheList.Values(), ",")))
	}

	if len(item.TheList.Diff("a", "b")) > 0 {
		return resp.WithError(errors.New("diff not correct after mutate"))
	}

	item2 := &Settable{}
	getErr := actor.GetByID(context.Background(), item.GetKeystoneID(), item2, keystone.WithProperties("the_list"))
	if getErr != nil {
		return resp.WithError(getErr)
	}

	if len(item2.TheList.Diff("a", "b")) > 0 {
		return resp.WithError(errors.New("diff not correct after retrieve"))
	}

	return resp
}
