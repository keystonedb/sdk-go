package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc/metadata"
)

func Test_ActorNils(t *testing.T) {
	var actor *Actor

	// Test methods with nil checks
	if actor.Authorization() != nil {
		t.Errorf("Actor.Authorization() = %v; want nil", actor.Authorization())
	}

	// Clone should handle nil
	if actor.CloneWithoutWorkspace() != nil {
		t.Errorf("Actor.CloneWithoutWorkspace() = %v; want nil", actor.CloneWithoutWorkspace())
	}
}

func Test_Actor_CloneWithoutWorkspace(t *testing.T) {
	// Create a test connection
	conn := &Connection{
		appID: proto.VendorApp{
			VendorId: "test-vendor",
			AppId:    "test-app",
		},
		token: "test-token",
	}

	// Create a test actor
	actor := &Actor{
		connection:  conn,
		workspaceID: "test-workspace",
		traceID:     "test-trace",
		user: &proto.User{
			UserAgent: "test-agent",
			RemoteIp:  "127.0.0.1",
			UserId:    "test-user",
			Client:    "test-client",
		},
	}

	// Clone the actor without workspace
	cloned := actor.CloneWithoutWorkspace()

	// Verify the cloned actor
	if cloned.connection != conn {
		t.Errorf("CloneWithoutWorkspace().connection = %v; want %v", cloned.connection, conn)
	}

	if cloned.workspaceID != noWorkspace {
		t.Errorf("CloneWithoutWorkspace().workspaceID = %s; want %s", cloned.workspaceID, noWorkspace)
	}

	if cloned.traceID != "test-trace" {
		t.Errorf("CloneWithoutWorkspace().traceID = %s; want %s", cloned.traceID, "test-trace")
	}

	if cloned.user != actor.user {
		t.Errorf("CloneWithoutWorkspace().user = %v; want %v", cloned.user, actor.user)
	}
}

func Test_Actor_Connection(t *testing.T) {
	// Create a test connection
	conn := &Connection{
		appID: proto.VendorApp{
			VendorId: "test-vendor",
			AppId:    "test-app",
		},
		token: "test-token",
	}

	// Create a test actor
	actor := &Actor{
		connection:  conn,
		workspaceID: "test-workspace",
	}

	// Test Connection method
	if actor.Connection() != conn {
		t.Errorf("Actor.Connection() = %v; want %v", actor.Connection(), conn)
	}

	// Test ReplaceConnection method
	newConn := &Connection{
		appID: proto.VendorApp{
			VendorId: "new-vendor",
			AppId:    "new-app",
		},
		token: "new-token",
	}

	actor.ReplaceConnection(newConn)

	if actor.Connection() != newConn {
		t.Errorf("Actor.Connection() after ReplaceConnection = %v; want %v", actor.Connection(), newConn)
	}
}

func Test_Actor_UserMethods(t *testing.T) {
	// Create a test actor with user
	actor := &Actor{
		user: &proto.User{
			UserAgent: "test-agent",
			RemoteIp:  "127.0.0.1",
			UserId:    "test-user",
			Client:    "test-client",
		},
	}

	// Test user-related methods
	if actor.UserAgent() != "test-agent" {
		t.Errorf("Actor.UserAgent() = %s; want %s", actor.UserAgent(), "test-agent")
	}

	if actor.RemoteIP() != "127.0.0.1" {
		t.Errorf("Actor.RemoteIP() = %s; want %s", actor.RemoteIP(), "127.0.0.1")
	}

	if actor.UserID() != "test-user" {
		t.Errorf("Actor.UserID() = %s; want %s", actor.UserID(), "test-user")
	}

	if actor.Client() != "test-client" {
		t.Errorf("Actor.Client() = %s; want %s", actor.Client(), "test-client")
	}

	if actor.User() != actor.user {
		t.Errorf("Actor.User() = %v; want %v", actor.User(), actor.user)
	}

	// Test SetClientName
	actor.SetClientName("new-client")

	if actor.Client() != "new-client" {
		t.Errorf("Actor.Client() after SetClientName = %s; want %s", actor.Client(), "new-client")
	}

	// Test SetClientName with nil user
	actor = &Actor{}
	actor.SetClientName("test-client")

	if actor.Client() != "test-client" {
		t.Errorf("Actor.Client() after SetClientName with nil user = %s; want %s", actor.Client(), "test-client")
	}
}

func Test_Actor_AppMethods(t *testing.T) {
	// Create a test connection
	conn := &Connection{
		appID: proto.VendorApp{
			VendorId: "test-vendor",
			AppId:    "test-app",
		},
	}

	// Create a test actor
	actor := &Actor{
		connection: conn,
	}

	// Test app-related methods
	if actor.VendorID() != "test-vendor" {
		t.Errorf("Actor.VendorID() = %s; want %s", actor.VendorID(), "test-vendor")
	}

	if actor.AppID() != "test-app" {
		t.Errorf("Actor.AppID() = %s; want %s", actor.AppID(), "test-app")
	}

	vendorApp := actor.VendorApp()
	if vendorApp.GetVendorId() != "test-vendor" || vendorApp.GetAppId() != "test-app" {
		t.Errorf("Actor.VendorApp() = %v; want VendorId: %s, AppId: %s",
			vendorApp, "test-vendor", "test-app")
	}
}

func Test_Actor_WorkspaceAndTrace(t *testing.T) {
	// Create a test actor
	actor := &Actor{
		workspaceID: "test-workspace",
		traceID:     "test-trace",
	}

	// Test workspace and trace methods
	if actor.WorkspaceID() != "test-workspace" {
		t.Errorf("Actor.WorkspaceID() = %s; want %s", actor.WorkspaceID(), "test-workspace")
	}

	if actor.TraceID() != "test-trace" {
		t.Errorf("Actor.TraceID() = %s; want %s", actor.TraceID(), "test-trace")
	}

	// Test SetTraceID
	actor.SetTraceID("new-trace")

	if actor.TraceID() != "new-trace" {
		t.Errorf("Actor.TraceID() after SetTraceID = %s; want %s", actor.TraceID(), "new-trace")
	}
}

func Test_Actor_Authorization(t *testing.T) {
	// Create a test connection
	conn := &Connection{
		appID: proto.VendorApp{
			VendorId: "test-vendor",
			AppId:    "test-app",
		},
		token: "test-token",
	}

	// Create a test actor
	actor := &Actor{
		connection:  conn,
		workspaceID: "test-workspace",
		traceID:     "test-trace",
		user: &proto.User{
			UserAgent: "test-agent",
			RemoteIp:  "127.0.0.1",
			UserId:    "test-user",
			Client:    "test-client",
		},
	}

	// Test Authorization method
	auth := actor.Authorization()

	if auth.GetSource().GetVendorId() != "test-vendor" || auth.GetSource().GetAppId() != "test-app" {
		t.Errorf("Authorization().Source = %v; want VendorId: %s, AppId: %s",
			auth.GetSource(), "test-vendor", "test-app")
	}

	if auth.GetToken() != "test-token" {
		t.Errorf("Authorization().Token = %s; want %s", auth.GetToken(), "test-token")
	}

	if auth.GetTraceId() != "test-trace" {
		t.Errorf("Authorization().TraceId = %s; want %s", auth.GetTraceId(), "test-trace")
	}

	if auth.GetWorkspaceId() != "test-workspace" {
		t.Errorf("Authorization().WorkspaceId = %s; want %s", auth.GetWorkspaceId(), "test-workspace")
	}

	if auth.GetUser() != actor.user {
		t.Errorf("Authorization().User = %v; want %v", auth.GetUser(), actor.user)
	}

	// Test Authorization with nil connection
	actor.connection = nil
	if actor.Authorization() != nil {
		t.Errorf("Authorization() with nil connection = %v; want nil", actor.Authorization())
	}
}

func Test_Actor_AuthorizeContext(t *testing.T) {
	// Create a test connection
	conn := &Connection{
		appID: proto.VendorApp{
			VendorId: "test-vendor",
			AppId:    "test-app",
		},
		token: "test-token",
	}

	// Create a test actor
	actor := &Actor{
		connection:  conn,
		workspaceID: "test-workspace",
		traceID:     "test-trace",
		user: &proto.User{
			UserAgent: "test-agent",
			RemoteIp:  "127.0.0.1",
			UserId:    "test-user",
			Client:    "test-client",
		},
	}

	// Test AuthorizeContext method
	ctx := context.Background()
	authorizedCtx := actor.AuthorizeContext(ctx)

	// Extract metadata from context
	md, ok := metadata.FromOutgoingContext(authorizedCtx)
	if !ok {
		t.Errorf("Failed to extract metadata from context")
	}

	// Verify metadata values
	expectedValues := map[string]string{
		"workspace_id": "test-workspace",
		"trace_id":     "test-trace",
		"vendor_id":    "test-vendor",
		"app_id":       "test-app",
		"token":        "test-token",
		"client":       "test-client",
		"user_id":      "test-user",
		"user_agent":   "test-agent",
		"remote_ip":    "127.0.0.1",
	}

	for key, expectedValue := range expectedValues {
		values := md.Get(key)
		if len(values) != 1 || values[0] != expectedValue {
			t.Errorf("Context metadata[%s] = %v; want [%s]", key, values, expectedValue)
		}
	}
}
