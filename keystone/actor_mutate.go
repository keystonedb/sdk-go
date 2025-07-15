package keystone

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/packaged/logger/v3/logger"
	"reflect"
)

var ErrCommentedMutations = errors.New("you must provide a mutation comment")

func (a *Actor) RemoteMutate(ctx context.Context, entityID ID, src interface{}, options ...MutateOption) error {
	mutation := &proto.Mutation{}

	if entityID == "" {
		if rawEntity, ok := src.(Entity); ok {
			entityID = rawEntity.GetKeystoneID()
		}
	}

	if entityID == "" {
		return errors.New("entityID is required for remote mutations")
	}

	if _, isDyn := src.(ConvertStructToDynamicProperties); isDyn {
		props, err := DynamicPropertiesFromStructWithoutDefaults(src)
		if err == nil {
			mutation.DynamicProperties = props
		} else {
			return err
		}
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
		EntityId:      entityID.String(),
		Mutation:      mutation,
	}

	for _, option := range options {
		option.apply(m)
	}

	mResp, err := a.connection.Mutate(ctx, m)

	if err == nil {
		for _, option := range options {
			if optObserver, ok := option.(MutationObserver); ok {
				optObserver.ObserveMutation(mResp)
			}
		}

		if rawEntity, ok := src.(MutationObserver); ok {
			rawEntity.ObserveMutation(mResp)
			observeMutation(rawEntity, mResp)
		}
	}

	return mutateToError(mResp, err)
}

func (a *Actor) MutateWithDefaultWatcher(ctx context.Context, src interface{}, options ...MutateOption) error {
	w, err := NewDefaultsWatcher(src)
	if err != nil {
		return err
	}
	return a.MutateWithWatcher(ctx, src, w, options...)
}

func (a *Actor) MutateWithWatcher(ctx context.Context, src interface{}, w *Watcher, options ...MutateOption) error {
	changes, err := w.Changes(src, true)
	if err != nil {
		return err
	}
	return a.mutateWithProperties(ctx, src, changes, options...)
}

func (a *Actor) mutateWithProperties(ctx context.Context, src interface{}, props map[Property]*proto.Value, options ...MutateOption) error {
	if reflect.TypeOf(src).Kind() != reflect.Pointer {
		return errors.New("mutate requires a pointer to a struct")
	}

	var onSuccess []func() error
	var onSuccessMutate []func(response *proto.MutateResponse)

	schema, registered := a.connection.registerType(src)
	if !registered {
		// wait for the type to be registered with the keystone server
		a.connection.SyncSchema().Wait()
	}

	mutation := &proto.Mutation{}
	//properties
	//children
	mutation.Mutator = a.user
	entityID := ID("")

	if rawEntity, ok := src.(Entity); ok {
		entityID = rawEntity.GetKeystoneID()
	}

	if entityWithLabels, ok := src.(LabelProvider); ok {
		mutation.Labels = entityWithLabels.GetLabels()
		onSuccess = append(onSuccess, entityWithLabels.ClearLabels)
	}

	if entityWithSensor, ok := src.(SensorProvider); ok {
		mutation.Measurements = entityWithSensor.GetSensorMeasurements()
		onSuccess = append(onSuccess, entityWithSensor.ClearSensorMeasurements)
	}

	if entityWithRelationships, ok := src.(RelationshipProvider); ok {
		mutation.Relationships = entityWithRelationships.GetRelationships()
		onSuccess = append(onSuccess, entityWithRelationships.ClearRelationships)
	}

	if entityWithEvents, ok := src.(EventProvider); ok {
		mutation.Events = entityWithEvents.GetEvents()
		onSuccess = append(onSuccess, entityWithEvents.ClearEvents)
	}

	if entityWithLogs, ok := src.(LogProvider); ok {
		mutation.Logs = entityWithLogs.GetLogs()
		onSuccess = append(onSuccess, entityWithLogs.ClearLogs)
	}

	if entityWithChildren, ok := src.(ChildProvider); ok {
		mutation.Children = entityWithChildren.GetChildrenToStore()
		mutation.RemoveChildren = entityWithChildren.GetChildrenToRemove()
		mutation.RemoveAllChildrenByType = entityWithChildren.GetTruncateChildrenType()

		if entityWithEChildren, okU := src.(ChildUpdateProvider); okU {
			onSuccessMutate = append(onSuccessMutate, entityWithEChildren.updateChildren)
		}
		onSuccess = append(onSuccess, entityWithChildren.ClearChildChanges)
	}

	if len(props) > 0 {
		for propName, prop := range props {
			mutation.Properties = append(mutation.Properties, &proto.EntityProperty{Property: propName.Name(), Value: prop})
		}
	}

	m := &proto.MutateRequest{
		Authorization: a.Authorization(),
		EntityId:      entityID.String(),
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

	if err == nil {

		if mResp.GetSuccess() {
			for _, onMutateSuccessFunc := range onSuccessMutate {
				onMutateSuccessFunc(mResp)
			}
		}

		for _, option := range options {
			if optObserver, ok := option.(MutationObserver); ok {
				optObserver.ObserveMutation(mResp)
			}
		}

		if rawEntity, ok := src.(Entity); ok && mResp.GetEntityId() != "" {
			rawEntity.SetKeystoneID(ID(mResp.GetEntityId()))
		}

		if rawEntity, ok := src.(MutationObserver); ok {
			rawEntity.ObserveMutation(mResp)
			observeMutation(rawEntity, mResp)
		}

		if mResp.GetSuccess() {
			for _, onSuccessFunc := range onSuccess {
				logger.I().ErrorIf(onSuccessFunc(), "failed to run onSuccess function")
			}
		}

	}

	return mutateToError(mResp, err)
}

// Mutate is a function that can mutate an entity
func (a *Actor) Mutate(ctx context.Context, src interface{}, options ...MutateOption) error {
	srcType := reflect.TypeOf(src)
	if srcType.Kind() != reflect.Pointer || srcType.Elem().Kind() == reflect.Pointer {
		return errors.New("mutate requires a pointer to a struct")
	}

	if watchable, ok := src.(WatchedEntity); ok && watchable.HasWatcher() {
		return a.MutateWithWatcher(ctx, src, watchable.Watcher(), options...)
	} else if entity, settable := src.(SettableWatchedEntity); settable && !watchable.HasWatcher() {
		if w, err := NewDefaultsWatcher(src); err == nil {
			entity.SetWatcher(w)
			for _, option := range options {
				if prepare, canPrepare := option.(MutationOptionWatcherPrepare); canPrepare {
					_ = prepare.prepare(w)
				}
			}
			return a.MutateWithWatcher(ctx, src, w, options...)
		}
	}

	props, err := Marshal(src)
	if err != nil {
		return err
	}
	return a.mutateWithProperties(ctx, src, props, options...)

}

func mutateToError(resp *proto.MutateResponse, err error) error {
	if err != nil {
		return err
	}

	if resp == nil {
		return errors.New("nil response")
	}

	if resp.ErrorCode > 0 || resp.ErrorMessage != "" {
		return &Error{
			ErrorMessage: resp.GetErrorMessage(), ErrorCode: resp.GetErrorCode(),
			Extended: resp.GetExtended().GetErrors(), Suggestions: resp.GetExtended().GetSuggestions(),
		}
	}
	return nil
}
