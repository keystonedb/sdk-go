package entity_state

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/google/uuid"
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

func (s *StateTestEntity) observedState() proto.EntityState { return s.retrievedState }

// IndexedStateTestEntity has indexed properties so its state is also written to the SQL index.
type IndexedStateTestEntity struct {
	keystone.BaseEntity
	Name   string `keystone:",indexed"`
	TestID string `keystone:",indexed"`

	retrievedState proto.EntityState
}

func (s *IndexedStateTestEntity) ObserveRetrieve(resp *proto.EntityResponse) {
	if resp.GetEntity() != nil {
		s.retrievedState = resp.GetEntity().GetState()
	}
}

func (s *IndexedStateTestEntity) observedState() proto.EntityState { return s.retrievedState }

type stateEntity interface {
	keystone.Entity
	observedState() proto.EntityState
}

type Requirement struct {
	requirements.QueryCoverage
	entityID         keystone.ID
	activeEntityID   keystone.ID
	archivedEntityID keystone.ID
	offlineEntityID  keystone.ID
	testID           string
}

func (d *Requirement) Name() string {
	return "Entity State"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(StateTestEntity{}, IndexedStateTestEntity{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	report(d.createEntity(actor))
	report(d.verifyDefaultState(actor))
	report(d.setArchivedState(actor))
	report(d.setOfflineState(actor))
	report(d.setCorruptState(actor))
	report(d.restoreToActiveState(actor))
	report(d.archiveEntityMethod(actor))
	report(d.corruptEntityMethod(actor))
	report(d.restoreFromCorrupt(actor))
	// State filtering tests
	report(d.createEntitiesForFiltering(actor))
	report(d.findOnlyActive(actor))
	report(d.findIncludeArchived(actor))
	report(d.findOnlyArchived(actor))
	report(d.findWithStates(actor))
	report(d.findAllStates(actor))

	d.testID = uuid.NewString()
	d.verifyStateConsistency(actor, report, false)
	d.verifyStateConsistency(actor, report, true)
	report(d.groupCountByState(actor))
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

	entities, err := d.Find(actor,
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

	entities, err := d.Find(actor,
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

	entities, err := d.Find(actor,
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

	entities, err := d.Find(actor,
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

	entities, err := d.Find(actor,
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

// CQL and index consistency tests

var allStates = []proto.EntityState{proto.EntityState_Active, proto.EntityState_Offline, proto.EntityState_Corrupt, proto.EntityState_Archived}

// newStateEntity sets TestID only when creating, so mutations by ID send nothing but Name.
func (d *Requirement) newStateEntity(indexed bool, id keystone.ID, name string) stateEntity {
	var e stateEntity = &StateTestEntity{Name: name}
	if indexed {
		e = &IndexedStateTestEntity{Name: name}
		if id == "" {
			e.(*IndexedStateTestEntity).TestID = d.testID
		}
	}
	e.SetKeystoneID(id)
	return e
}

func (d *Requirement) verifyStateConsistency(actor *keystone.Actor, report requirements.Reporter, indexed bool) {
	ctx := context.Background()
	label := "CQL"
	if indexed {
		label = "SQL index"
	}
	step := func(name string, fn func() error) bool {
		err := fn()
		report(requirements.TestResult{Name: fmt.Sprintf("State Consistency (%s): %s", label, name)}.WithError(err))
		return err == nil
	}
	var id keystone.ID
	propertyOnly := func() error {
		if err := actor.Mutate(ctx, d.newStateEntity(indexed, id, "Renamed While Archived")); err != nil {
			return err
		}
		return d.expectState(actor, indexed, id, proto.EntityState_Archived, "Renamed While Archived")
	}

	_ = step("Create is Active", func() error {
		e := d.newStateEntity(indexed, "", "Consistency Entity")
		if err := actor.Mutate(ctx, e); err != nil {
			return err
		}
		id = e.GetKeystoneID()
		return d.expectState(actor, indexed, id, proto.EntityState_Active, "Consistency Entity")
	}) && step("ArchiveEntity is Archived", func() error {
		if err := actor.ArchiveEntity(ctx, d.newStateEntity(indexed, id, "")); err != nil {
			return err
		}
		return d.expectState(actor, indexed, id, proto.EntityState_Archived, "Consistency Entity")
	}) && step("Property-only mutation keeps Archived", propertyOnly) &&
		step("Repeated property-only mutation keeps Archived", propertyOnly) &&
		step("WithState Active restores Active", func() error {
			if err := actor.Mutate(ctx, d.newStateEntity(indexed, id, ""), keystone.WithState(proto.EntityState_Active)); err != nil {
				return err
			}
			return d.expectState(actor, indexed, id, proto.EntityState_Active, "Renamed While Archived")
		})
}

// expectState asserts Get and state-filtered lookups agree on want; wantName is only checked against the index.
func (d *Requirement) expectState(actor *keystone.Actor, indexed bool, id keystone.ID, want proto.EntityState, wantName string) error {
	read := d.newStateEntity(indexed, "", "")
	if err := actor.Get(context.Background(), keystone.ByEntityID(read, id), read, keystone.WithSummary()); err != nil {
		return err
	}
	if read.observedState() != want {
		return fmt.Errorf("Get: expected %v, got %v", want, read.observedState())
	}

	found, err := d.lookupState(actor, indexed, id, keystone.WithStates(want))
	if err != nil {
		return err
	}
	if found == nil {
		return fmt.Errorf("lookup filtered to %v did not return the entity", want)
	}
	if indexed {
		got := &IndexedStateTestEntity{}
		if err := keystone.Unmarshal(found, got); err != nil {
			return err
		}
		if got.Name != wantName {
			return fmt.Errorf("lookup filtered to %v: expected name %q, got %q", want, wantName, got.Name)
		}
	}

	others := slices.DeleteFunc(slices.Clone(allStates), func(s proto.EntityState) bool { return s == want })
	found, err = d.lookupState(actor, indexed, id, keystone.WithStates(others...))
	if err != nil {
		return err
	}
	if found != nil {
		return fmt.Errorf("lookup filtered to %v returned the entity, expected only %v", others, want)
	}
	return nil
}

func (d *Requirement) lookupState(actor *keystone.Actor, indexed bool, id keystone.ID, states keystone.FindOption) (*proto.EntityResponse, error) {
	var entities []*proto.EntityResponse
	var err error
	if indexed {
		entities, err = d.QueryIndex(actor, context.Background(), keystone.Type(IndexedStateTestEntity{}), []string{"name"}, states, keystone.WhereEquals("test_id", d.testID))
	} else {
		entities, err = d.Find(actor, context.Background(), keystone.Type(StateTestEntity{}), keystone.WithSummary(), states, keystone.WhereIn("_entity_id", id.String()))
	}
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		if e.GetEntity().GetEntityId() == id.String() {
			return e, nil
		}
	}
	return nil, nil
}

func (d *Requirement) groupCountByState(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "GroupCount by _state matches entity states"}
	ctx := context.Background()
	groupID := d.testID + "-group"
	for _, state := range []proto.EntityState{proto.EntityState_Active, proto.EntityState_Active, proto.EntityState_Archived, proto.EntityState_Offline} {
		e := &IndexedStateTestEntity{Name: "Group " + state.String(), TestID: groupID}
		if err := actor.Mutate(ctx, e); err != nil {
			return res.WithError(err)
		}
		if state == proto.EntityState_Active {
			continue
		}
		if err := actor.Mutate(ctx, e, keystone.WithState(state)); err != nil {
			return res.WithError(err)
		}
		if state == proto.EntityState_Archived {
			if err := actor.Mutate(ctx, d.newStateEntity(true, e.GetKeystoneID(), "Group Archived Renamed")); err != nil {
				return res.WithError(err)
			}
		}
	}

	counts := map[string]int32{}
	if d.UseQuery {
		// GroupBy always plans on SQL; ungrouped Query reads recent entities' state from CQL instead.
		resp, err := actor.Query(ctx, keystone.Type(IndexedStateTestEntity{}), keystone.WhereEquals("test_id", groupID), keystone.GroupBy("_state"), keystone.Count("count"))
		if err != nil {
			return res.WithError(err)
		}
		for _, row := range resp.GetRows() {
			counts[strconv.FormatInt(row.GetValues()["_state"].GetInt(), 10)] = int32(row.GetValues()["count"].GetInt())
		}
	} else {
		grouped, err := actor.GroupCount(ctx, keystone.Type(IndexedStateTestEntity{}), []string{"_state"}, keystone.WhereEquals("test_id", groupID))
		if err != nil {
			return res.WithError(err)
		}
		for key, row := range grouped {
			counts[key] = row.GetCount()
		}
	}
	want := map[proto.EntityState]int32{proto.EntityState_Active: 2, proto.EntityState_Archived: 1, proto.EntityState_Offline: 1}
	if len(counts) != len(want) {
		return res.WithError(fmt.Errorf("expected %d state groups, got %d: %v", len(want), len(counts), counts))
	}
	for state, count := range want {
		if got := counts[strconv.Itoa(int(state))]; got != count {
			return res.WithError(fmt.Errorf("expected %d %v, got %d", count, state, got))
		}
	}
	return res
}
