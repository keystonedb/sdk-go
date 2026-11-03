package keystone

import (
	"context"
	"errors"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
	pb "google.golang.org/protobuf/proto"
)

type typedDynamicSettings struct {
	DynamicRecordEntity
	Enabled bool
	Retries int64
	Name    string `keystone:"display_name"`
	Secret  SecureString
	Nested  *typedDynamicNested
}
type typedDynamicNested struct {
	Endpoint string
	Enabled  bool
}

func TestDynamicRecordEntityMutationAndTracking(t *testing.T) {
	actor, mock, cleanup := newAKVWorkspaceTestActor(t)
	defer cleanup()
	ctx := context.Background()
	settings := &typedDynamicSettings{DynamicRecordEntity: NewDynamicRecordEntity("settings"), Secret: NewSecureString("secret", "****"), Nested: &typedDynamicNested{Endpoint: "example"}}
	calls := 0
	mock.MutateDynamicRecordFunc = func(_ context.Context, req *proto.MutateDynamicRecordRequest) (*proto.DynamicRecordMutationResponse, error) {
		calls++
		if !pb.Equal(req.Authorization, actor.Authorization()) || req.Key != "settings" {
			t.Fatal("scope mismatch")
		}
		switch calls {
		case 1:
			if len(req.Set) != 6 || req.Set["enabled"].KnownType != proto.Property_Boolean || req.Set["enabled"].Bool || req.Set["retries"].KnownType != proto.Property_Number || req.Set["display_name"].Text != "" || req.Set["secret"].SecureText != "secret" || req.Set["nested.endpoint"].Text != "example" {
				t.Fatalf("lost typed values: %v", req)
			}
		case 2:
			if len(req.Set) != 1 || req.Set["enabled"] == nil || !req.Set["enabled"].Bool || req.Comment != "enable" {
				t.Fatalf("expected one changed property: %v", req)
			}
		case 3:
			if len(req.Set) != 1 || req.Set["enabled"] == nil || req.Set["enabled"].Bool {
				t.Fatalf("false must remain a value: %v", req)
			}
		case 4:
			if len(req.Set) != 1 || req.Set["retries"] == nil || req.Set["retries"].Int != 0 {
				t.Fatalf("explicit zero must be written: %v", req)
			}
		}
		return &proto.DynamicRecordMutationResponse{Revision: "saved"}, nil
	}
	if err := actor.Mutate(ctx, settings); err != nil {
		t.Fatal(err)
	}
	if !settings.DynamicRecordExists() || settings.DynamicRecordRevision() != "saved" {
		t.Fatal("mutation metadata missing")
	}
	if err := actor.Mutate(ctx, settings); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("unchanged entity should not write")
	}
	settings.Enabled = true
	if err := actor.Mutate(ctx, settings, WithMutationComment("enable")); err != nil {
		t.Fatal(err)
	}
	settings.Enabled = false
	if err := actor.Mutate(ctx, settings); err != nil {
		t.Fatal(err)
	}
	settings.Name = "pending"
	if err := actor.Mutate(ctx, settings, MutateProperties("retries")); err != nil {
		t.Fatal(err)
	}
	if calls != 4 {
		t.Fatalf("unexpected calls: %d", calls)
	}
	if len(actor.connection.registerQueue) != 0 {
		t.Fatal("dynamic record registered a schema")
	}
}
func TestDynamicRecordEntityLoadSelectionAndMissing(t *testing.T) {
	actor, mock, cleanup := newAKVWorkspaceTestActor(t)
	defer cleanup()
	ctx := context.Background()
	settings := &typedDynamicSettings{Enabled: true, Retries: 99, Name: "stale", Nested: &typedDynamicNested{Endpoint: "keep"}}
	mock.GetDynamicRecordFunc = func(_ context.Context, req *proto.GetDynamicRecordRequest) (*proto.DynamicRecord, error) {
		if req.Key != "settings" || len(req.Properties) != 2 {
			t.Fatalf("selection lost: %v", req)
		}
		return &proto.DynamicRecord{Key: req.Key, Exists: true, Revision: "loaded", Properties: map[string]*proto.Value{"enabled": {KnownType: proto.Property_Boolean, Bool: false}}}, nil
	}
	if err := actor.GetDynamicRecord(ctx, "settings", settings, "enabled", "display_name"); err != nil {
		t.Fatal(err)
	}
	if settings.Enabled || settings.Name != "" || settings.Retries != 99 || settings.Nested.Endpoint != "keep" {
		t.Fatalf("partial hydration failed: %+v", settings)
	}
	if settings.GetDynamicRecordKey() != "settings" || !settings.DynamicRecordExists() || settings.DynamicRecordRevision() != "loaded" {
		t.Fatal("load metadata lost")
	}
	settings.Enabled = true
	mock.MutateDynamicRecordFunc = func(_ context.Context, req *proto.MutateDynamicRecordRequest) (*proto.DynamicRecordMutationResponse, error) {
		if len(req.Set) != 1 || req.Set["enabled"] == nil {
			t.Fatalf("unselected fields were written: %v", req)
		}
		return &proto.DynamicRecordMutationResponse{Revision: "saved"}, nil
	}
	if err := actor.Mutate(ctx, settings); err != nil {
		t.Fatal(err)
	}
	mock.GetDynamicRecordFunc = func(_ context.Context, req *proto.GetDynamicRecordRequest) (*proto.DynamicRecord, error) {
		return &proto.DynamicRecord{Key: req.Key}, nil
	}
	if err := actor.DynamicRecord("missing").Load(ctx, settings); err != nil {
		t.Fatal(err)
	}
	if settings.DynamicRecordExists() || settings.Enabled || settings.Retries != 0 || settings.Nested != nil || settings.Name != "" {
		t.Fatalf("missing record retained stale data: %+v", settings)
	}
}
func TestDynamicRecordEntityFailedMutationRetainsChanges(t *testing.T) {
	actor, mock, cleanup := newAKVWorkspaceTestActor(t)
	defer cleanup()
	ctx := context.Background()
	settings := &typedDynamicSettings{DynamicRecordEntity: NewDynamicRecordEntity("settings")}
	mock.MutateDynamicRecordFunc = func(context.Context, *proto.MutateDynamicRecordRequest) (*proto.DynamicRecordMutationResponse, error) {
		return nil, errors.New("failed")
	}
	if err := actor.Mutate(ctx, settings); err == nil {
		t.Fatal("expected error")
	}
	if settings.DynamicRecordExists() || settings.DynamicRecordRevision() != "" {
		t.Fatal("failed mutation changed metadata")
	}
	mock.MutateDynamicRecordFunc = func(_ context.Context, r *proto.MutateDynamicRecordRequest) (*proto.DynamicRecordMutationResponse, error) {
		if len(r.Set) == 0 {
			t.Fatal("retry lost changes")
		}
		return &proto.DynamicRecordMutationResponse{}, nil
	}
	if err := actor.Mutate(ctx, settings); err != nil {
		t.Fatal(err)
	}
	if err := actor.Mutate(ctx, settings, OnConflictIgnore()); err == nil {
		t.Fatal("unsupported option must fail")
	}
	if err := actor.Mutate(ctx, settings, MutateProperties("unknown")); err == nil {
		t.Fatal("unknown selection must fail")
	}
	var nilEntity *typedDynamicSettings
	if err := actor.Mutate(ctx, nilEntity); err == nil {
		t.Fatal("expected nil error")
	}
	settings.SetDynamicRecordKey("other")
	if settings.DynamicRecordExists() || settings.DynamicRecordRevision() != "" {
		t.Fatal("key change retained metadata")
	}
}

func TestDynamicRecordEntitySkipsSchemaRegistration(t *testing.T) {
	actor, _, cleanup := newAKVWorkspaceTestActor(t)
	defer cleanup()
	if got := actor.connection.RegisterTypes(typedDynamicSettings{}, &typedDynamicSettings{}); got != 0 {
		t.Fatalf("registered %d record types", got)
	}
	if len(actor.connection.registerQueue) != 0 {
		t.Fatal("record entered schema queue")
	}
	settings := &typedDynamicSettings{}
	if err := actor.GetByID(context.Background(), "settings", settings); err == nil {
		t.Fatal("normal entity lookup should reject record structs before schema registration")
	}
	if err := actor.MutateWithDefaultWatcher(context.Background(), settings); err == nil {
		t.Fatal("external watcher should be rejected")
	}
}

func TestDynamicRecordEntityNestedSelectionAndRemoval(t *testing.T) {
	actor, mock, cleanup := newAKVWorkspaceTestActor(t)
	defer cleanup()
	ctx := context.Background()
	nested := &typedDynamicNested{Endpoint: "old", Enabled: true}
	settings := &typedDynamicSettings{Nested: nested}
	mock.GetDynamicRecordFunc = func(context.Context, *proto.GetDynamicRecordRequest) (*proto.DynamicRecord, error) {
		return &proto.DynamicRecord{Exists: true, Properties: map[string]*proto.Value{"nested.endpoint": {Text: "loaded"}}}, nil
	}
	if err := actor.GetDynamicRecord(ctx, "settings", settings, "nested.endpoint"); err != nil {
		t.Fatal(err)
	}
	if settings.Nested.Endpoint != "loaded" || !settings.Nested.Enabled {
		t.Fatal("nested selection overwrote sibling")
	}
	if nested.Endpoint != "old" {
		t.Fatal("load mutated an aliased nested struct")
	}
	settings.Nested = nil
	mock.MutateDynamicRecordFunc = func(_ context.Context, r *proto.MutateDynamicRecordRequest) (*proto.DynamicRecordMutationResponse, error) {
		if len(r.Set) != 0 || len(r.Remove) != 2 {
			t.Fatalf("expected removed nested leaves: %v", r)
		}
		return &proto.DynamicRecordMutationResponse{Revision: "removed"}, nil
	}
	if err := actor.Mutate(ctx, settings); err != nil {
		t.Fatal(err)
	}
}
