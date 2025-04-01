package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
)

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
	_exists   *bool
}

func (e *EmbeddedEntity) ObserveMutation(resp *proto.MutateResponse) {
	if e._entityID == "" && resp.EntityId != "" {
		e._entityID = ID(resp.EntityId)
		e._exists = Pointer(true)
	}
}

func (e *EmbeddedEntity) ObserveRetrieve(resp *proto.EntityResponse) {
	if e._entityID == "" {
		e.SetKeystoneID(ID(resp.GetEntity().GetEntityId()))
	}
	if resp.Exists != nil {
		e._exists = resp.Exists
	}
}

func (e *EmbeddedEntity) StoredInKeystone() *bool {
	return e._exists
}

func (e *EmbeddedEntity) ExistsInKeystone() bool {
	return e._exists != nil && *e._exists
}

func (e *EmbeddedEntity) GetKeystoneID() ID {
	return e._entityID
}
func (e *EmbeddedEntity) SetKeystoneID(id ID) { e._entityID = id }
