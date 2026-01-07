package keystone

import "github.com/keystonedb/sdk-go/proto"

type DocumentObserver interface {
	addDocuments(documents ...*proto.EntityDocument)
	setRevisions(revisions []string)
}

type EmbeddedDocuments struct {
	documentRevisions []string
	documents         []*proto.EntityDocument
}

func (e *EmbeddedDocuments) GetDocumentRevisions() []string {
	return e.documentRevisions
}

func (e *EmbeddedDocuments) ClearDocumentRevisions() {
	e.documentRevisions = []string{}
}

func (e *EmbeddedDocuments) GetDocuments() []*proto.EntityDocument {
	return e.documents
}

func (e *EmbeddedDocuments) LatestDocument() *proto.EntityDocument {
	if len(e.documents) == 0 {
		return nil
	}
	var latest *proto.EntityDocument
	for _, doc := range e.documents {
		if latest == nil || doc.GetCreated().AsTime().After(latest.GetCreated().AsTime()) {
			latest = doc
		}
	}

	return latest
}

func (e *EmbeddedDocuments) addDocuments(documents ...*proto.EntityDocument) {
	e.documents = documents
}

func (e *EmbeddedDocuments) setRevisions(revisions []string) {
	e.documentRevisions = revisions
}
