package nested_children

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	fileID   string
	child1ID string
	child2ID string
	child3ID string
}

func (d *Requirement) Name() string {
	return "Nested Child"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.File{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.writeChildren(actor),
		d.removeChild(actor),
	}
}

func (d *Requirement) writeChildren(actor *keystone.Actor) requirements.TestResult {

	sub := &models.File{
		FileName: "testfile",
	}

	f1 := &models.FileLine{
		Data:       []byte("test data"),
		LineNumber: 1,
	}
	f2 := &models.FileLine{
		Data:       []byte("line two"),
		LineNumber: 2,
	}
	f3 := &models.FileLine{
		Data:       []byte("third one"),
		LineNumber: 3,
	}

	sub.Lines = append(sub.Lines, f1, f2, f3)
	sub.AddChildren(sub.Lines)

	createErr := actor.Mutate(context.Background(), sub, keystone.WithMutationComment("Create a file"))
	if createErr == nil {

		d.fileID = sub.GetKeystoneID()
		d.child1ID = f1.ChildID()
		d.child2ID = f2.ChildID()
		d.child3ID = f3.ChildID()

		if d.child1ID == "" || d.child2ID == "" || d.child3ID == "" {
			createErr = errors.New("failed to create children, or children did not return a child ID")
		}
	}

	return requirements.TestResult{
		Name:  "Create File with Nested Lines",
		Error: createErr,
	}
}

func (d *Requirement) removeChild(actor *keystone.Actor) requirements.TestResult {

	sub := &models.File{}
	sub.SetKeystoneID(d.fileID)
	sub.RemoveChild(models.FileLine{}, d.child2ID)

	removeErr := actor.Mutate(context.Background(), sub, keystone.WithMutationComment("Remove a line"))
	return requirements.TestResult{
		Name:  "Remove a line from the file",
		Error: removeErr,
	}
}
