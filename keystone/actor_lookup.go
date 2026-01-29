package keystone

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/proto"
)

// Lookup performs a reverse lookup to find entities by a property value.
// It returns all matching entity references.
func (a *Actor) Lookup(ctx context.Context, property, value string, opts ...LookupOption) ([]*proto.EntityReference, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	options := &lookupOptions{}
	for _, opt := range opts {
		opt(options)
	}

	req := &proto.LookupRequest{
		Authorization: a.Authorization(),
		Lookup:        value,
		SchemeId:      options.schemeID,
		Property:      &property,
	}

	resp, err := a.connection.client.Lookup(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetResults(), nil
}

// LookupOne performs a reverse lookup and returns the first matching entity reference.
// If no results are found, it returns nil without an error.
func (a *Actor) LookupOne(ctx context.Context, property, value string, opts ...LookupOption) (*proto.EntityReference, error) {
	results, err := a.Lookup(ctx, property, value, opts...)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return results[0], nil
}

type lookupOptions struct {
	schemeID string
}

// LookupOption is a functional option for the Lookup and LookupOne methods
type LookupOption func(*lookupOptions)

// WithLookupSchemeID filters lookup results to a specific entity scheme/type
func WithLookupSchemeID(schemeID string) LookupOption {
	return func(o *lookupOptions) {
		o.schemeID = schemeID
	}
}
