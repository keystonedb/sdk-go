package keystone

import (
	"context"
	"errors"
	"github.com/keystonedb/sdk-go/proto"
	gproto "google.golang.org/protobuf/proto"
)

// EntityAlias identifies a record in a linked app. Empty RoutingKey selects the
// default namespace; it never searches across routing keys.
type EntityAlias = proto.EntityAlias

// NewEntityAlias identifies a linked app's record. Vendor, app and external reference
// are required; no values are inferred from the actor. Use WithRouting to select a
// routing namespace instead of the default empty namespace.
func NewEntityAlias(vendor, app, extRef string) *EntityAlias {
	return &EntityAlias{Source: &proto.VendorApp{VendorId: vendor, AppId: app}, ExternalId: extRef}
}

func (a *Actor) aliasReady() error {
	if a == nil || a.connection == nil || a.connection.client == nil {
		return errors.New("actor or connection is nil")
	}
	return nil
}

// AttachAlias associates an existing canonical entity with an external identity.
// The schema owner or the linked app can attach. Ownership cannot be transferred.
func (a *Actor) AttachAlias(ctx context.Context, entityType string, entityID ID, alias *EntityAlias, opts ...AliasOption) error {
	if err := a.aliasReady(); err != nil {
		return err
	}
	schema, err := a.aliasSchemaKey(entityType, opts)
	if err != nil {
		return err
	}
	resp, err := a.connection.AttachAlias(ctx, &proto.AliasMutationRequest{Authorization: a.Authorization(), Schema: schema, EntityId: string(entityID), Alias: alias})
	if err != nil {
		return err
	}
	if !resp.GetSuccess() {
		return errors.New("alias attachment was not successful")
	}
	return nil
}

// RemoveAlias deactivates an alias while preserving its permanent ownership.
func (a *Actor) RemoveAlias(ctx context.Context, entityType string, entityID ID, alias *EntityAlias, opts ...AliasOption) error {
	if err := a.aliasReady(); err != nil {
		return err
	}
	schema, err := a.aliasSchemaKey(entityType, opts)
	if err != nil {
		return err
	}
	resp, err := a.connection.RemoveAlias(ctx, &proto.AliasMutationRequest{Authorization: a.Authorization(), Schema: schema, EntityId: string(entityID), Alias: alias})
	if err != nil {
		return err
	}
	if !resp.GetSuccess() {
		return errors.New("alias removal was not successful")
	}
	return nil
}

// ResolveAlias returns only the canonical ID. A linked app must supply the owner's
// entity type and use WithAliasOwner for the owning app. No schema opt-in is required.
func (a *Actor) ResolveAlias(ctx context.Context, entityType string, alias *EntityAlias, opts ...AliasOption) (ID, error) {
	if err := a.aliasReady(); err != nil {
		return "", err
	}
	schema, err := a.aliasSchemaKey(entityType, opts)
	if err != nil {
		return "", err
	}
	resp, err := a.connection.ResolveAlias(ctx, &proto.ResolveAliasRequest{Authorization: a.Authorization(), Schema: schema, Alias: alias})
	if err != nil {
		return "", err
	}
	if resp.GetEntityId() == "" {
		return "", errors.New("alias resolution returned an empty entity ID")
	}
	return ID(resp.GetEntityId()), nil
}

// ListAliases returns active aliases. Only the schema owner may list them.
func (a *Actor) ListAliases(ctx context.Context, entityType string, entityID ID, opts ...AliasOption) ([]*EntityAlias, error) {
	if err := a.aliasReady(); err != nil {
		return nil, err
	}
	schema, err := a.aliasSchemaKey(entityType, opts)
	if err != nil {
		return nil, err
	}
	resp, err := a.connection.ListAliases(ctx, &proto.ListAliasesRequest{Authorization: a.Authorization(), Schema: schema, EntityId: string(entityID)})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("alias listing returned no response")
	}
	return resp.GetAliases(), nil
}

// GetByAlias loads an entity owned by the actor's app using an external identity.
func (a *Actor) GetByAlias(ctx context.Context, alias *EntityAlias, dst interface{}, retrieve ...RetrieveOption) error {
	if err := a.aliasReady(); err != nil {
		return err
	}
	if alias == nil {
		return errors.New("alias is required")
	}
	return a.Get(ctx, ByAlias(Type(dst), alias), dst, retrieve...)
}

type byAlias struct {
	entityType string
	alias      *EntityAlias
}

func ByAlias(entityType interface{}, alias *EntityAlias) RetrieveBy {
	if name, ok := entityType.(string); ok {
		return byAlias{name, alias}
	}
	return byAlias{Type(entityType), alias}
}
func (b byAlias) EntityType() string { return b.entityType }
func (b byAlias) BaseRequest() *proto.EntityRequest {
	var alias *EntityAlias
	if b.alias != nil {
		alias = gproto.Clone(b.alias).(*EntityAlias)
	}
	return &proto.EntityRequest{Alias: alias, View: &proto.EntityView{}}
}

// WithAliases requests active aliases and captures them after a successful Get.
// Pass the returned loader directly as a Get option; reuse replaces previous results.
func WithAliases() *AliasLoader { return &AliasLoader{} }

type AliasLoader struct{ aliases []*EntityAlias }

func (l *AliasLoader) Apply(view *proto.EntityView)               { view.Aliases = true }
func (l *AliasLoader) ObserveRetrieve(resp *proto.EntityResponse) { l.aliases = resp.GetAliases() }
func (l *AliasLoader) Aliases() []*EntityAlias                    { return l.aliases }

// AliasOption configures the canonical entity's owner, independently of the alias source.
type AliasOption func(*aliasOptions)
type aliasOptions struct{ vendor, app string }

// WithAliasOwner targets an entity type owned by another app. Omit it when the
// actor owns the entity type. This never changes the authenticated caller or alias.
func WithAliasOwner(vendorID, appID string) AliasOption {
	return func(o *aliasOptions) { o.vendor, o.app = vendorID, appID }
}

func (a *Actor) aliasSchemaKey(entityType string, opts []AliasOption) (*proto.Key, error) {
	if entityType == "" {
		return nil, errors.New("entity type is required")
	}
	o := aliasOptions{vendor: a.VendorID(), app: a.AppID()}
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	if o.vendor == "" || o.app == "" {
		return nil, errors.New("entity owner vendor and app are required")
	}
	return &proto.Key{Key: entityType, Source: &proto.VendorApp{VendorId: o.vendor, AppId: o.app}}, nil
}
