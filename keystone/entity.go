package keystone

import "github.com/keystonedb/sdk-go/proto"

type BaseEntity struct {
	EmbeddedEntity
	EmbeddedChildren
	EmbeddedWatcher
	EmbeddedDetails
	EmbeddedEvents
	EmbeddedLabels
	EmbeddedLock
	EmbeddedLogs
	EmbeddedRelationships
	EmbeddedSensors
	EmbeddedObjects
}

type EmbeddedEntity struct {
	_entityID ID
}

func (e *EmbeddedEntity) MutationSuccess(resp *proto.MutateResponse) {
	if e._entityID == "" && resp.EntityId != "" {
		e._entityID = ID(resp.EntityId)
	}
}

func (e *EmbeddedEntity) GetKeystoneID() ID {
	return e._entityID
}
func (e *EmbeddedEntity) SetKeystoneID(id ID) { e._entityID = id }
