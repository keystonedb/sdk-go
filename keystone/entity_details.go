package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"time"
)

type EntityDetail interface {
	SetEntityDetail(entity *proto.Entity)
}

type EmbeddedDetails struct {
	ksCreated     time.Time
	ksStateChange time.Time
	ksState       proto.EntityState
	ksLastUpdate  time.Time
}

func (e *EmbeddedDetails) DateCreated() time.Time           { return e.ksCreated }
func (e *EmbeddedDetails) LastUpdated() time.Time           { return e.ksLastUpdate }
func (e *EmbeddedDetails) KeystoneState() proto.EntityState { return e.ksState }

func (e *EmbeddedDetails) SetEntityDetail(entity *proto.Entity) {
	if entity == nil {
		return
	}

	e.ksCreated = entity.GetCreated().AsTime()
	e.ksStateChange = entity.GetStateChange().AsTime()
	e.ksState = entity.GetState()
	e.ksLastUpdate = entity.GetLastUpdate().AsTime()
}
