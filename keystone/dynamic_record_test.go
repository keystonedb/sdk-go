package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	pb "google.golang.org/protobuf/proto"
	"testing"
)

func TestDynamicRecordClient(t *testing.T) {
	actor, mock, cleanup := newAKVWorkspaceTestActor(t)
	defer cleanup()
	ctx := context.Background()
	record := actor.DynamicRecord("settings")
	check := func(auth *proto.Authorization, key string) {
		t.Helper()
		if !pb.Equal(auth, actor.Authorization()) || key != "settings" {
			t.Fatalf("lost scope: %v %s", auth, key)
		}
	}
	mock.MutateDynamicRecordFunc = func(_ context.Context, r *proto.MutateDynamicRecordRequest) (*proto.DynamicRecordMutationResponse, error) {
		check(r.Authorization, r.Key)
		if len(r.Set) != 3 || r.Set["enabled"].KnownType != proto.Property_Boolean || r.Set["enabled"].Bool || r.Set["retries"].KnownType != proto.Property_Number || r.Set["retries"].Int != 0 || r.Set["name"].Text != "" {
			t.Fatalf("lost zero values or types: %v", r)
		}
		if len(r.Remove) != 1 || r.Remove[0] != "old" || r.Comment != "update" {
			t.Fatalf("lost patch metadata: %v", r)
		}
		return &proto.DynamicRecordMutationResponse{Revision: "revision"}, nil
	}
	_, err := record.Mutate(ctx, DynamicRecordPatch{Set: map[string]any{"enabled": false, "retries": int64(0), "name": ""}, Remove: []string{"old"}, Comment: "update"})
	if err != nil {
		t.Fatal(err)
	}
	mock.GetDynamicRecordFunc = func(_ context.Context, r *proto.GetDynamicRecordRequest) (*proto.DynamicRecord, error) {
		check(r.Authorization, r.Key)
		if len(r.Properties) != 1 || r.Properties[0] != "enabled" {
			t.Fatal("selection lost")
		}
		return &proto.DynamicRecord{Key: r.Key, Exists: true, Properties: map[string]*proto.Value{"enabled": {KnownType: proto.Property_Boolean, Bool: false}}}, nil
	}
	r, err := record.Get(ctx, "enabled")
	if err != nil || !r.Exists || r.Properties["enabled"].Bool {
		t.Fatalf("read: %v %v", r, err)
	}
	mock.ListDynamicRecordHistoryFunc = func(_ context.Context, r *proto.ListDynamicRecordHistoryRequest) (*proto.ListDynamicRecordHistoryResponse, error) {
		check(r.Authorization, r.Key)
		if r.Limit != 10 || r.BeforeRevision != "cursor" {
			t.Fatal("history options lost")
		}
		return &proto.ListDynamicRecordHistoryResponse{NextBeforeRevision: "next"}, nil
	}
	h, err := record.History(ctx, DynamicRecordHistoryOptions{Limit: 10, BeforeRevision: "cursor"})
	if err != nil || h.NextBeforeRevision != "next" {
		t.Fatalf("history: %v %v", h, err)
	}
	mock.DeleteDynamicRecordFunc = func(_ context.Context, r *proto.DeleteDynamicRecordRequest) (*proto.DynamicRecordMutationResponse, error) {
		check(r.Authorization, r.Key)
		if r.Comment != "retired" {
			t.Fatal("comment lost")
		}
		return &proto.DynamicRecordMutationResponse{}, nil
	}
	if _, err := record.Delete(ctx, "retired"); err != nil {
		t.Fatal(err)
	}
}
func TestDynamicRecordInvalidValuesAndNilActor(t *testing.T) {
	var actor *Actor
	if _, err := actor.DynamicRecord("settings").Get(context.Background()); err == nil {
		t.Fatal("expected nil actor error")
	}
	for _, v := range []any{nil, (*string)(nil), make(chan int)} {
		if _, err := dynamicRecordValue(v); err == nil {
			t.Fatalf("expected conversion error for %T", v)
		}
	}
	raw := &proto.Value{KnownType: proto.Property_SecureText, SecureText: "secret"}
	converted, err := dynamicRecordValue(raw)
	if err != nil || !pb.Equal(raw, converted) {
		t.Fatal("raw conversion failed")
	}
	converted.SecureText = "changed"
	if raw.SecureText != "secret" {
		t.Fatal("raw value was mutated")
	}
}

func TestDynamicRecordSecureStringConversion(t *testing.T) {
	for _, original := range []string{"secret", ""} {
		v, err := dynamicRecordValue(NewSecureString(original, "****"))
		if err != nil || v.KnownType != proto.Property_SecureText || v.SecureText != original {
			t.Fatalf("secure conversion: %v %v", v, err)
		}
	}
}
