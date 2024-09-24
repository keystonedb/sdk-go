package keystone

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
}

type EmbeddedEntity struct {
	_entityID string
}

func (e *EmbeddedEntity) GetKeystoneID() string {
	return e._entityID
}
func (e *EmbeddedEntity) SetKeystoneID(id string) { e._entityID = id }
