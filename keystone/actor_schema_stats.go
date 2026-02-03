package keystone

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/proto"
)

// SchemaStatistics retrieves statistics for a schema type
func (a *Actor) SchemaStatistics(ctx context.Context, schemaType string, opts ...SchemaStatisticsOption) (*proto.SchemaStatisticsResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	options := &schemaStatisticsOptions{}
	for _, opt := range opts {
		opt(options)
	}

	req := &proto.SchemaStatisticsRequest{
		Authorization: a.Authorization(),
		Schema: &proto.Key{
			Source: a.VendorApp(),
			Key:    schemaType,
		},
		CreatedFrom:      options.createdFrom,
		CreatedUntil:     options.createdUntil,
		IncludeBreakdown: options.includeBreakdown,
		DayLimit:         options.dayLimit,
	}

	return a.connection.SchemaStatistics(ctx, req)
}

type schemaStatisticsOptions struct {
	createdFrom      *proto.Date
	createdUntil     *proto.Date
	includeBreakdown bool
	dayLimit         int32
}

// SchemaStatisticsOption is a functional option for the SchemaStatistics method
type SchemaStatisticsOption func(*schemaStatisticsOptions)

// WithStatsCreatedFrom sets the inclusive start date for statistics
func WithStatsCreatedFrom(date *proto.Date) SchemaStatisticsOption {
	return func(o *schemaStatisticsOptions) {
		o.createdFrom = date
	}
}

// WithStatsCreatedUntil sets the end date for statistics
func WithStatsCreatedUntil(date *proto.Date) SchemaStatisticsOption {
	return func(o *schemaStatisticsOptions) {
		o.createdUntil = date
	}
}

// WithStatsIncludeBreakdown enables daily breakdown in the response
func WithStatsIncludeBreakdown(include bool) SchemaStatisticsOption {
	return func(o *schemaStatisticsOptions) {
		o.includeBreakdown = include
	}
}

// WithStatsDayLimit sets the maximum number of days with results to return
func WithStatsDayLimit(limit int32) SchemaStatisticsOption {
	return func(o *schemaStatisticsOptions) {
		o.dayLimit = limit
	}
}
