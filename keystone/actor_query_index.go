package keystone

import (
	"context"

	"github.com/keystonedb/sdk-go/proto"
)

// Deprecated: use keystone.Connection.List instead
func (a *Actor) List(ctx context.Context, entityType string, retrieveProperties []string, options ...FindOption) ([]*proto.EntityResponse, error) {
	return a.QueryIndex(ctx, entityType, retrieveProperties, options...)
}

// QueryIndex returns a list of entities within the index
func (a *Actor) QueryIndex(ctx context.Context, entityType string, retrieveProperties []string, options ...FindOption) ([]*proto.EntityResponse, error) {
	listRequest := &proto.QueryIndexRequest{
		Authorization: a.Authorization(),
		Schema:        &proto.Key{Key: entityType, Source: a.Authorization().Source},
		Properties:    retrieveProperties,
	}

	/*hasState := false
	for _, opt := range options {
		if fOpt, ok := opt.(propertyFilter); ok {
			if fOpt.key == statePropertyName {
				hasState = true
				break
			}
		}
	}
	if !hasState {
		// Default to only active entities if no state filter is provided
		options = append(options, OnlyActive())
	}*/

	fReq := &filterRequest{}
	for _, opt := range options {
		opt.Apply(fReq)
	}

	listRequest.EntityIds = fReq.EntityIds
	listRequest.ParentEntityId = fReq.ParentEntityID
	listRequest.Filters = fReq.Filters
	listRequest.Sort = fReq.sortBy
	listRequest.Page = &proto.PageRequest{
		PerPage:    fReq.PerPage,
		PageNumber: fReq.PageNumber,
	}

	resp, err := a.connection.QueryIndex(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	return resp.Entities, nil
}
