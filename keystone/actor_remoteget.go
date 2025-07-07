package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
)

// RemoteGet retrieves an entity by the given ID, storing the result in dst
func (a *Actor) RemoteGet(ctx context.Context, remoteID ID, dst interface{}, retrieve ...RetrieveOption) error {
	entityRequest := &proto.EntityRequest{
		View:     &proto.EntityView{},
		EntityId: string(remoteID),
	}
	entityRequest.Authorization = a.Authorization()

	for _, rOpt := range retrieve {
		rOpt.Apply(entityRequest.View)
		if reOpt, ok := rOpt.(RetrieveEntityOption); ok {
			reOpt.ApplyRequest(entityRequest)
		}
	}

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
