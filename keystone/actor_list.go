package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
)

// List returns a list of entities within an active set
func (a *Actor) List(ctx context.Context, entityType string, retrieveProperties []string, options ...FindOption) ([]*proto.EntityResponse, error) {
	listRequest := &proto.ListRequest{
		Authorization: a.Authorization(),
		Schema:        &proto.Key{Key: entityType, Source: a.Authorization().Source},
		Properties:    retrieveProperties,
	}

	fReq := &filterRequest{}
	for _, opt := range options {
		opt.Apply(fReq)
	}

	listRequest.Filters = fReq.Filters
	listRequest.Page = &proto.PageRequest{
		PerPage:    fReq.PerPage,
		PageNumber: fReq.PageNumber,
	}

	if fReq.SortProperty != "" {
		listRequest.Sort = []*proto.PropertySort{{
			Property:   fReq.SortProperty,
			Descending: fReq.SortDescending,
		}}
	}

	resp, err := a.connection.List(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	return resp.Entities, nil
}