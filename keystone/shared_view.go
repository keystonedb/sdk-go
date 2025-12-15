package keystone

import (
	"context"

	"github.com/keystonedb/sdk-go/proto"
)

type SharedView struct {
	properties       map[string]string
	piiProperties    map[string]string
	secureProperties map[string]string
	comment          string
	entityID         ID
	allWorkspaces    bool
	entityType       string // Specific type if no entity ID specified
}

func NewSharedView(properties ...string) *SharedView {
	v := &SharedView{
		properties:       make(map[string]string),
		piiProperties:    make(map[string]string),
		secureProperties: make(map[string]string),
	}

	for _, p := range properties {
		v.properties[p] = p
	}

	return v
}

func (s *SharedView) ForType(forType string) *SharedView {
	s.entityType = forType
	return s
}

func (s *SharedView) ForDynamicProperties() *SharedView {
	s.entityType = "__dynamic"
	return s
}

func (s *SharedView) ForEntity(entityID ID) *SharedView {
	s.entityID = entityID
	return s
}

func (s *SharedView) ForAllWorkspaces() *SharedView {
	s.allWorkspaces = true
	return s
}

func (s *SharedView) Add(property string, allowPii, allowSecure bool) *SharedView {
	s.properties[property] = property
	if allowPii {
		s.piiProperties[property] = property
	}
	if allowSecure {
		s.secureProperties[property] = property
	}
	return s
}

func (s *SharedView) WithComment(comment string) *SharedView {
	s.comment = comment
	return s
}

func (a *Actor) ShareView(ctx context.Context, with *proto.VendorApp, def *SharedView) (*proto.SharedViewResponse, error) {
	req := &proto.ShareViewRequest{
		Authorization:         a.Authorization(),
		EntityId:              string(def.entityID),
		AllWorkspaces:         def.allWorkspaces,
		EntityType:            def.entityType,
		ShareWith:             with,
		Comment:               def.comment,
		AllowProperties:       mapKeys(def.properties),
		AllowPiiProperties:    mapKeys(def.piiProperties),
		AllowSecureProperties: mapKeys(def.secureProperties),
	}

	return a.connection.ShareView(ctx, req)
}

func (a *Actor) SharedViews(ctx context.Context, with *proto.VendorApp, entityID ID, entityType string, anyWorkspace bool) (*proto.SharedViewsResponse, error) {
	req := &proto.SharedViewsRequest{
		Authorization: a.Authorization(),
		ShareWith:     with,
		EntityId:      string(entityID),
		EntityType:    entityType,
		AllWorkspaces: anyWorkspace,
	}
	return a.connection.SharedViews(ctx, req)
}
