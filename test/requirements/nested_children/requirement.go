package nested_children

import (
	"context"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	fileID string
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
	}
}

func (d *Requirement) writeChildren(actor *keystone.Actor) requirements.TestResult {

	sub := &models.File{
		FileName: "testfile",
	}

	sub.AddChild(&models.FileLine{
		Data:       []byte("test data"),
		LineNumber: 1}, sub.Lines)

	sub.AddChild(&models.FileLine{
		Data:       []byte("line two"),
		LineNumber: 2}, sub.Lines)

	sub.AddChild(&models.FileLine{
		Data:       []byte("third one"),
		LineNumber: 3}, sub.Lines)

	createErr := actor.Mutate(context.Background(), sub, keystone.WithMutationComment("Create a file"))
	if createErr == nil {
		d.fileID = sub.GetKeystoneID()
	}

	return requirements.TestResult{
		Name:  "Create File with Nested Lines",
		Error: createErr,
	}
}
