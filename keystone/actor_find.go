package keystone

import (
	"context"

	"github.com/keystonedb/sdk-go/proto"
)

// Find returns a list of entities matching the given entityType and retrieveProperties
func (a *Actor) Find(ctx context.Context, entityType string, retrieve RetrieveOption, options ...FindOption) ([]*proto.EntityResponse, error) {
	findRequest := &proto.FindRequest{
		Authorization: a.Authorization(),
		Schema:        &proto.Key{Key: entityType, Source: a.Authorization().Source},
		View:          &proto.EntityView{},
	}

	if retrieve != nil {
		retrieve.Apply(findRequest.View)
	}

	fReq := &filterRequest{Properties: []*proto.PropertyRequest{}}

	for _, opt := range options {
		opt.Apply(fReq)
	}

	findRequest.PropertyFilters = fReq.Filters
	findRequest.LabelFilters = fReq.Labels
	findRequest.RelationOf = fReq.RelationOf
	findRequest.ParentEntityId = fReq.ParentEntityID
	findRequest.EntityIds = fReq.EntityIds

	resp, err := a.connection.Find(ctx, findRequest)
	if err != nil {
		return nil, err
	}
	return resp.Entities, nil
}
