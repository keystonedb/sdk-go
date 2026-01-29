package keystone

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/proto"
)

// DailyEntities retrieves entities created on a specific date for a given schema type.
// It returns a map of creation ID to entity ID, along with pagination information.
func (a *Actor) DailyEntities(ctx context.Context, schemaType string, date *proto.Date, opts ...DailyEntitiesOption) (*proto.DailyEntityResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	options := &dailyEntitiesOptions{}
	for _, opt := range opts {
		opt(options)
	}

	req := &proto.DailyEntityRequest{
		Authorization: a.Authorization(),
		Schema:        &proto.Key{Key: schemaType, Source: a.Authorization().Source},
		Date:          date,
		AfterId:       options.afterID,
		ReverseOrder:  options.reverseOrder,
		Limit:         options.limit,
	}

	return a.connection.DailyEntities(ctx, req)
}

type dailyEntitiesOptions struct {
	afterID      string
	reverseOrder bool
	limit        int32
}

// DailyEntitiesOption is a functional option for the DailyEntities method
type DailyEntitiesOption func(*dailyEntitiesOptions)

// WithDailyEntitiesAfterID sets the cursor for pagination (fetch entities after this ID)
func WithDailyEntitiesAfterID(afterID string) DailyEntitiesOption {
	return func(o *dailyEntitiesOptions) {
		o.afterID = afterID
	}
}

// WithDailyEntitiesReverseOrder sets whether to return results in reverse order
func WithDailyEntitiesReverseOrder(reverse bool) DailyEntitiesOption {
	return func(o *dailyEntitiesOptions) {
		o.reverseOrder = reverse
	}
}

// WithDailyEntitiesLimit sets the maximum number of entities to return
func WithDailyEntitiesLimit(limit int32) DailyEntitiesOption {
	return func(o *dailyEntitiesOptions) {
		o.limit = limit
	}
}
