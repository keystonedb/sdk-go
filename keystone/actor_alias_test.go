package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	gproto "google.golang.org/protobuf/proto"
	"testing"
)

func TestAliasRPCRequestsAndErrors(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	actor.connection.appID = proto.VendorApp{VendorId: "test-vendor", AppId: "test-app"}
	schema := &proto.Key{Source: &proto.VendorApp{VendorId: "owner", AppId: "customers"}, Key: "customer"}
	alias := NewEntityAlias("billing", "stripe", "cus_123").WithRouting("eu")
	ctx := context.Background()
	var failure error
	check := func(auth *proto.Authorization, key *proto.Key) {
		t.Helper()
		if !gproto.Equal(auth, actor.Authorization()) || !gproto.Equal(key, schema) {
			t.Errorf("scope not preserved: %v %v", auth, key)
		}
	}
	mutation := func(_ context.Context, req *proto.AliasMutationRequest) (*proto.GenericResponse, error) {
		check(req.Authorization, req.Schema)
		if req.EntityId != "customer123" || !gproto.Equal(req.Alias, alias) {
			t.Errorf("mutation: %v", req)
		}
		return &proto.GenericResponse{Success: true}, failure
	}
	mock.AttachAliasFunc = mutation
	mock.RemoveAliasFunc = mutation
	mock.ResolveAliasFunc = func(_ context.Context, req *proto.ResolveAliasRequest) (*proto.ResolveAliasResponse, error) {
		check(req.Authorization, req.Schema)
		if !gproto.Equal(req.Alias, alias) {
			t.Errorf("alias changed: %v", req)
		}
		return &proto.ResolveAliasResponse{EntityId: "customer123"}, failure
	}
	mock.ListAliasesFunc = func(_ context.Context, req *proto.ListAliasesRequest) (*proto.ListAliasesResponse, error) {
		check(req.Authorization, req.Schema)
		if req.EntityId != "customer123" {
			t.Errorf("entity: %v", req)
		}
		return &proto.ListAliasesResponse{Aliases: []*EntityAlias{alias}}, failure
	}
	calls := []func() error{
		func() error {
			return actor.AttachAlias(ctx, schema.Key, "customer123", alias, WithAliasOwner("owner", "customers"))
		},
		func() error {
			return actor.RemoveAlias(ctx, schema.Key, "customer123", alias, WithAliasOwner("owner", "customers"))
		},
		func() error {
			id, err := actor.ResolveAlias(ctx, schema.Key, alias, WithAliasOwner("owner", "customers"))
			if err == nil && id != "customer123" {
				t.Error(id)
			}
			return err
		},
		func() error {
			rows, err := actor.ListAliases(ctx, schema.Key, "customer123", WithAliasOwner("owner", "customers"))
			if err == nil && (len(rows) != 1 || !gproto.Equal(rows[0], alias)) {
				t.Error(rows)
			}
			return err
		},
	}
	for _, call := range calls {
		if err := call(); err != nil {
			t.Fatal(err)
		}
	}
	for _, code := range []codes.Code{codes.NotFound, codes.AlreadyExists, codes.PermissionDenied, codes.InvalidArgument, codes.Unavailable, codes.Canceled, codes.DeadlineExceeded} {
		failure = status.Error(code, "test")
		for _, call := range calls {
			if err := call(); status.Code(err) != code {
				t.Fatalf("error lost: %v", err)
			}
		}
	}
}
func TestAliasNilActors(t *testing.T) {
	for _, a := range []*Actor{nil, {}, {connection: &Connection{}}} {
		ctx := context.Background()
		if err := a.AttachAlias(ctx, "customer", "", nil); err == nil {
			t.Fatal("attach")
		}
		if err := a.RemoveAlias(ctx, "customer", "", nil); err == nil {
			t.Fatal("remove")
		}
		if _, err := a.ResolveAlias(ctx, "customer", nil); err == nil {
			t.Fatal("resolve")
		}
		if _, err := a.ListAliases(ctx, "customer", ""); err == nil {
			t.Fatal("list")
		}
		if err := a.GetByAlias(ctx, nil, nil); err == nil {
			t.Fatal("get")
		}
	}
}
func TestAliasUnsuccessfulResponses(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	actor.connection.appID = proto.VendorApp{VendorId: "test-vendor", AppId: "test-app"}
	ctx := context.Background()
	mock.AttachAliasFunc = func(context.Context, *proto.AliasMutationRequest) (*proto.GenericResponse, error) {
		return &proto.GenericResponse{}, nil
	}
	mock.RemoveAliasFunc = mock.AttachAliasFunc
	mock.ResolveAliasFunc = func(context.Context, *proto.ResolveAliasRequest) (*proto.ResolveAliasResponse, error) {
		return &proto.ResolveAliasResponse{}, nil
	}
	if err := actor.AttachAlias(ctx, "customer", "", nil); err == nil {
		t.Fatal("false success")
	}
	if err := actor.RemoveAlias(ctx, "customer", "", nil); err == nil {
		t.Fatal("false success")
	}
	if _, err := actor.ResolveAlias(ctx, "customer", nil); err == nil {
		t.Fatal("empty id")
	}
	if err := actor.GetByAlias(ctx, nil, &retrieveTestEntity{}); err == nil {
		t.Fatal("nil alias")
	}
	mock.ListAliasesFunc = func(context.Context, *proto.ListAliasesRequest) (*proto.ListAliasesResponse, error) {
		return &proto.ListAliasesResponse{}, nil
	}
	if rows, err := actor.ListAliases(ctx, "customer", ""); err != nil || len(rows) != 0 {
		t.Fatalf("empty: %v %v", rows, err)
	}
}
func TestGetByAliasAndLoader(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	actor.connection.appID = proto.VendorApp{VendorId: "test-vendor", AppId: "test-app"}
	alias := &EntityAlias{Source: &proto.VendorApp{VendorId: "vendor", AppId: "app"}, ExternalId: "123"}
	by := ByAlias(&retrieveTestEntity{}, alias)
	if by.EntityType() != Type(&retrieveTestEntity{}) {
		t.Fatal(by.EntityType())
	}
	req := by.BaseRequest()
	req.Alias.ExternalId = "changed"
	if alias.ExternalId != "123" || req.UniqueId != nil || req.EntityId != "" {
		t.Fatal("selector mutated")
	}
	mock.DefineFunc = func(_ context.Context, r *proto.SchemaRequest) (*proto.Schema, error) { return r.Schema, nil }
	calls := 0
	mock.RetrieveFunc = func(_ context.Context, r *proto.EntityRequest) (*proto.EntityResponse, error) {
		calls++
		if !gproto.Equal(r.Alias, alias) || !r.View.Aliases || r.Schema.Key != Type(&retrieveTestEntity{}) || !gproto.Equal(r.Schema.Source, actor.VendorApp()) {
			t.Errorf("request: %v", r)
		}
		result := &proto.EntityResponse{Properties: []*proto.EntityProperty{{Property: "name", Value: &proto.Value{Text: "Customer"}}}}
		if calls == 1 {
			result.Aliases = []*EntityAlias{alias}
		}
		return result, nil
	}
	loader := WithAliases()
	dst := &retrieveTestEntity{}
	if err := actor.GetByAlias(context.Background(), alias, dst, WithProperties("name"), loader); err != nil {
		t.Fatal(err)
	}
	if dst.Name != "Customer" || len(loader.Aliases()) != 1 {
		t.Fatalf("result: %v %v", dst, loader.Aliases())
	}
	if err := actor.GetByAlias(context.Background(), alias, dst, loader); err != nil {
		t.Fatal(err)
	}
	if len(loader.Aliases()) != 0 {
		t.Fatal("stale aliases retained")
	}
}

type nilAliasListClient struct{ proto.KeystoneClient }

func (nilAliasListClient) ListAliases(context.Context, *proto.ListAliasesRequest, ...grpc.CallOption) (*proto.ListAliasesResponse, error) {
	return nil, nil
}
func TestAliasNilListResponse(t *testing.T) {
	actor := NewConnection(nilAliasListClient{}, "vendor", "app", "token").Actor("ws", "", "", "")
	if _, err := actor.ListAliases(context.Background(), "customer", ""); err == nil {
		t.Fatal("nil list response accepted")
	}
}

func TestAliasTypeDefaultsAndValidation(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	actor.connection.appID = proto.VendorApp{VendorId: "test-vendor", AppId: "test-app"}
	ctx := context.Background()
	alias := &EntityAlias{Source: &proto.VendorApp{VendorId: "billing", AppId: "stripe"}, ExternalId: "123"}
	entityType := Type(&retrieveTestEntity{})
	check := func(key *proto.Key) {
		t.Helper()
		if key.GetKey() != entityType || !gproto.Equal(key.Source, actor.VendorApp()) {
			t.Errorf("default owner/type: %v", key)
		}
	}
	mock.AttachAliasFunc = func(_ context.Context, r *proto.AliasMutationRequest) (*proto.GenericResponse, error) {
		check(r.Schema)
		return &proto.GenericResponse{Success: true}, nil
	}
	mock.RemoveAliasFunc = mock.AttachAliasFunc
	mock.ResolveAliasFunc = func(_ context.Context, r *proto.ResolveAliasRequest) (*proto.ResolveAliasResponse, error) {
		check(r.Schema)
		return &proto.ResolveAliasResponse{EntityId: "entity123"}, nil
	}
	mock.ListAliasesFunc = func(_ context.Context, r *proto.ListAliasesRequest) (*proto.ListAliasesResponse, error) {
		check(r.Schema)
		return &proto.ListAliasesResponse{}, nil
	}
	if err := actor.AttachAlias(ctx, entityType, "entity123", alias, nil); err != nil {
		t.Fatal(err)
	}
	if err := actor.RemoveAlias(ctx, entityType, "entity123", alias); err != nil {
		t.Fatal(err)
	}
	if _, err := actor.ResolveAlias(ctx, entityType, alias); err != nil {
		t.Fatal(err)
	}
	if _, err := actor.ListAliases(ctx, entityType, "entity123"); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		entityType string
		opts       []AliasOption
	}{{"", nil}, {entityType, []AliasOption{WithAliasOwner("", "app")}}, {entityType, []AliasOption{WithAliasOwner("vendor", "")}}} {
		if err := actor.AttachAlias(ctx, test.entityType, "entity123", alias, test.opts...); err == nil {
			t.Fatal("invalid attach scope")
		}
		if err := actor.RemoveAlias(ctx, test.entityType, "entity123", alias, test.opts...); err == nil {
			t.Fatal("invalid remove scope")
		}
		if _, err := actor.ResolveAlias(ctx, test.entityType, alias, test.opts...); err == nil {
			t.Fatal("invalid resolve scope")
		}
		if _, err := actor.ListAliases(ctx, test.entityType, "entity123", test.opts...); err == nil {
			t.Fatal("invalid list scope")
		}
	}
	if got := ByAlias(entityType, alias).EntityType(); got != entityType {
		t.Fatalf("string entity type was reflected again: %q", got)
	}
}

func TestNewEntityAlias(t *testing.T) {
	alias := NewEntityAlias("billing", "stripe", "cus_123")
	if alias.GetSource().GetVendorId() != "billing" || alias.GetSource().GetAppId() != "stripe" || alias.GetExternalId() != "cus_123" || alias.GetRoutingKey() != "" {
		t.Fatalf("alias: %v", alias)
	}
	if alias.WithRouting("eu") != alias || alias.GetRoutingKey() != "eu" {
		t.Fatal("routing is not chainable")
	}
	if alias.GetSource().GetVendorId() != "billing" || alias.GetSource().GetAppId() != "stripe" || alias.GetExternalId() != "cus_123" {
		t.Fatal("routing changed identity")
	}
	if alias.WithRouting("").GetRoutingKey() != "" {
		t.Fatal("default routing was not restored")
	}
	other := NewEntityAlias("billing", "stripe", "cus_123")
	alias.Source.AppId = "changed"
	if other.Source.AppId != "stripe" {
		t.Fatal("constructors share source state")
	}
}
