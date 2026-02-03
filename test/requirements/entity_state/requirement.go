package entity_state

import (
	"context"
	"errors"
	"fmt"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type StateTestEntity struct {
	keystone.BaseEntity
	Name string

	retrievedState proto.EntityState
}

func (s *StateTestEntity) ObserveRetrieve(resp *proto.EntityResponse) {
	if resp.GetEntity() != nil {
		s.retrievedState = resp.GetEntity().GetState()
	}
}

type Requirement struct {
	entityID         keystone.ID
	activeEntityID   keystone.ID
	archivedEntityID keystone.ID
	offlineEntityID  keystone.ID
}

func (d *Requirement) Name() string {
	return "Entity State"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(StateTestEntity{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.createEntity(actor),
		d.verifyDefaultState(actor),
		d.setArchivedState(actor),
		d.setOfflineState(actor),
		d.setCorruptState(actor),
		d.restoreToActiveState(actor),
		d.archiveEntityMethod(actor),
		d.corruptEntityMethod(actor),
		d.restoreFromCorrupt(actor),
		// State filtering tests
		d.createEntitiesForFiltering(actor),
		d.findOnlyActive(actor),
		d.findIncludeArchived(actor),
		d.findOnlyArchived(actor),
		d.findWithStates(actor),
		d.findAllStates(actor),
	}
}

func (d *Requirement) createEntity(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create Entity"}
	entity := &StateTestEntity{Name: "State Test Entity"}
	err := actor.Mutate(context.Background(), entity)
	if err == nil {
		d.entityID = entity.GetKeystoneID()
	}
	return res.WithError(err)
}

func (d *Requirement) verifyDefaultState(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Verify Default State is Active"}
	entity := &StateTestEntity{}
	err := actor.Get(context.Background(), keystone.ByEntityID(entity, d.entityID), entity, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}
	if entity.retrievedState != proto.EntityState_Active {
		return res.WithError(fmt.Errorf("expected Active state, got %v", entity.retrievedState))
	}
	return res
}

func (d *Requirement) setArchivedState(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Set Archived State"}
	entity := &StateTestEntity{}
	entity.SetKeystoneID(d.entityID)

	err := actor.Mutate(context.Background(), entity, keystone.WithState(proto.EntityState_Archived))
	if err != nil {
		return res.WithError(err)
	}

	// Verify the state was set
	readEntity := &StateTestEntity{}
	err = actor.Get(context.Background(), keystone.ByEntityID(readEntity, d.entityID), readEntity, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}
	if readEntity.retrievedState != proto.EntityState_Archived {
		return res.WithError(fmt.Errorf("expected Archived state, got %v", readEntity.retrievedState))
	}
	return res
}

func (d *Requirement) setOfflineState(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Set Offline State"}
	entity := &StateTestEntity{}
	entity.SetKeystoneID(d.entityID)

	err := actor.Mutate(context.Background(), entity, keystone.WithState(proto.EntityState_Offline))
	if err != nil {
		return res.WithError(err)
	}

	// Verify the state was set
	readEntity := &StateTestEntity{}
	err = actor.Get(context.Background(), keystone.ByEntityID(readEntity, d.entityID), readEntity, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}
	if readEntity.retrievedState != proto.EntityState_Offline {
		return res.WithError(fmt.Errorf("expected Offline state, got %v", readEntity.retrievedState))
	}
	return res
}

func (d *Requirement) setCorruptState(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Set Corrupt State"}
	entity := &StateTestEntity{}
	entity.SetKeystoneID(d.entityID)

	err := actor.Mutate(context.Background(), entity, keystone.WithState(proto.EntityState_Corrupt))
	if err != nil {
		return res.WithError(err)
	}

	// Verify the state was set
	readEntity := &StateTestEntity{}
	err = actor.Get(context.Background(), keystone.ByEntityID(readEntity, d.entityID), readEntity, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}
	if readEntity.retrievedState != proto.EntityState_Corrupt {
		return res.WithError(fmt.Errorf("expected Corrupt state, got %v", readEntity.retrievedState))
	}
	return res
}

func (d *Requirement) restoreToActiveState(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Restore to Active State"}
	entity := &StateTestEntity{}
	entity.SetKeystoneID(d.entityID)

	err := actor.Mutate(context.Background(), entity, keystone.WithState(proto.EntityState_Active))
	if err != nil {
		return res.WithError(err)
	}

	// Verify the state was restored
	readEntity := &StateTestEntity{}
	err = actor.Get(context.Background(), keystone.ByEntityID(readEntity, d.entityID), readEntity, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}
	if readEntity.retrievedState != proto.EntityState_Active {
		return res.WithError(fmt.Errorf("expected Active state after restore, got %v", readEntity.retrievedState))
	}
	return res
}

func (d *Requirement) archiveEntityMethod(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "ArchiveEntity Convenience Method"}
	entity := &StateTestEntity{}
	entity.SetKeystoneID(d.entityID)

	err := actor.ArchiveEntity(context.Background(), entity)
	if err != nil {
		return res.WithError(err)
	}

	// Verify the state was set
	readEntity := &StateTestEntity{}
	err = actor.Get(context.Background(), keystone.ByEntityID(readEntity, d.entityID), readEntity, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}
	if readEntity.retrievedState != proto.EntityState_Archived {
		return res.WithError(fmt.Errorf("expected Archived state from ArchiveEntity, got %v", readEntity.retrievedState))
	}
	return res
}

func (d *Requirement) corruptEntityMethod(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "CorruptEntity Convenience Method"}
	entity := &StateTestEntity{}
	entity.SetKeystoneID(d.entityID)

	err := actor.CorruptEntity(context.Background(), entity)
	if err != nil {
		return res.WithError(err)
	}

	// Verify the state was set
	readEntity := &StateTestEntity{}
	err = actor.Get(context.Background(), keystone.ByEntityID(readEntity, d.entityID), readEntity, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}
	if readEntity.retrievedState != proto.EntityState_Corrupt {
		return res.WithError(fmt.Errorf("expected Corrupt state from CorruptEntity, got %v", readEntity.retrievedState))
	}
	return res
}

func (d *Requirement) restoreFromCorrupt(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Restore from Corrupt to Active"}
	entity := &StateTestEntity{}
	entity.SetKeystoneID(d.entityID)

	err := actor.Mutate(context.Background(), entity, keystone.WithState(proto.EntityState_Active))
	if err != nil {
		return res.WithError(err)
	}

	// Verify the state was restored
	readEntity := &StateTestEntity{}
	err = actor.Get(context.Background(), keystone.ByEntityID(readEntity, d.entityID), readEntity, keystone.WithSummary())
	if err != nil {
		return res.WithError(err)
	}
	if readEntity.retrievedState != proto.EntityState_Active {
		return res.WithError(fmt.Errorf("expected Active state after restore from Corrupt, got %v", readEntity.retrievedState))
	}
	return res
}

// State filtering integration tests

func (d *Requirement) createEntitiesForFiltering(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create Entities for State Filtering"}

	// Create Active entity
	activeEntity := &StateTestEntity{Name: "Active Entity for Filter"}
	if err := actor.Mutate(context.Background(), activeEntity); err != nil {
		return res.WithError(fmt.Errorf("failed to create active entity: %w", err))
	}
	d.activeEntityID = activeEntity.GetKeystoneID()

	// Create Archived entity
	archivedEntity := &StateTestEntity{Name: "Archived Entity for Filter"}
	if err := actor.Mutate(context.Background(), archivedEntity); err != nil {
		return res.WithError(fmt.Errorf("failed to create archived entity: %w", err))
	}
	if err := actor.ArchiveEntity(context.Background(), archivedEntity); err != nil {
		return res.WithError(fmt.Errorf("failed to archive entity: %w", err))
	}
	d.archivedEntityID = archivedEntity.GetKeystoneID()

	// Create Offline entity
	offlineEntity := &StateTestEntity{Name: "Offline Entity for Filter"}
	if err := actor.Mutate(context.Background(), offlineEntity); err != nil {
		return res.WithError(fmt.Errorf("failed to create offline entity: %w", err))
	}
	if err := actor.Mutate(context.Background(), offlineEntity, keystone.WithState(proto.EntityState_Offline)); err != nil {
		return res.WithError(fmt.Errorf("failed to set offline state: %w", err))
	}
	d.offlineEntityID = offlineEntity.GetKeystoneID()

	return res
}

func (d *Requirement) findOnlyActive(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Find OnlyActive"}

	entities, err := actor.Find(
		context.Background(),
		keystone.Type(StateTestEntity{}),
		keystone.WithSummary(),
		keystone.OnlyActive(),
		keystone.WhereIn("_entity_id", d.activeEntityID.String(), d.archivedEntityID.String(), d.offlineEntityID.String()),
	)
	if err != nil {
		return res.WithError(err)
	}

	// Should only find the active entity
	if len(entities) != 1 {
		return res.WithError(fmt.Errorf("expected 1 entity with OnlyActive, got %d", len(entities)))
	}
	if entities[0].GetEntity().GetEntityId() != d.activeEntityID.String() {
		return res.WithError(errors.New("expected to find the active entity"))
	}
	return res
}

func (d *Requirement) findIncludeArchived(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Find IncludeArchived"}

	entities, err := actor.Find(
		context.Background(),
		keystone.Type(StateTestEntity{}),
		keystone.WithSummary(),
		keystone.IncludeArchived(),
		keystone.WhereIn("_entity_id", d.activeEntityID.String(), d.archivedEntityID.String(), d.offlineEntityID.String()),
	)
	if err != nil {
		return res.WithError(err)
	}

	// Should find active and archived entities (2 total)
	if len(entities) != 2 {
		return res.WithError(fmt.Errorf("expected 2 entities with IncludeArchived, got %d", len(entities)))
	}

	foundActive := false
	foundArchived := false
	for _, e := range entities {
		if e.GetEntity().GetEntityId() == d.activeEntityID.String() {
			foundActive = true
		}
		if e.GetEntity().GetEntityId() == d.archivedEntityID.String() {
			foundArchived = true
		}
	}
	if !foundActive || !foundArchived {
		return res.WithError(errors.New("expected to find both active and archived entities"))
	}
	return res
}

func (d *Requirement) findOnlyArchived(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Find OnlyArchived"}

	entities, err := actor.Find(
		context.Background(),
		keystone.Type(StateTestEntity{}),
		keystone.WithSummary(),
		keystone.OnlyArchived(),
		keystone.WhereIn("_entity_id", d.activeEntityID.String(), d.archivedEntityID.String(), d.offlineEntityID.String()),
	)
	if err != nil {
		return res.WithError(err)
	}

	// Should only find the archived entity
	if len(entities) != 1 {
		return res.WithError(fmt.Errorf("expected 1 entity with OnlyArchived, got %d", len(entities)))
	}
	if entities[0].GetEntity().GetEntityId() != d.archivedEntityID.String() {
		return res.WithError(errors.New("expected to find the archived entity"))
	}
	return res
}

func (d *Requirement) findWithStates(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Find WithStates (Active, Offline)"}

	entities, err := actor.Find(
		context.Background(),
		keystone.Type(StateTestEntity{}),
		keystone.WithSummary(),
		keystone.WithStates(proto.EntityState_Active, proto.EntityState_Offline),
		keystone.WhereIn("_entity_id", d.activeEntityID.String(), d.archivedEntityID.String(), d.offlineEntityID.String()),
	)
	if err != nil {
		return res.WithError(err)
	}

	// Should find active and offline entities (2 total)
	if len(entities) != 2 {
		return res.WithError(fmt.Errorf("expected 2 entities with WithStates(Active, Offline), got %d", len(entities)))
	}

	foundActive := false
	foundOffline := false
	for _, e := range entities {
		if e.GetEntity().GetEntityId() == d.activeEntityID.String() {
			foundActive = true
		}
		if e.GetEntity().GetEntityId() == d.offlineEntityID.String() {
			foundOffline = true
		}
	}
	if !foundActive || !foundOffline {
		return res.WithError(errors.New("expected to find both active and offline entities"))
	}
	return res
}

func (d *Requirement) findAllStates(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Find AllStates"}

	entities, err := actor.Find(
		context.Background(),
		keystone.Type(StateTestEntity{}),
		keystone.WithSummary(),
		keystone.AllStates(),
		keystone.WhereIn("_entity_id", d.activeEntityID.String(), d.archivedEntityID.String(), d.offlineEntityID.String()),
	)
	if err != nil {
		return res.WithError(err)
	}

	// Should find all 3 entities
	if len(entities) != 3 {
		return res.WithError(fmt.Errorf("expected 3 entities with AllStates, got %d", len(entities)))
	}

	foundActive := false
	foundArchived := false
	foundOffline := false
	for _, e := range entities {
		if e.GetEntity().GetEntityId() == d.activeEntityID.String() {
			foundActive = true
		}
		if e.GetEntity().GetEntityId() == d.archivedEntityID.String() {
			foundArchived = true
		}
		if e.GetEntity().GetEntityId() == d.offlineEntityID.String() {
			foundOffline = true
		}
	}
	if !foundActive || !foundArchived || !foundOffline {
		return res.WithError(errors.New("expected to find all three entities"))
	}
	return res
}
