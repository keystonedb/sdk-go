package keystone

import (
	"testing"
	"time"

	"github.com/keystonedb/sdk-go/proto"
)

func TestWithMutationComment(t *testing.T) {
	opt := WithMutationComment("test comment")
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if req.Mutation.Comment != "test comment" {
		t.Errorf("Expected comment 'test comment', got '%s'", req.Mutation.Comment)
	}
}

func TestOnConflictUseID(t *testing.T) {
	opt := OnConflictUseID("email", "username")
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if len(req.ConflictUniquePropertyAcquire) != 2 {
		t.Errorf("Expected 2 conflict properties, got %d", len(req.ConflictUniquePropertyAcquire))
	}
	if req.ConflictUniquePropertyAcquire[0] != "email" {
		t.Errorf("Expected first property 'email', got '%s'", req.ConflictUniquePropertyAcquire[0])
	}
	if req.ConflictUniquePropertyAcquire[1] != "username" {
		t.Errorf("Expected second property 'username', got '%s'", req.ConflictUniquePropertyAcquire[1])
	}
}

func TestOnConflictUseIDSingleProperty(t *testing.T) {
	opt := OnConflictUseID("id")
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if len(req.ConflictUniquePropertyAcquire) != 1 {
		t.Errorf("Expected 1 conflict property, got %d", len(req.ConflictUniquePropertyAcquire))
	}
	if req.ConflictUniquePropertyAcquire[0] != "id" {
		t.Errorf("Expected property 'id', got '%s'", req.ConflictUniquePropertyAcquire[0])
	}
}

func TestOnConflictIgnore(t *testing.T) {
	opt := OnConflictIgnore()
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if len(req.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(req.Options))
	}
	if req.Options[0] != proto.MutateRequest_OnConflictIgnore {
		t.Errorf("Expected OnConflictIgnore option, got %v", req.Options[0])
	}
}

func TestMutateProperties(t *testing.T) {
	opt := MutateProperties("name", "email")
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{
			Properties: []*proto.EntityProperty{
				{Property: "name", Value: &proto.Value{Text: "John"}},
				{Property: "email", Value: &proto.Value{Text: "john@example.com"}},
				{Property: "age", Value: &proto.Value{Int: 30}},
				{Property: "status", Value: &proto.Value{Text: "active"}},
			},
		},
	}

	opt.apply(req)

	if len(req.Mutation.Properties) != 2 {
		t.Errorf("Expected 2 properties after filter, got %d", len(req.Mutation.Properties))
	}

	// Verify only the specified properties remain
	propNames := make(map[string]bool)
	for _, p := range req.Mutation.Properties {
		propNames[p.Property] = true
	}

	if !propNames["name"] {
		t.Error("Expected 'name' property to be kept")
	}
	if !propNames["email"] {
		t.Error("Expected 'email' property to be kept")
	}
	if propNames["age"] {
		t.Error("Did not expect 'age' property to be kept")
	}
	if propNames["status"] {
		t.Error("Did not expect 'status' property to be kept")
	}
}

func TestMutatePropertiesEmpty(t *testing.T) {
	opt := MutateProperties()
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{
			Properties: []*proto.EntityProperty{
				{Property: "name", Value: &proto.Value{Text: "John"}},
				{Property: "email", Value: &proto.Value{Text: "john@example.com"}},
			},
		},
	}

	opt.apply(req)

	if len(req.Mutation.Properties) != 0 {
		t.Errorf("Expected 0 properties when no properties specified, got %d", len(req.Mutation.Properties))
	}
}

func TestMutatePropertiesPrepare(t *testing.T) {
	opt := MutateProperties("name", "email")
	mp, ok := opt.(mutateProperties)
	if !ok {
		t.Fatal("Expected mutateProperties type")
	}

	w := &Watcher{
		knownValues: map[string]*watcherValue{
			"name":   {Value: &proto.Value{Text: "John"}},
			"email":  {Value: &proto.Value{Text: "john@example.com"}},
			"age":    {Value: &proto.Value{Int: 30}},
			"status": {Value: &proto.Value{Text: "active"}},
		},
	}

	err := mp.prepare(w)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// The prepare method should delete the specified properties from known values
	if _, exists := w.knownValues["name"]; exists {
		t.Error("Expected 'name' to be deleted from known values")
	}
	if _, exists := w.knownValues["email"]; exists {
		t.Error("Expected 'email' to be deleted from known values")
	}
	if _, exists := w.knownValues["age"]; !exists {
		t.Error("Expected 'age' to remain in known values")
	}
	if _, exists := w.knownValues["status"]; !exists {
		t.Error("Expected 'status' to remain in known values")
	}
}

func TestMatchExisting(t *testing.T) {
	opt := MatchExisting(WhereEquals("status", "active"))
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if len(req.Where) != 1 {
		t.Errorf("Expected 1 where filter, got %d", len(req.Where))
	}
	if req.Where[0].Property != "status" {
		t.Errorf("Expected property 'status', got '%s'", req.Where[0].Property)
	}
	if req.Where[0].Operator != proto.Operator_Equal {
		t.Errorf("Expected Equal operator, got %v", req.Where[0].Operator)
	}
}

func TestMatchExistingMultipleFilters(t *testing.T) {
	opt := MatchExisting(
		WhereEquals("status", "active"),
		WhereGreaterThan("age", 18),
	)
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if len(req.Where) != 2 {
		t.Errorf("Expected 2 where filters, got %d", len(req.Where))
	}
}

func TestPrepareUploads(t *testing.T) {
	obj1, err := NewUpload("path/to/file1.txt", proto.ObjectType_Standard)
	if err != nil {
		t.Fatalf("Failed to create upload: %v", err)
	}
	obj2, err := NewUpload("path/to/file2.txt", proto.ObjectType_NearLine)
	if err != nil {
		t.Fatalf("Failed to create upload: %v", err)
	}
	obj2.SetPublic(true)

	opt := PrepareUploads(obj1, obj2)
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if len(req.Mutation.Objects) != 2 {
		t.Errorf("Expected 2 objects, got %d", len(req.Mutation.Objects))
	}

	if req.Mutation.Objects[0].Path != "path/to/file1.txt" {
		t.Errorf("Expected path 'path/to/file1.txt', got '%s'", req.Mutation.Objects[0].Path)
	}
	if req.Mutation.Objects[0].Type != proto.ObjectType_Standard {
		t.Errorf("Expected Standard type, got %v", req.Mutation.Objects[0].Type)
	}
	if req.Mutation.Objects[0].Public {
		t.Error("Expected first object to not be public")
	}

	if req.Mutation.Objects[1].Path != "path/to/file2.txt" {
		t.Errorf("Expected path 'path/to/file2.txt', got '%s'", req.Mutation.Objects[1].Path)
	}
	if req.Mutation.Objects[1].Type != proto.ObjectType_NearLine {
		t.Errorf("Expected NearLine type, got %v", req.Mutation.Objects[1].Type)
	}
	if !req.Mutation.Objects[1].Public {
		t.Error("Expected second object to be public")
	}
}

func TestPrepareUploadsWithExpiry(t *testing.T) {
	obj, err := NewUpload("path/to/file.txt", proto.ObjectType_Standard)
	if err != nil {
		t.Fatalf("Failed to create upload: %v", err)
	}
	expiry := time.Now().Add(24 * time.Hour)
	obj.SetExpiry(expiry)

	opt := PrepareUploads(obj)
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if len(req.Mutation.Objects) != 1 {
		t.Errorf("Expected 1 object, got %d", len(req.Mutation.Objects))
	}

	if req.Mutation.Objects[0].Expiry == nil {
		t.Error("Expected expiry to be set")
	}
	if req.Mutation.Objects[0].Expiry.AsTime().Unix() != expiry.Unix() {
		t.Errorf("Expiry time mismatch: expected %v, got %v", expiry, req.Mutation.Objects[0].Expiry.AsTime())
	}
}

func TestPrepareUploadsWithData(t *testing.T) {
	obj, err := NewUpload("path/to/file.txt", proto.ObjectType_Standard)
	if err != nil {
		t.Fatalf("Failed to create upload: %v", err)
	}
	data := []byte("test data content")
	obj.SetData(data)

	opt := PrepareUploads(obj)
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if len(req.Mutation.Objects) != 1 {
		t.Errorf("Expected 1 object, got %d", len(req.Mutation.Objects))
	}

	if string(req.Mutation.Objects[0].Data) != "test data content" {
		t.Errorf("Expected data 'test data content', got '%s'", string(req.Mutation.Objects[0].Data))
	}
}

func TestPrepareUploadsObserveMutation(t *testing.T) {
	obj, err := NewUpload("path/to/file.txt", proto.ObjectType_Standard)
	if err != nil {
		t.Fatalf("Failed to create upload: %v", err)
	}
	opt := PrepareUploads(obj)
	po := opt.(prepareObjects)

	// Test successful response
	response := &proto.MutateResponse{
		Success: true,
		SignedObjectUrls: []*proto.EntityObject{
			{
				Path:          "path/to/file.txt",
				Url:           "https://example.com/upload/123",
				UploadHeaders: map[string]string{"Content-Type": "application/octet-stream"},
			},
		},
	}

	po.ObserveMutation(response)

	if obj.uploadURL != "https://example.com/upload/123" {
		t.Errorf("Expected upload URL to be set, got '%s'", obj.uploadURL)
	}
	if obj.uploadHeaders["Content-Type"] != "application/octet-stream" {
		t.Errorf("Expected Content-Type header, got %v", obj.uploadHeaders)
	}
}

func TestPrepareUploadsObserveMutationFailure(t *testing.T) {
	obj, err := NewUpload("path/to/file.txt", proto.ObjectType_Standard)
	if err != nil {
		t.Fatalf("Failed to create upload: %v", err)
	}
	opt := PrepareUploads(obj)
	po := opt.(prepareObjects)

	// Test failed response
	response := &proto.MutateResponse{
		Success: false,
	}

	po.ObserveMutation(response)

	if obj.uploadURL != "" {
		t.Errorf("Expected upload URL to remain empty on failure, got '%s'", obj.uploadURL)
	}
}

func TestWithPiiToken(t *testing.T) {
	opt := WithPiiToken("pii-token-123")
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if req.Mutation.PiiToken != "pii-token-123" {
		t.Errorf("Expected PII token 'pii-token-123', got '%s'", req.Mutation.PiiToken)
	}
}

func TestWithPiiReference(t *testing.T) {
	opt := WithPiiReference("vendor1", "app1", "pii-key-abc")
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if req.Mutation.PiiReference == nil {
		t.Fatal("Expected PII reference to be set")
	}
	if req.Mutation.PiiReference.Key != "pii-key-abc" {
		t.Errorf("Expected PII key 'pii-key-abc', got '%s'", req.Mutation.PiiReference.Key)
	}
	if req.Mutation.PiiReference.Source == nil {
		t.Fatal("Expected PII reference source to be set")
	}
	if req.Mutation.PiiReference.Source.VendorId != "vendor1" {
		t.Errorf("Expected vendor ID 'vendor1', got '%s'", req.Mutation.PiiReference.Source.VendorId)
	}
	if req.Mutation.PiiReference.Source.AppId != "app1" {
		t.Errorf("Expected app ID 'app1', got '%s'", req.Mutation.PiiReference.Source.AppId)
	}
}

func TestBackgroundIndex(t *testing.T) {
	opt := BackgroundIndex()
	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{},
	}

	opt.apply(req)

	if len(req.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(req.Options))
	}
	if req.Options[0] != proto.MutateRequest_BackgroundIndex {
		t.Errorf("Expected BackgroundIndex option, got %v", req.Options[0])
	}
}

func TestMultipleMutateOptions(t *testing.T) {
	comment := WithMutationComment("update user")
	bgIndex := BackgroundIndex()
	onConflict := OnConflictIgnore()

	req := &proto.MutateRequest{
		Mutation: &proto.Mutation{
			Properties: []*proto.EntityProperty{
				{Property: "name", Value: &proto.Value{Text: "John"}},
			},
		},
	}

	comment.apply(req)
	bgIndex.apply(req)
	onConflict.apply(req)

	if req.Mutation.Comment != "update user" {
		t.Errorf("Expected comment 'update user', got '%s'", req.Mutation.Comment)
	}
	if len(req.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(req.Options))
	}

	hasBackgroundIndex := false
	hasOnConflictIgnore := false
	for _, opt := range req.Options {
		if opt == proto.MutateRequest_BackgroundIndex {
			hasBackgroundIndex = true
		}
		if opt == proto.MutateRequest_OnConflictIgnore {
			hasOnConflictIgnore = true
		}
	}

	if !hasBackgroundIndex {
		t.Error("Expected BackgroundIndex option to be present")
	}
	if !hasOnConflictIgnore {
		t.Error("Expected OnConflictIgnore option to be present")
	}
}

func TestMutateToError(t *testing.T) {
	tests := []struct {
		name        string
		resp        *proto.MutateResponse
		err         error
		expectError bool
		errorCode   int32
	}{
		{
			name:        "nil response",
			resp:        nil,
			err:         nil,
			expectError: true,
		},
		{
			name: "successful response",
			resp: &proto.MutateResponse{
				Success:      true,
				ErrorCode:    0,
				ErrorMessage: "",
			},
			err:         nil,
			expectError: false,
		},
		{
			name: "error response with code",
			resp: &proto.MutateResponse{
				Success:      false,
				ErrorCode:    400,
				ErrorMessage: "Bad request",
			},
			err:         nil,
			expectError: true,
			errorCode:   400,
		},
		{
			name: "error response with message only",
			resp: &proto.MutateResponse{
				Success:      false,
				ErrorCode:    0,
				ErrorMessage: "Something went wrong",
			},
			err:         nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mutateToError(tt.resp, tt.err)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Did not expect error but got: %v", err)
			}
			if tt.errorCode > 0 {
				if ksErr, ok := err.(*Error); ok {
					if ksErr.ErrorCode != tt.errorCode {
						t.Errorf("Expected error code %d, got %d", tt.errorCode, ksErr.ErrorCode)
					}
				}
			}
		})
	}
}

func TestMutationOptionImplementsInterface(t *testing.T) {
	// Verify all options implement MutateOption interface
	var _ MutateOption = WithMutationComment("test")
	var _ MutateOption = OnConflictUseID("id")
	var _ MutateOption = OnConflictIgnore()
	var _ MutateOption = MutateProperties("prop")
	var _ MutateOption = MatchExisting()
	var _ MutateOption = PrepareUploads()
	var _ MutateOption = WithPiiToken("token")
	var _ MutateOption = WithPiiReference("vendor", "app", "key")
	var _ MutateOption = BackgroundIndex()
}

func TestMutatePropertiesImplementsWatcherPrepare(t *testing.T) {
	// Verify MutateProperties implements MutationOptionWatcherPrepare
	opt := MutateProperties("name")
	_, ok := opt.(MutationOptionWatcherPrepare)
	if !ok {
		t.Error("Expected MutateProperties to implement MutationOptionWatcherPrepare")
	}
}

func TestPrepareObjectsImplementsMutationObserver(t *testing.T) {
	// Verify PrepareUploads implements MutationObserver
	opt := PrepareUploads()
	_, ok := opt.(MutationObserver)
	if !ok {
		t.Error("Expected PrepareUploads to implement MutationObserver")
	}
}

func TestWithState(t *testing.T) {
	tests := []struct {
		name  string
		state proto.EntityState
	}{
		{"Active", proto.EntityState_Active},
		{"Offline", proto.EntityState_Offline},
		{"Corrupt", proto.EntityState_Corrupt},
		{"Archived", proto.EntityState_Archived},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithState(tt.state)
			req := &proto.MutateRequest{
				Mutation: &proto.Mutation{},
			}

			opt.apply(req)

			if req.Mutation.State != tt.state {
				t.Errorf("Expected state %v, got %v", tt.state, req.Mutation.State)
			}
		})
	}
}

func TestWithStateCanTransitionBetweenStates(t *testing.T) {
	transitions := []struct {
		name string
		from proto.EntityState
		to   proto.EntityState
	}{
		{"Active to Archived", proto.EntityState_Active, proto.EntityState_Archived},
		{"Archived to Active", proto.EntityState_Archived, proto.EntityState_Active},
		{"Active to Offline", proto.EntityState_Active, proto.EntityState_Offline},
		{"Offline to Active", proto.EntityState_Offline, proto.EntityState_Active},
		{"Active to Corrupt", proto.EntityState_Active, proto.EntityState_Corrupt},
		{"Corrupt to Active", proto.EntityState_Corrupt, proto.EntityState_Active},
		{"Archived to Offline", proto.EntityState_Archived, proto.EntityState_Offline},
		{"Offline to Archived", proto.EntityState_Offline, proto.EntityState_Archived},
	}

	for _, tt := range transitions {
		t.Run(tt.name, func(t *testing.T) {
			req := &proto.MutateRequest{
				Mutation: &proto.Mutation{State: tt.from},
			}

			opt := WithState(tt.to)
			opt.apply(req)

			if req.Mutation.State != tt.to {
				t.Errorf("Expected state to transition from %v to %v, got %v", tt.from, tt.to, req.Mutation.State)
			}
		})
	}
}

func TestWithStateForbidsInvalidState(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected WithState to panic for EntityState_Invalid")
		}
	}()

	WithState(proto.EntityState_Invalid)
}

func TestWithStateForbidsRemovedState(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected WithState to panic for EntityState_Removed")
		}
	}()

	WithState(proto.EntityState_Removed)
}

func TestWithStateImplementsInterface(t *testing.T) {
	var _ MutateOption = WithState(proto.EntityState_Active)
}
