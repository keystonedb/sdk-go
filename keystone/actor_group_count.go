package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
)

// GroupCount returns a list of entities within an active set
func (a *Actor) GroupCount(ctx context.Context, entityType string, groupBy []string, options ...FindOption) ([]*proto.GroupCountResponse_Result, error) {
	listRequest := &proto.GroupCountRequest{
		Authorization: a.Authorization(),
		Schema:        &proto.Key{Key: entityType, Source: a.Authorization().Source},
		Properties:    groupBy,
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

	resp, err := a.connection.GroupCount(ctx, listRequest)
	if err != nil {
		return nil, err
	}
	return resp.Results, nil
}
