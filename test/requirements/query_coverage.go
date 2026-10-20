package requirements

import (
	"context"
	"strings"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
)

// QueryCoverage lets requirements exercise their existing assertions through Query.
// AllProperties supplies an explicit projection where legacy Find selected all.
type QueryCoverage struct {
	UseQuery      bool
	AllProperties []string
}

func (q QueryCoverage) Find(actor *keystone.Actor, ctx context.Context, entityType string, retrieve keystone.RetrieveOption, options ...keystone.FindOption) ([]*proto.EntityResponse, error) {
	if !q.UseQuery {
		return actor.Find(ctx, entityType, retrieve, options...)
	}
	view := &proto.EntityView{}
	if retrieve != nil {
		retrieve.Apply(view)
	}
	req := &proto.QueryRequest{Include: &proto.QueryInclude{
		Labels: view.Labels, Relationships: view.RelationshipByType,
		RelationshipCount: view.RelationshipCount, RelationshipCountTypes: view.RelationshipCountType,
	}}
	for _, property := range view.Properties {
		selected := property.Properties
		if len(selected) == 0 {
			selected = q.AllProperties
		}
		for _, name := range selected {
			parts := strings.Split(name, ".")
			for i := range parts {
				parts[i] = keystone.PropertyName(parts[i])
			}
			req.Select = append(req.Select, strings.Join(parts, "."))
		}
		if property.Decrypt {
			req.Decrypt = append(req.Decrypt, selected...)
		}
	}
	resp, err := actor.QueryWithRequest(ctx, entityType, req, options...)
	return resp.GetEntities(), err
}

func (q QueryCoverage) QueryIndex(actor *keystone.Actor, ctx context.Context, entityType string, properties []string, options ...keystone.FindOption) ([]*proto.EntityResponse, error) {
	if !q.UseQuery {
		return actor.QueryIndex(ctx, entityType, properties, options...)
	}
	resp, err := actor.QueryWithRequest(ctx, entityType, &proto.QueryRequest{Select: properties}, options...)
	return resp.GetEntities(), err
}
