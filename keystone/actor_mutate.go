package keystone

import (
	"context"
	"errors"
	"fmt"
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

var ErrCommentedMutations = errors.New("you must provide a mutation comment")

func (a *Actor) RemoteMutate(ctx context.Context, src interface{}) error {
	mutation := &proto.Mutation{}
	entityID := ""
	if rawEntity, ok := src.(Entity); ok {
		entityID = rawEntity.GetKeystoneID()
	}

	if entityID == "" {
		return errors.New("entityID is required for remote mutations")
	}

	if entityWithSensor, ok := src.(SensorProvider); ok {
		mutation.Measurements = entityWithSensor.GetSensorMeasurements()
	}
	if entityWithEvents, ok := src.(EventProvider); ok {
		mutation.Events = entityWithEvents.GetEvents()
	}
	if entityWithLogs, ok := src.(LogProvider); ok {
		mutation.Logs = entityWithLogs.GetLogs()
	}

	m := &proto.MutateRequest{
		Authorization: a.Authorization(),
		EntityId:      entityID,
		Mutation:      mutation,
	}

	return mutateToError(a.connection.Mutate(ctx, m))
}

// Mutate is a function that can mutate an entity
func (a *Actor) Mutate(ctx context.Context, src interface{}, comment string, options ...MutateOption) error {
	if reflect.TypeOf(src).Kind() != reflect.Pointer {
		return errors.New("mutate requires a pointer to a struct")
	}

	schema, registered := a.connection.registerType(src)
	if !registered {
		// wait for the type to be registered with the keystone server
		a.connection.SyncSchema().Wait()
	}

	mutation := &proto.Mutation{}
	//properties
	//children
	mutation.Mutator = a.user
	entityID := ""
	mutation.Comment = comment

	if rawEntity, ok := src.(Entity); ok {
		entityID = rawEntity.GetKeystoneID()
	}

	if entityWithLabels, ok := src.(LabelProvider); ok {
		mutation.Labels = entityWithLabels.GetLabels()
	}

	if entityWithSensor, ok := src.(SensorProvider); ok {
		mutation.Measurements = entityWithSensor.GetSensorMeasurements()
	}

	if entityWithRelationships, ok := src.(RelationshipProvider); ok {
		mutation.Relationships = entityWithRelationships.GetRelationships()
	}

	if entityWithEvents, ok := src.(EventProvider); ok {
		mutation.Events = entityWithEvents.GetEvents()
	}

	if entityWithLogs, ok := src.(LogProvider); ok {
		mutation.Logs = entityWithLogs.GetLogs()
	}

	m := &proto.MutateRequest{
		Authorization: a.Authorization(),
		EntityId:      entityID,
		Schema:        &proto.Key{Key: schema.Type, Source: a.VendorApp()},
		Mutation:      mutation,
	}

	for _, option := range options {
		option.apply(m)
	}

	if schema.HasOption(proto.Schema_StoreMutations) && mutation.GetComment() == "" {
		return ErrCommentedMutations
	}

	mResp, err := a.connection.Mutate(ctx, m)

	if err == nil && mResp.Success {
		if rawEntity, ok := src.(Entity); ok && entityID == "" {
			rawEntity.SetKeystoneID(mResp.GetEntityId())
		}
	}

	return mutateToError(mResp, err)
}

func mutateToError(resp *proto.MutateResponse, err error) error {
	if err != nil {
		return err
	}

	if resp == nil {
		return errors.New("nil response")
	}

	if resp.ErrorCode > 0 || resp.ErrorMessage != "" {
		return fmt.Errorf("error %d: %s", resp.ErrorCode, resp.ErrorMessage)
	}
	return nil
}