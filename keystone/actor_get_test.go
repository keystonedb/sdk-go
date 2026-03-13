package keystone

import (
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

// retrieveTestEntity is a test entity type for testing retrieve options
type retrieveTestEntity struct {
	Name  string
	Email string
}

// TestRetrieveByEntityID tests the ByEntityID retriever
func TestRetrieveByEntityID(t *testing.T) {
	entityID := ID("retrieve-test-entity-123")
	retrieveBy := ByEntityID(&retrieveTestEntity{}, entityID)

	if retrieveBy.EntityType() != "retrieve-test-entity" {
		t.Errorf("ByEntityID().EntityType() = %s; want %s", retrieveBy.EntityType(), "retrieve-test-entity")
	}

	req := retrieveBy.BaseRequest()
	if req.GetEntityId() != "retrieve-test-entity-123" {
		t.Errorf("ByEntityID().BaseRequest().EntityId = %s; want %s", req.GetEntityId(), "retrieve-test-entity-123")
	}

	if req.View == nil {
		t.Error("ByEntityID().BaseRequest().View should not be nil")
	}
}

// TestRetrieveByHashID tests the ByHashID retriever
func TestRetrieveByHashID(t *testing.T) {
	retrieveBy := ByHashID(&retrieveTestEntity{}, "my-unique-string")

	if retrieveBy.EntityType() != "retrieve-test-entity" {
		t.Errorf("ByHashID().EntityType() = %s; want %s", retrieveBy.EntityType(), "retrieve-test-entity")
	}

	req := retrieveBy.BaseRequest()
	expectedHashID, err := HashID("my-unique-string")
	if err != nil {
		t.Fatalf("Failed to create hash ID: %v", err)
	}
	if req.GetEntityId() != string(expectedHashID) {
		t.Errorf("ByHashID().BaseRequest().EntityId = %s; want %s", req.GetEntityId(), string(expectedHashID))
	}
}

// TestRetrieveByUniqueProperty tests the ByUniqueProperty retriever
func TestRetrieveByUniqueProperty(t *testing.T) {
	retrieveBy := ByUniqueProperty(&retrieveTestEntity{}, "user@example.com", "email")

	if retrieveBy.EntityType() != "retrieve-test-entity" {
		t.Errorf("ByUniqueProperty().EntityType() = %s; want %s", retrieveBy.EntityType(), "retrieve-test-entity")
	}

	req := retrieveBy.BaseRequest()
	uniqueId := req.GetUniqueId()
	if uniqueId == nil {
		t.Fatal("ByUniqueProperty().BaseRequest().UniqueId should not be nil")
	}

	if uniqueId.GetSchemaId() != "retrieve-test-entity" {
		t.Errorf("UniqueId.SchemaId = %s; want %s", uniqueId.GetSchemaId(), "retrieve-test-entity")
	}

	if uniqueId.GetProperty() != "email" {
		t.Errorf("UniqueId.Property = %s; want %s", uniqueId.GetProperty(), "email")
	}

	if uniqueId.GetUniqueId() != "user@example.com" {
		t.Errorf("UniqueId.UniqueId = %s; want %s", uniqueId.GetUniqueId(), "user@example.com")
	}
}

// TestWithProperties tests the WithProperties option
func TestWithProperties(t *testing.T) {
	opt := WithProperties("name", "email", "phone")
	view := &proto.EntityView{}

	opt.Apply(view)

	if len(view.Properties) != 1 {
		t.Fatalf("WithProperties should add 1 PropertyRequest, got %d", len(view.Properties))
	}

	props := view.Properties[0].GetProperties()
	if len(props) != 3 {
		t.Errorf("Expected 3 properties, got %d", len(props))
	}

	expectedProps := []string{"name", "email", "phone"}
	for i, expected := range expectedProps {
		if props[i] != expected {
			t.Errorf("Property[%d] = %s; want %s", i, props[i], expected)
		}
	}

	if view.Properties[0].GetDecrypt() {
		t.Error("WithProperties should not set decrypt to true")
	}
}

// TestWithDecryptedProperties tests the WithDecryptedProperties option
func TestWithDecryptedProperties(t *testing.T) {
	opt := WithDecryptedProperties("ssn", "credit_card")
	view := &proto.EntityView{}

	opt.Apply(view)

	if len(view.Properties) != 1 {
		t.Fatalf("WithDecryptedProperties should add 1 PropertyRequest, got %d", len(view.Properties))
	}

	if !view.Properties[0].GetDecrypt() {
		t.Error("WithDecryptedProperties should set decrypt to true")
	}

	props := view.Properties[0].GetProperties()
	if len(props) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(props))
	}
}

// TestWithPropertyDecryptFlag tests the WithProperty option with decrypt flag
func TestWithPropertyDecryptFlag(t *testing.T) {
	tests := []struct {
		name    string
		decrypt bool
		props   []string
	}{
		{"decrypt true", true, []string{"password"}},
		{"decrypt false", false, []string{"username"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithProperty(tt.decrypt, tt.props...)
			view := &proto.EntityView{}

			opt.Apply(view)

			if view.Properties[0].GetDecrypt() != tt.decrypt {
				t.Errorf("Decrypt = %v; want %v", view.Properties[0].GetDecrypt(), tt.decrypt)
			}
		})
	}
}

// TestWithRelationships tests the WithRelationships option
func TestWithRelationships(t *testing.T) {
	opt := WithRelationships("friends", "followers", "following")
	view := &proto.EntityView{}

	opt.Apply(view)

	if len(view.RelationshipByType) != 3 {
		t.Fatalf("WithRelationships should add 3 keys, got %d", len(view.RelationshipByType))
	}

	expectedKeys := []string{"friends", "followers", "following"}
	for i, expected := range expectedKeys {
		if view.RelationshipByType[i].GetKey() != expected {
			t.Errorf("RelationshipByType[%d].Key = %s; want %s", i, view.RelationshipByType[i].GetKey(), expected)
		}
	}
}

// TestWithLabels tests the WithLabels option
func TestWithLabels(t *testing.T) {
	opt := WithLabels()
	view := &proto.EntityView{}

	opt.Apply(view)

	if !view.Labels {
		t.Error("WithLabels should set Labels to true")
	}
}

// TestWithView tests the WithView option
func TestWithView(t *testing.T) {
	opt := WithView("full_profile")
	view := &proto.EntityView{}

	opt.Apply(view)

	if view.Name != "full_profile" {
		t.Errorf("WithView().Name = %s; want %s", view.Name, "full_profile")
	}
}

// TestWithSummary tests the WithSummary option
func TestWithSummary(t *testing.T) {
	opt := WithSummary()
	view := &proto.EntityView{}

	opt.Apply(view)

	if !view.Summary {
		t.Error("WithSummary should set Summary to true")
	}
}

// TestWithDocument tests the WithDocument option
func TestWithDocument(t *testing.T) {
	doc := &Document{}
	opt := WithDocument(doc)
	view := &proto.EntityView{}

	opt.Apply(view)

	if !view.LatestDocument {
		t.Error("WithDocument should set LatestDocument to true")
	}

	if view.DocumentRevision != "" {
		t.Error("WithDocument should not set DocumentRevision")
	}
}

// TestWithDocumentRevision tests the WithDocumentRevision option
func TestWithDocumentRevision(t *testing.T) {
	doc := &Document{}
	opt := WithDocumentRevision("rev-123", doc)
	view := &proto.EntityView{}

	opt.Apply(view)

	if view.LatestDocument {
		t.Error("WithDocumentRevision should not set LatestDocument to true")
	}

	if view.DocumentRevision != "rev-123" {
		t.Errorf("WithDocumentRevision().DocumentRevision = %s; want %s", view.DocumentRevision, "rev-123")
	}
}

// TestWithDocumentRevisionList tests the WithDocumentRevisionList option
func TestWithDocumentRevisionList(t *testing.T) {
	opt := WithDocumentRevisionList()
	view := &proto.EntityView{}

	opt.Apply(view)

	if !view.DocumentRevisions {
		t.Error("WithDocumentRevisionList should set DocumentRevisions to true")
	}
}

// TestWithTotalRelationshipCount tests the WithTotalRelationshipCount option
func TestWithTotalRelationshipCount(t *testing.T) {
	opt := WithTotalRelationshipCount()
	view := &proto.EntityView{}

	opt.Apply(view)

	if !view.RelationshipCount {
		t.Error("WithTotalRelationshipCount should set RelationshipCount to true")
	}
}

// TestWithRelationshipCount tests the WithRelationshipCount option
func TestWithRelationshipCount(t *testing.T) {
	tests := []struct {
		name           string
		relationType   string
		appId          string
		vendorId       string
		wantTotal      bool
		wantTypesCount int
	}{
		{
			name:           "empty params returns total count",
			relationType:   "",
			appId:          "",
			vendorId:       "",
			wantTotal:      true,
			wantTypesCount: 0,
		},
		{
			name:           "with relation type",
			relationType:   "friends",
			appId:          "app1",
			vendorId:       "vendor1",
			wantTotal:      false,
			wantTypesCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithRelationshipCount(tt.relationType, tt.appId, tt.vendorId)
			view := &proto.EntityView{}

			opt.Apply(view)

			if view.RelationshipCount != tt.wantTotal {
				t.Errorf("RelationshipCount = %v; want %v", view.RelationshipCount, tt.wantTotal)
			}

			if len(view.RelationshipCountType) != tt.wantTypesCount {
				t.Errorf("len(RelationshipCountType) = %d; want %d", len(view.RelationshipCountType), tt.wantTypesCount)
			}

			if tt.wantTypesCount > 0 {
				key := view.RelationshipCountType[0]
				if key.GetKey() != tt.relationType {
					t.Errorf("RelationshipCountType[0].Key = %s; want %s", key.GetKey(), tt.relationType)
				}
				if key.GetSource().GetAppId() != tt.appId {
					t.Errorf("RelationshipCountType[0].Source.AppId = %s; want %s", key.GetSource().GetAppId(), tt.appId)
				}
				if key.GetSource().GetVendorId() != tt.vendorId {
					t.Errorf("RelationshipCountType[0].Source.VendorId = %s; want %s", key.GetSource().GetVendorId(), tt.vendorId)
				}
			}
		})
	}
}

// TestWithSiblingRelationshipCount tests the WithSiblingRelationshipCount option
func TestWithSiblingRelationshipCount(t *testing.T) {
	opt := WithSiblingRelationshipCount("siblings")
	view := &proto.EntityView{}

	opt.Apply(view)

	if len(view.RelationshipCountType) != 1 {
		t.Fatalf("WithSiblingRelationshipCount should add 1 key, got %d", len(view.RelationshipCountType))
	}

	key := view.RelationshipCountType[0]
	if key.GetKey() != "siblings" {
		t.Errorf("RelationshipCountType[0].Key = %s; want %s", key.GetKey(), "siblings")
	}

	// Sibling relationship count should have empty source
	if key.GetSource().GetAppId() != "" {
		t.Errorf("RelationshipCountType[0].Source.AppId should be empty, got %s", key.GetSource().GetAppId())
	}
}

// TestWithChildSummary tests the WithChildSummary option
func TestWithChildSummary(t *testing.T) {
	opt := WithChildSummary()
	view := &proto.EntityView{}

	opt.Apply(view)

	if !view.ChildSummary {
		t.Error("WithChildSummary should set ChildSummary to true")
	}
}

// TestWithDescendantCount tests the WithDescendantCount option
func TestWithDescendantCount(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		wantCount  int
	}{
		{
			name:       "with entity type",
			entityType: "Comment",
			wantCount:  1,
		},
		{
			name:       "empty entity type returns nil",
			entityType: "",
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithDescendantCount(tt.entityType)
			view := &proto.EntityView{}

			if opt != nil {
				opt.Apply(view)
			}

			if len(view.DescendantCountType) != tt.wantCount {
				t.Errorf("len(DescendantCountType) = %d; want %d", len(view.DescendantCountType), tt.wantCount)
			}

			if tt.wantCount > 0 {
				if view.DescendantCountType[0].GetKey() != tt.entityType {
					t.Errorf("DescendantCountType[0].Key = %s; want %s", view.DescendantCountType[0].GetKey(), tt.entityType)
				}
			}
		})
	}
}

// TestWithLock tests the WithLock option
func TestWithLock(t *testing.T) {
	opt := WithLock("Processing order", 60)
	view := &proto.EntityView{}
	req := &proto.EntityRequest{View: view}

	// Apply to view (no-op for lock)
	opt.Apply(view)

	// Apply to request
	if reOpt, ok := opt.(RetrieveEntityOption); ok {
		reOpt.ApplyRequest(req)
	}

	if !req.RequestLock {
		t.Error("WithLock should set RequestLock to true")
	}

	if req.LockTtlSeconds != 60 {
		t.Errorf("WithLock().LockTtlSeconds = %d; want %d", req.LockTtlSeconds, 60)
	}

	if req.LockMessage != "Processing order" {
		t.Errorf("WithLock().LockMessage = %s; want %s", req.LockMessage, "Processing order")
	}
}

// TestWithVerifiedProperty tests the WithVerifiedProperty option
func TestWithVerifiedProperty(t *testing.T) {
	opt := WithVerifiedProperty("password", "hashed_password_value")
	view := &proto.EntityView{}
	req := &proto.EntityRequest{View: view}

	// Apply to view
	opt.Apply(view)

	if len(view.Properties) != 1 {
		t.Fatalf("WithVerifiedProperty should add 1 property request, got %d", len(view.Properties))
	}

	if view.Properties[0].GetProperties()[0] != "password" {
		t.Errorf("Property = %s; want %s", view.Properties[0].GetProperties()[0], "password")
	}

	// Apply to request
	if reOpt, ok := opt.(RetrieveEntityOption); ok {
		reOpt.ApplyRequest(req)
	}

	if len(req.VerifyProperties) != 1 {
		t.Fatalf("WithVerifiedProperty should add 1 verify property, got %d", len(req.VerifyProperties))
	}

	verifyProp := req.VerifyProperties[0]
	if verifyProp.GetProperty() != "password" {
		t.Errorf("VerifyProperty.Property = %s; want %s", verifyProp.GetProperty(), "password")
	}

	if verifyProp.GetValue().GetSecureText() != "hashed_password_value" {
		t.Errorf("VerifyProperty.Value.SecureText = %s; want %s", verifyProp.GetValue().GetSecureText(), "hashed_password_value")
	}
}

// TestWithObjects tests the WithObjects option
func TestWithObjects(t *testing.T) {
	tests := []struct {
		name            string
		paths           []string
		wantListObjects bool
		wantPathsCount  int
	}{
		{
			name:            "no paths lists all objects",
			paths:           []string{},
			wantListObjects: true,
			wantPathsCount:  0,
		},
		{
			name:            "with specific paths",
			paths:           []string{"avatar.png", "documents/resume.pdf"},
			wantListObjects: false,
			wantPathsCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithObjects(tt.paths...)
			view := &proto.EntityView{}

			opt.Apply(view)

			if view.ListObjects != tt.wantListObjects {
				t.Errorf("ListObjects = %v; want %v", view.ListObjects, tt.wantListObjects)
			}

			if len(view.ObjectPaths) != tt.wantPathsCount {
				t.Errorf("len(ObjectPaths) = %d; want %d", len(view.ObjectPaths), tt.wantPathsCount)
			}

			for i, path := range tt.paths {
				if view.ObjectPaths[i] != path {
					t.Errorf("ObjectPaths[%d] = %s; want %s", i, view.ObjectPaths[i], path)
				}
			}
		})
	}
}

// TestRetrieveOptionsComposition tests combining multiple retrieve options
func TestRetrieveOptionsComposition(t *testing.T) {
	opts := RetrieveOptions(
		WithProperties("name", "email"),
		WithLabels(),
		WithSummary(),
		WithView("full"),
	)

	view := &proto.EntityView{}
	opts.Apply(view)

	if len(view.Properties) != 1 {
		t.Errorf("Expected 1 property request, got %d", len(view.Properties))
	}

	if !view.Labels {
		t.Error("Labels should be true")
	}

	if !view.Summary {
		t.Error("Summary should be true")
	}

	if view.Name != "full" {
		t.Errorf("Name = %s; want %s", view.Name, "full")
	}
}

// TestDocumentLoaderObserveRetrieve tests the document loader's ObserveRetrieve method
func TestDocumentLoaderObserveRetrieve(t *testing.T) {
	doc := &Document{}
	loader := documentLoader{doc: doc, revisionID: ""}

	resp := &proto.EntityResponse{
		Documents: []*proto.EntityDocument{
			{
				RevisionId: "rev-abc",
				Data:       []byte(`{"key": "value"}`),
				Meta:       map[string]string{"author": "test"},
			},
		},
	}

	loader.ObserveRetrieve(resp)

	if doc.RevisionID != "rev-abc" {
		t.Errorf("Document.RevisionID = %s; want %s", doc.RevisionID, "rev-abc")
	}

	if string(doc.Data) != `{"key": "value"}` {
		t.Errorf("Document.Data = %s; want %s", string(doc.Data), `{"key": "value"}`)
	}

	if doc.Meta["author"] != "test" {
		t.Errorf("Document.Meta[author] = %s; want %s", doc.Meta["author"], "test")
	}
}

// TestDocumentLoaderObserveRetrieveWithRevision tests loading a specific revision
func TestDocumentLoaderObserveRetrieveWithRevision(t *testing.T) {
	doc := &Document{}
	loader := documentLoader{doc: doc, revisionID: "rev-xyz"}

	resp := &proto.EntityResponse{
		Documents: []*proto.EntityDocument{
			{
				RevisionId: "rev-abc",
				Data:       []byte(`{"wrong": "doc"}`),
			},
			{
				RevisionId: "rev-xyz",
				Data:       []byte(`{"right": "doc"}`),
			},
		},
	}

	loader.ObserveRetrieve(resp)

	if doc.RevisionID != "rev-xyz" {
		t.Errorf("Document.RevisionID = %s; want %s", doc.RevisionID, "rev-xyz")
	}

	if string(doc.Data) != `{"right": "doc"}` {
		t.Errorf("Document.Data = %s; want %s", string(doc.Data), `{"right": "doc"}`)
	}
}

// TestDocumentLoaderObserveRetrieveNilDoc tests that nil document is handled gracefully
func TestDocumentLoaderObserveRetrieveNilDoc(t *testing.T) {
	loader := documentLoader{doc: nil}

	resp := &proto.EntityResponse{
		Documents: []*proto.EntityDocument{
			{
				RevisionId: "rev-abc",
				Data:       []byte(`{"key": "value"}`),
			},
		},
	}

	// Should not panic
	loader.ObserveRetrieve(resp)
}

// TestMultiplePropertyRequests tests adding multiple property requests
func TestMultiplePropertyRequests(t *testing.T) {
	view := &proto.EntityView{}

	WithProperties("name", "email").Apply(view)
	WithDecryptedProperties("ssn").Apply(view)
	WithProperty(false, "phone").Apply(view)

	if len(view.Properties) != 3 {
		t.Errorf("Expected 3 property requests, got %d", len(view.Properties))
	}

	// First request: name, email (not decrypted)
	if view.Properties[0].GetDecrypt() {
		t.Error("First request should not be decrypted")
	}

	// Second request: ssn (decrypted)
	if !view.Properties[1].GetDecrypt() {
		t.Error("Second request should be decrypted")
	}

	// Third request: phone (not decrypted)
	if view.Properties[2].GetDecrypt() {
		t.Error("Third request should not be decrypted")
	}
}

// TestMultipleRelationships tests adding relationships incrementally
func TestMultipleRelationships(t *testing.T) {
	view := &proto.EntityView{}

	WithRelationships("friends").Apply(view)
	WithRelationships("followers", "following").Apply(view)

	if len(view.RelationshipByType) != 3 {
		t.Errorf("Expected 3 relationships, got %d", len(view.RelationshipByType))
	}

	expectedKeys := []string{"friends", "followers", "following"}
	for i, expected := range expectedKeys {
		if view.RelationshipByType[i].GetKey() != expected {
			t.Errorf("RelationshipByType[%d].Key = %s; want %s", i, view.RelationshipByType[i].GetKey(), expected)
		}
	}
}

// TestWithObjectsImplementsRetrieveEntityOption tests that withObjects implements RetrieveEntityOption
func TestWithObjectsImplementsRetrieveEntityOption(t *testing.T) {
	opt := WithObjects("test.png")

	// Verify it implements RetrieveEntityOption
	if _, ok := opt.(RetrieveEntityOption); !ok {
		t.Error("WithObjects should implement RetrieveEntityOption")
	}

	// ApplyRequest should be a no-op but not panic
	req := &proto.EntityRequest{}
	if reOpt, ok := opt.(RetrieveEntityOption); ok {
		reOpt.ApplyRequest(req)
	}
}
