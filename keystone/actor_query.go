package keystone

import (
	"context"
	"fmt"

	"github.com/keystonedb/sdk-go/proto"
	gproto "google.golang.org/protobuf/proto"
)

// Query queries entities using filters, projections and cursor pagination options.
// Without Select or SelectSecure it returns entity metadata only.
// Use Limit(size) and Cursor(nextCursor) to paginate results.
func (a *Actor) Query(ctx context.Context, entityType string, options ...QueryOption) (*proto.QueryResponse, error) {
	return a.QueryWithRequest(ctx, entityType, nil, options...)
}

// QueryWithRequest queries entities or aggregate rows using a protobuf request.
// The request is copied before options, actor authorization and source schema are
// applied. Options extend projections and filters and override scalar settings.
func (a *Actor) QueryWithRequest(ctx context.Context, entityType string, request *proto.QueryRequest, options ...QueryOption) (*proto.QueryResponse, error) {
	req := &proto.QueryRequest{}
	if request != nil {
		req = gproto.Clone(request).(*proto.QueryRequest)
	}
	req.Authorization = a.Authorization()
	req.From = &proto.Key{Key: entityType, Source: req.Authorization.Source}
	filters := &filterRequest{}
	idsSet := false
	for _, opt := range options {
		if opt == nil {
			return nil, fmt.Errorf("query: nil option")
		}
		opt.Apply(filters)
		if _, ok := opt.(withEntityIDs); ok {
			idsSet = true
		}
	}
	if filters.PageNumber != 0 {
		return nil, fmt.Errorf("query: page numbers are unsupported; use a cursor")
	}
	if filters.ListObjects || len(filters.ObjectPaths) > 0 {
		return nil, fmt.Errorf("query: object loading is unsupported")
	}
	for _, apply := range filters.queryOptions {
		apply(req)
	}
	req.Where = append(req.Where, filters.Filters...)
	req.OrderBy = append(req.OrderBy, filters.sortBy...)
	if filters.PerPage != 0 {
		req.Limit = filters.PerPage
	}
	if filters.ParentEntityID != "" {
		req.ParentEntityId = filters.ParentEntityID
	}
	if idsSet {
		req.EntityIds = &proto.EntityIDList{Values: filters.EntityIds}
	}
	predicates := []*proto.QueryPredicate{}
	if req.Predicate != nil {
		predicates = append(predicates, req.Predicate)
	}
	for _, label := range filters.Labels {
		predicates = append(predicates, &proto.QueryPredicate{Condition: &proto.QueryPredicate_Label{Label: label}})
	}
	if rel := filters.RelationOf; rel != nil {
		relationship := &proto.QueryRelationship{Type: rel.Relationship}
		if rel.SourceId != "" {
			relationship.Direction = proto.QueryRelationship_INCOMING
			relationship.EntityIds = []string{rel.SourceId}
		} else {
			relationship.Direction = proto.QueryRelationship_OUTGOING
			relationship.EntityIds = []string{rel.DestinationId}
		}
		predicates = append(predicates, &proto.QueryPredicate{Condition: &proto.QueryPredicate_Relationship{Relationship: relationship}})
	}
	if len(predicates) == 1 {
		req.Predicate = predicates[0]
	} else if len(predicates) > 1 {
		req.Predicate = &proto.QueryPredicate{Condition: &proto.QueryPredicate_All{All: &proto.QueryPredicateGroup{Predicates: predicates}}}
	}
	return a.connection.Query(ctx, req)
}
