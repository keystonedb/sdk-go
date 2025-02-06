package keystone

import "github.com/keystonedb/sdk-go/proto"

type BaseEntity struct {
	EmbeddedEntity
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
	_entityID string
}

func (e *EmbeddedEntity) MutationSuccess(resp *proto.MutateResponse) {
	if e._entityID == "" && resp.EntityId != "" {
		e._entityID = resp.EntityId
	}
}

func (e *EmbeddedEntity) GetKeystoneID() string {
	return e._entityID
}
func (e *EmbeddedEntity) SetKeystoneID(id string) { e._entityID = id }
