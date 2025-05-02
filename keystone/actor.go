package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc/metadata"
)

const noWorkspace = "__"

// Actor is a struct that represents an actor making requests to keystone
type Actor struct {
	connection  *Connection
	workspaceID string
	traceID     string
	user        *proto.User
}

func (a *Actor) CloneWithoutWorkspace() *Actor {
	if a == nil {
		return nil
	}

	return &Actor{
		connection:  a.connection,
		workspaceID: noWorkspace,
		traceID:     a.traceID,
		user:        a.user,
	}
}

func (a *Actor) ReplaceConnection(c *Connection) { a.connection = c }
func (a *Actor) Connection() *Connection         { return a.connection }

func (a *Actor) UserAgent() string { return a.user.GetUserAgent() }
func (a *Actor) RemoteIP() string  { return a.user.GetRemoteIp() }
func (a *Actor) UserID() string    { return a.user.GetUserId() }
func (a *Actor) Client() string    { return a.user.GetClient() }
func (a *Actor) User() *proto.User { return a.user }

func (a *Actor) VendorID() string {
	return a.connection.appID.GetVendorId()
}
func (a *Actor) AppID() string {
	return a.connection.appID.GetAppId()
}
func (a *Actor) VendorApp() *proto.VendorApp {
	return &a.connection.appID
}
func (a *Actor) WorkspaceID() string {
	return a.workspaceID
}

func (a *Actor) TraceID() string {
	return a.traceID
}
func (a *Actor) SetTraceID(id string) { a.traceID = id }

func (a *Actor) Authorization() *proto.Authorization {
	if a == nil || a.connection == nil {
		return nil
	}
	return &proto.Authorization{
		Source:      &a.connection.appID,
		Token:       a.connection.token,
		TraceId:     a.traceID,
		WorkspaceId: a.workspaceID,
		User:        a.User(),
	}
}

func (a *Actor) AuthorizeContext(ctx context.Context) context.Context {
	meta := metadata.New(map[string]string{
		"workspace_id": a.WorkspaceID(),
		"trace_id":     a.TraceID(),
		"vendor_id":    a.VendorID(),
		"app_id":       a.AppID(),
		"token":        a.Connection().token,
		"client":       a.Client(),
		"user_id":      a.UserID(),
		"user_agent":   a.UserAgent(),
		"remote_ip":    a.RemoteIP(),
	})
	return metadata.NewOutgoingContext(ctx, meta)
}

// SetClientName sets the client name for the actor
func (a *Actor) SetClientName(client string) {
	if a.user == nil {
		a.user = &proto.User{}
	}
	a.user.Client = client
}
