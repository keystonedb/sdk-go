package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func newAKVWorkspaceTestActor(t *testing.T) (*Actor, *MockServer, func()) {
	t.Helper()
	conn, mock, _, server := MockConnection()
	go func() { _ = server.Serve(mockListener) }()
	actor := conn.Actor("workspace-1", "127.0.0.1", "user-1", "go-test")
	return &actor, mock, func() {
		server.Stop()
		_ = mockListener.Close()
	}
}

func TestActor_AKVWorkspaceMethods_NilActor(t *testing.T) {
	var actor *Actor
	ctx := context.Background()

	if _, err := actor.AKVWorkspacePut(ctx, AKV("key", "value")); err == nil {
		t.Error("AKVWorkspacePut() expected an error")
	}
	if _, err := actor.AKVWorkspaceGet(ctx, "key"); err == nil {
		t.Error("AKVWorkspaceGet() expected an error")
	}
	if _, err := actor.AKVWorkspaceDel(ctx, "key"); err == nil {
		t.Error("AKVWorkspaceDel() expected an error")
	}
}

func TestActor_AKVWorkspaceMethods_UseActorWorkspace(t *testing.T) {
	actor, mock, cleanup := newAKVWorkspaceTestActor(t)
	defer cleanup()

	mock.AKVWorkspacePutFunc = func(_ context.Context, req *proto.AKVPutRequest) (*proto.GenericResponse, error) {
		if got := req.GetAuthorization().GetWorkspaceId(); got != "workspace-1" {
			t.Errorf("AKVWorkspacePut() workspace = %q, want workspace-1", got)
		}
		if got := req.GetProperties()[0].GetValue().GetText(); got != "value" {
			t.Errorf("AKVWorkspacePut() value = %q, want value", got)
		}
		return &proto.GenericResponse{Success: true}, nil
	}
	putResp, err := actor.AKVWPut(context.Background(), AKV("key", "value"))
	if err != nil || !putResp.GetSuccess() {
		t.Fatalf("AKVWPut() resp=%v err=%v", putResp, err)
	}

	mock.AKVWorkspaceGetFunc = func(_ context.Context, req *proto.AKVGetRequest) (*proto.AKVGetResponse, error) {
		if got := req.GetAuthorization().GetWorkspaceId(); got != "workspace-1" {
			t.Errorf("AKVWorkspaceGet() workspace = %q, want workspace-1", got)
		}
		if got := req.GetProperties(); len(got) != 1 || got[0] != "key" {
			t.Errorf("AKVWorkspaceGet() properties = %v, want [key]", got)
		}
		return &proto.AKVGetResponse{
			Summary:    &proto.GenericResponse{Success: true},
			Properties: map[string]*proto.Value{"key": {Text: "value"}},
		}, nil
	}
	values, err := actor.AKVWGet(context.Background(), "key")
	if err != nil || values["key"].GetText() != "value" {
		t.Fatalf("AKVWGet() values=%v err=%v", values, err)
	}

	mock.AKVWorkspaceDelFunc = func(_ context.Context, req *proto.AKVDelRequest) (*proto.GenericResponse, error) {
		if got := req.GetAuthorization().GetWorkspaceId(); got != "workspace-1" {
			t.Errorf("AKVWorkspaceDel() workspace = %q, want workspace-1", got)
		}
		if got := req.GetProperties(); len(got) != 1 || got[0] != "key" {
			t.Errorf("AKVWorkspaceDel() properties = %v, want [key]", got)
		}
		return &proto.GenericResponse{Success: true}, nil
	}
	delResp, err := actor.AKVWDel(context.Background(), "key")
	if err != nil || !delResp.GetSuccess() {
		t.Fatalf("AKVWDel() resp=%v err=%v", delResp, err)
	}
}
