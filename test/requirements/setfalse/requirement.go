package setfalse

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Todo struct {
	keystone.BaseEntity
	ID        string `keystone:"_entity_id" json:"id"`
	Title     string `json:"title"`
	Details   string `json:"details"`
	Completed bool   `json:"completed"`
}

type Requirement struct {
	createdID string
}

func (d *Requirement) Name() string {
	return "Set False Value"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Person{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.create(actor),
		d.complete(actor),
		d.inProgress(actor),
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {

	item := &Todo{
		Title:   "Check Tests",
		Details: "Check the tests are running",
	}

	testErr := actor.Mutate(context.Background(), item)
	if testErr == nil {
		d.createdID = item.GetKeystoneID()
	}

	if testErr == nil && d.currentState(actor).Title != item.Title {
		testErr = errors.New("item title incorrect")
	}

	return requirements.TestResult{
		Name:  "Create",
		Error: testErr,
	}
}

func (d *Requirement) complete(actor *keystone.Actor) requirements.TestResult {
	var testErr error

	item := &Todo{Completed: true}
	item.SetKeystoneID(d.createdID)
	testErr = actor.MutateWithDefaultWatcher(context.Background(), item)

	if testErr == nil && d.currentState(actor).Completed != item.Completed {
		testErr = errors.New("item complete state incorrect")
	}

	return requirements.TestResult{
		Name:  "Complete Todo Item",
		Error: testErr,
	}
}
func (d *Requirement) inProgress(actor *keystone.Actor) requirements.TestResult {
	var testErr error

	item := &Todo{Completed: false}
	item.SetKeystoneID(d.createdID)
	testErr = actor.Mutate(context.Background(), item, keystone.MutateProperties("completed"))

	if testErr == nil && d.currentState(actor).Completed != item.Completed {
		testErr = errors.New("item complete state incorrect")
	}

	return requirements.TestResult{
		Name:  "Set completed to in-progress",
		Error: testErr,
	}
}

func (d *Requirement) currentState(actor *keystone.Actor) *Todo {
	item := &Todo{}
	_ = actor.Get(context.Background(), keystone.ByEntityID(item, d.createdID), item, keystone.WithProperties())
	return item
}
