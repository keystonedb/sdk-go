package keystone

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/proto"
)

func EidHash(entityID ID) string {
	return "des" + entityID.String() + "troy"
}

func (a *Actor) PermanentlyDestroyEntity(ctx context.Context, schemaType string, entityID ID, eidHash, deletionReason string) (bool, error) {
	if len(deletionReason) < 10 {
		return false, errors.New("a valid deletion reason must be provided")
	}

	if eidHash != EidHash(entityID) {
		// This is just an additional hoop to jump through to make sure you really wanted to call this
		return false, errors.New("invalid entity ID hash")
	}

	destroyed, err := a.Connection().Destroy(ctx, &proto.DestroyRequest{
		Authorization: a.Authorization(),
		Schema:        &proto.Key{Key: schemaType, Source: a.VendorApp()},
		Eid:           entityID.String(),
		Reason:        deletionReason,
	})

	if err != nil {
		return false, err
	}

	return destroyed.GetDestroyed(), nil
}
