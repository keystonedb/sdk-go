package nested_children

import (
	"context"
	"errors"
	"fmt"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"sort"
)

type Requirement struct {
	fileID   keystone.ID
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
		d.summaryChildren(actor),
		d.removeChild(actor),
		d.loadChildren(actor),
		d.updateChildren(actor),
		d.verifyChildren(actor),
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

func (d *Requirement) summaryChildren(actor *keystone.Actor) requirements.TestResult {

	res := requirements.TestResult{
		Name: "Load the file with child summaries",
	}

	sub := &models.File{}
	getErr := actor.Get(context.Background(), keystone.ByEntityID(sub, d.fileID), sub, keystone.WithChildSummary())

	if sub.NumLines != 3 {
		return res.WithError(fmt.Errorf("expected 3 lines, got %d", sub.NumLines))
	}
	if sub.SumLines != 6 {
		return res.WithError(fmt.Errorf("expected 6 sum, got %d", sub.SumLines))
	}
	if sub.MinLines != 1 {
		return res.WithError(fmt.Errorf("expected 1 min, got %d", sub.MinLines))
	}
	if sub.MaxLines != 3 {
		return res.WithError(fmt.Errorf("expected 3 max, got %d", sub.MaxLines))
	}
	if sub.AvgLines != 2 {
		return res.WithError(fmt.Errorf("expected 2 avg, got %d", sub.AvgLines))
	}

	if getErr != nil {
		return res.WithError(getErr)
	}

	return res
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

func (d *Requirement) loadChildren(actor *keystone.Actor) requirements.TestResult {

	res := requirements.TestResult{
		Name: "Load the file with nested lines",
	}

	sub := &models.File{}
	lines := keystone.WithChildren(keystone.Type(models.FileLine{}))
	getErr := actor.Get(context.Background(), keystone.ByEntityID(sub, d.fileID), sub, &lines)
	sub.Lines = keystone.ChildrenFromLoader[models.FileLine](lines)
	sort.Sort(sub)

	if getErr != nil {
		return res.WithError(getErr)
	}
	if len(sub.Lines) != 2 {
		return res.WithError(fmt.Errorf("expected 2 lines, got %d", len(sub.Lines)))
	}
	if sub.Lines[0].LineNumber != 1 || sub.Lines[1].LineNumber != 3 {
		return res.WithError(fmt.Errorf("expected line numbers 1 and 3, got %d and %d", sub.Lines[0].LineNumber, sub.Lines[1].LineNumber))
	}
	if sub.Lines[0].ChildID() != d.child1ID {
		return res.WithError(errors.New("child 1 ID was not set"))
	}
	if sub.Lines[1].ChildID() != d.child3ID {
		return res.WithError(errors.New("child 3 ID was not set"))
	}
	if sub.Lines[1].ChildID() == d.child2ID {
		return res.WithError(errors.New("child 2 ID was not removed"))
	}
	if sub.Lines[0].Data == nil || sub.Lines[1].Data == nil {
		return res.WithError(errors.New("child data was not loaded"))
	}
	if string(sub.Lines[0].Data) != "test data" || string(sub.Lines[1].Data) != "third one" {
		return res.WithError(errors.New("child data was not correct"))
	}

	return res
}

func (d *Requirement) updateChildren(actor *keystone.Actor) requirements.TestResult {

	sub := &models.File{}
	sub.SetKeystoneID(d.fileID)
	chd := keystone.NewDynamicChild(models.FileLine{})
	chd.SetChildID(d.child1ID)
	chd.Append("data", []byte("updated data"))
	sub.AddChild(chd)
	writeErr := actor.Mutate(context.Background(), sub)

	return requirements.TestResult{
		Name:  "Update the file with nested lines",
		Error: writeErr,
	}
}

func (d *Requirement) verifyChildren(actor *keystone.Actor) requirements.TestResult {

	res := requirements.TestResult{
		Name: "Verify the file with nested lines",
	}

	sub := &models.File{}
	lines := keystone.WithChildren(keystone.Type(models.FileLine{}), d.child1ID)
	getErr := actor.Get(context.Background(), keystone.ByEntityID(sub, d.fileID), sub, &lines)
	sub.Lines = keystone.ChildrenFromLoader[models.FileLine](lines)

	if getErr != nil {
		return res.WithError(getErr)
	}

	if len(sub.Lines) != 1 {
		return res.WithError(fmt.Errorf("expected 1 line, got %d", len(sub.Lines)))
	}

	if sub.Lines[0].LineNumber != 1 {
		return res.WithError(fmt.Errorf("expected line number 1, got %d", sub.Lines[0].LineNumber))
	}

	if sub.Lines[0].ChildID() != d.child1ID {
		return res.WithError(errors.New("child 1 ID was not set"))
	}

	if sub.Lines[0].Data == nil {
		return res.WithError(errors.New("child data was not loaded"))
	}

	if string(sub.Lines[0].Data) != "updated data" {
		return res.WithError(errors.New("child data was not correct, got " + string(sub.Lines[0].Data)))
	}

	return res
}
