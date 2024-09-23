package keystone

type BaseEntity struct {
	EmbeddedDetails
	EmbeddedEvents
	EmbeddedLabels
	EmbeddedLock
	EmbeddedLogs
	EmbeddedRelationships
	EmbeddedSensors
	_entityID string
}

func (e *BaseEntity) GetKeystoneID() string {
	return e._entityID
}
func (e *BaseEntity) SetKeystoneID(id string) { e._entityID = id }
