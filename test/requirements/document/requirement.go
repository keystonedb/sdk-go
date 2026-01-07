package document

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

type Requirement struct {
	EntityID          keystone.ID
	InitialRevisionID string
	RevisionIDs       []string
	LatestRevisionID  string
}

func (d *Requirement) Name() string {
	return "Entity Documents"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.createEntity(actor),
		d.checkNoDocuments(actor),
		d.createDocument(actor),
		d.checkInitialDocument(actor),
		d.newRevision(actor),
		d.checkRevision(actor),
		d.updateRevision(actor),
		d.loadRevision(actor),
	}
}

func (d *Requirement) createEntity(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create Entity"}
	e := models.User{
		ExternalID: "document-holder-" + k4id.New().String(),
	}
	err := actor.Mutate(context.Background(), &e)
	d.EntityID = e.GetKeystoneID()
	return res.WithError(err)
}

func (d *Requirement) checkNoDocuments(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Check No Documents"}
	e := models.User{}
	err := actor.GetByID(context.Background(), d.EntityID, &e, keystone.WithDocumentRevisionList(), keystone.WithDocument(nil))
	if err != nil {
		return res.WithError(err)
	}
	if len(e.GetDocumentRevisions()) != 0 {
		return res.WithError(errors.New("expected no document revisions"))
	}
	if e.LatestDocument() != nil {
		return res.WithError(errors.New("expected no latest document"))
	}
	if len(e.GetDocuments()) > 0 {
		return res.WithError(errors.New("expected no documents"))
	}
	return res
}

func (d *Requirement) createDocument(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create Document"}

	e := models.User{}
	e.SetKeystoneID(d.EntityID)

	doc := keystone.NewDocument([]byte("Version 1"), nil)
	err := actor.Mutate(context.Background(), &e, doc)
	if err != nil {
		return res.WithError(err)
	}

	if doc.RevisionID == "" {
		return res.WithError(errors.New("expected document revision ID to be set"))
	}
	d.InitialRevisionID = doc.RevisionID
	d.RevisionIDs = append(d.RevisionIDs, doc.RevisionID)

	return res
}

func (d *Requirement) checkInitialDocument(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Check Initial Document"}

	e := models.User{}
	initial := &keystone.Document{}
	err := actor.GetByID(context.Background(), d.EntityID, &e, keystone.WithDocumentRevisionList(), keystone.WithDocument(initial))
	if err != nil {
		return res.WithError(err)
	}
	if len(e.GetDocumentRevisions()) != 1 {
		return res.WithError(errors.New("expected one document revision"))
	}
	if e.LatestDocument() == nil {
		return res.WithError(errors.New("expected a latest document"))
	}
	if len(e.GetDocuments()) != 1 {
		return res.WithError(errors.New("expected one document"))
	}

	if initial.RevisionID != d.InitialRevisionID {
		return res.WithError(errors.New("initial document revision ID does not match"))
	}

	if string(initial.Data) != "Version 1" {
		return res.WithError(errors.New("initial document data does not match"))
	}

	if string(e.LatestDocument().GetData()) != "Version 1" {
		return res.WithError(errors.New("latest document data does not match"))
	}

	return res
}

func (d *Requirement) newRevision(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create Revision"}

	e := models.User{}
	e.SetKeystoneID(d.EntityID)

	doc := keystone.NewDocument([]byte("Version 2"), map[string]string{"author": "harry"})
	err := actor.Mutate(context.Background(), &e, doc)
	if err != nil {
		return res.WithError(err)
	}

	if doc.RevisionID == "" {
		return res.WithError(errors.New("expected document revision ID to be set"))
	}
	d.LatestRevisionID = doc.RevisionID
	d.RevisionIDs = append(d.RevisionIDs, doc.RevisionID)

	return res
}

func (d *Requirement) checkRevision(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Check Revised Document"}

	e := models.User{}
	revised := &keystone.Document{}
	err := actor.GetByID(context.Background(), d.EntityID, &e, keystone.WithDocumentRevisionList(), keystone.WithDocument(revised))
	if err != nil {
		return res.WithError(err)
	}
	if len(e.GetDocumentRevisions()) != 2 {
		return res.WithError(errors.New("expected two document revisions"))
	}
	if e.LatestDocument() == nil {
		return res.WithError(errors.New("expected a latest document"))
	}
	if len(e.GetDocuments()) != 1 {
		return res.WithError(errors.New("expected latest document"))
	}

	if revised.RevisionID != d.LatestRevisionID {
		return res.WithError(errors.New("latest document revision ID does not match"))
	}

	if string(revised.Data) != "Version 2" {
		return res.WithError(errors.New("latest document data does not match"))
	}

	if string(e.LatestDocument().GetData()) != "Version 2" {
		return res.WithError(errors.New("latest document data does not match"))
	}

	if revised.Meta == nil {
		return res.WithError(errors.New("expected metadata to be present"))
	} else if revised.Meta["author"] != "harry" {
		return res.WithError(errors.New("metadata author does not match"))
	}

	return res
}

func (d *Requirement) updateRevision(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Update Revision"}

	e := models.User{}
	e.SetKeystoneID(d.EntityID)

	doc := keystone.UpdateDocument(d.LatestRevisionID)
	doc.AppendMeta("approved", "john")
	doc.AppendMeta("approval", "now")
	err := actor.Mutate(context.Background(), &e, doc)
	if err != nil {
		return res.WithError(err)
	}

	if doc.RevisionID == "" || doc.RevisionID != d.LatestRevisionID {
		return res.WithError(errors.New("expected document revision ID to be set"))
	}

	doc2 := keystone.UpdateDocument(d.LatestRevisionID)
	doc2.RemoveMeta("approval")
	err = actor.Mutate(context.Background(), &e, doc2)
	if err != nil {
		return res.WithError(err)
	}

	revised := &keystone.Document{}
	err = actor.GetByID(context.Background(), d.EntityID, &e, keystone.WithDocument(revised))

	if revised.Meta == nil {
		return res.WithError(errors.New("expected metadata to be present"))
	} else if revised.Meta["author"] != "harry" {
		return res.WithError(errors.New("original meta corrupt"))
	} else if revised.Meta["approved"] != "john" {
		return res.WithError(errors.New("appended meta not present"))
	}

	_, ok := revised.Meta["approval"]
	if ok {
		return res.WithError(errors.New("removed meta still present"))
	}

	if string(revised.Data) != "Version 2" {
		return res.WithError(errors.New("latest document looks to have become corrupt"))
	}

	return res
}

func (d *Requirement) loadRevision(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Load Revision"}

	e := models.User{}
	initial := &keystone.Document{}
	err := actor.GetByID(context.Background(), d.EntityID, &e, keystone.WithDocumentRevision(d.InitialRevisionID, initial))
	if err != nil {
		return res.WithError(err)
	}
	if e.LatestDocument() == nil {
		return res.WithError(errors.New("expected a latest document"))
	}
	if len(e.GetDocuments()) != 1 {
		return res.WithError(errors.New("expected one document"))
	}

	if initial.RevisionID != d.InitialRevisionID {
		return res.WithError(errors.New("initial document revision ID does not match"))
	}

	if string(initial.Data) != "Version 1" {
		return res.WithError(errors.New("initial document data does not match"))
	}

	if string(e.LatestDocument().GetData()) != "Version 1" {
		return res.WithError(errors.New("latest document data does not match"))
	}

	return res
}
