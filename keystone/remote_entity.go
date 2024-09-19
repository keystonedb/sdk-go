package keystone

// RemoteEntity is a remote entity that is not stored in the local database
func RemoteEntity(entityID string) Remote {
	return Remote{_entityID: entityID}
}

type Remote struct {
	_entityID string
	EmbeddedSensors
	EmbeddedLogs
	EmbeddedRelationships
	EmbeddedLabels
	EmbeddedEvents
}

func (r Remote) GetKeystoneID() string   { return r._entityID }
func (r Remote) SetKeystoneID(id string) { r._entityID = id }
