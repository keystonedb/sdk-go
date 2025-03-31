package keystone

import "github.com/keystonedb/sdk-go/proto"

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
func (a *Actor) RemoteIp() string  { return a.user.GetRemoteIp() }
func (a *Actor) UserId() string    { return a.user.GetUserId() }
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

// SetClientName sets the client name for the actor
func (a *Actor) SetClientName(client string) {
	if a.user == nil {
		a.user = &proto.User{}
	}
	a.user.Client = client
}
