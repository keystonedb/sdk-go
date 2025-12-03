package keystone

import (
	"context"
	"errors"
	"strings"

	"github.com/keystonedb/sdk-go/proto"
)

func (a *Actor) GetByID(ctx context.Context, entityID ID, dst interface{}, retrieve ...RetrieveOption) error {
	return a.Get(ctx, ByEntityID(Type(dst), entityID), dst, retrieve...)
}

func (a *Actor) GetByUniqueProperty(ctx context.Context, uniqueId, propertyName string, dst interface{}, retrieve ...RetrieveOption) error {
	return a.Get(ctx, ByUniqueProperty(Type(dst), uniqueId, propertyName), dst, retrieve...)
}

// Get retrieves an entity by the given retrieveBy, storing the result in dst
func (a *Actor) Get(ctx context.Context, retrieveBy RetrieveBy, dst interface{}, retrieve ...RetrieveOption) error {
	entityRequest := retrieveBy.BaseRequest()
	entityRequest.Authorization = a.Authorization()
	for _, rOpt := range retrieve {
		rOpt.Apply(entityRequest.View)
		if reOpt, ok := rOpt.(RetrieveEntityOption); ok {
			reOpt.ApplyRequest(entityRequest)
		}
	}

	_, loadByUnique := retrieveBy.(byUniqueProperty)
	_, genericResult := dst.(GenericResult)
	if loadByUnique && genericResult {
		return errors.New("invalid retrieveBy and dst combination")
	}

	view := entityRequest.View

	// set source
	for _, p := range view.Properties {
		p.Source = a.Authorization().GetSource()
	}

	for _, r := range view.RelationshipByType {
		r.Source = a.Authorization().GetSource()
	}

	schema, registered := a.connection.registerType(dst)
	if !registered {
		// wait for the type to be registered with the keystone server
		a.connection.SyncSchema().Wait()
	}

	entityRequest.Schema = &proto.Key{Key: schema.Type, Source: a.Authorization().Source}

	if _, ok := retrieveBy.(byUniqueProperty); ok {
		schemaID := schema.id
		if schemaID == "" {
			schemaID = schema.Type
		}
		entityRequest.UniqueId.SchemaId = schemaID
	}

	resp, err := a.connection.Retrieve(ctx, entityRequest)
	if err != nil {
		return err
	}

	if lk, ok := dst.(Locker); ok && resp.GetLock() != nil {
		LockData := &LockInfo{
			LockAcquired: resp.GetLock().GetLockAcquired(),
			ID:           resp.GetLock().GetLockId(),
			LockedUntil:  resp.GetLock().GetLockedUntil().AsTime(),
			Message:      resp.GetLock().GetMessage(),
		}
		lk.SetLockResult(LockData)
	}

	var watcher *Watcher
	if watchable, ok := dst.(WatchedEntity); ok && watchable.HasWatcher() {
		watcher = watchable.Watcher()
	} else if entity, settable := dst.(SettableWatchedEntity); settable && !watchable.HasWatcher() {
		if w, err := NewDefaultsWatcher(dst); err == nil {
			entity.SetWatcher(w)
			watcher = w
		}
	}

	if watcher != nil {
		newProps := map[Property]*proto.Value{}
		for _, p := range resp.GetProperties() {
			bits := strings.Split(p.Property, ".")
			name := bits[len(bits)-1]
			prefix := ""
			if len(bits) > 1 {
				prefix = strings.Join(bits[:len(bits)-1], ".")
			}
			newProps[knownPrefixProperty(prefix, name)] = p.Value
		}
		watcher.AppendKnownValues(newProps)
	}

	for _, option := range retrieve {
		if observe, ok := option.(RetrieveObserver); ok {
			observe.ObserveRetrieve(resp)
		}
	}

	if observe, ok := dst.(RetrieveObserver); ok {
		observe.ObserveRetrieve(resp)
	}

	if gr, ok := dst.(GenericResult); ok {
		return UnmarshalGeneric(resp, gr)
	}

	return Unmarshal(resp, dst)
}

type RetrieveObserver interface {
	ObserveRetrieve(resp *proto.EntityResponse)
}

// GetSharedByID retrieves an entity by the given retrieveBy, storing the result in dst
func (a *Actor) GetSharedByID(ctx context.Context, owner *proto.VendorApp, entityID ID, dst interface{}, retrieve ...RetrieveOption) error {
	retrieveBy := ByEntityID(Type(dst), entityID)
	entityRequest := retrieveBy.BaseRequest()
	entityRequest.Authorization = a.Authorization()
	for _, rOpt := range retrieve {
		rOpt.Apply(entityRequest.View)
		if reOpt, ok := rOpt.(RetrieveEntityOption); ok {
			reOpt.ApplyRequest(entityRequest)
		}
	}

	view := entityRequest.View

	// set source
	for _, p := range view.Properties {
		p.Source = owner
	}

	entityRequest.Schema = &proto.Key{Key: Type(dst), Source: owner}

	resp, err := a.connection.Retrieve(ctx, entityRequest)
	if err != nil {
		return err
	}

	for _, option := range retrieve {
		if observe, ok := option.(RetrieveObserver); ok {
			observe.ObserveRetrieve(resp)
		}
	}

	if observe, ok := dst.(RetrieveObserver); ok {
		observe.ObserveRetrieve(resp)
	}

	if gr, ok := dst.(GenericResult); ok {
		return UnmarshalGeneric(resp, gr)
	}

	return Unmarshal(resp, dst)
}
