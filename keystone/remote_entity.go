package keystone

import "context"

// RemoteEntity is a remote entity that is not stored in the local database
func RemoteEntity(entityID ID) *Remote {
	return &Remote{_entityID: entityID}
}

type Remote struct {
	_entityID ID
	EmbeddedSensors
	EmbeddedLogs
	EmbeddedEvents
	//EmbeddedRelationships //TODO: Review is this is possible
}

func (r Remote) GetKeystoneID() ID   { return r._entityID }
func (r Remote) SetKeystoneID(id ID) { r._entityID = id }

func (r Remote) Mutate(ctx context.Context, actor *Actor, options ...MutateOption) error {
	return actor.RemoteMutate(ctx, r.GetKeystoneID(), &r, options...)
}

type DynamicRemoteEntity struct {
	Remote
}

func (d *DynamicRemoteEntity) convertToDynamicProperties() bool {
	return true
}

type ConvertStructToDynamicProperties interface {
	convertToDynamicProperties() bool
}
