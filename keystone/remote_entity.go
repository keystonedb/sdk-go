package keystone

import "context"

// RemoteEntity is a remote entity that is not stored in the local database
func RemoteEntity(entityID string) *Remote {
	return &Remote{_entityID: entityID}
}

type Remote struct {
	_entityID string
	EmbeddedSensors
	EmbeddedLogs
	EmbeddedEvents
	//EmbeddedRelationships //TODO: Review is this is possible
}

func (r Remote) GetKeystoneID() string   { return r._entityID }
func (r Remote) SetKeystoneID(id string) { r._entityID = id }

func (r Remote) Mutate(ctx context.Context, actor *Actor, options ...MutateOption) error {
	return actor.RemoteMutate(ctx, r.GetKeystoneID(), &r, options...)
}
