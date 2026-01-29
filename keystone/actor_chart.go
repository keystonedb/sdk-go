package keystone

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChartTimeSeries retrieves time-series chart data for a schema type
func (a *Actor) ChartTimeSeries(ctx context.Context, schemaType string, opts ...ChartTimeSeriesOption) (map[string]*proto.ChartTimeSeriesResponse_ChartSeries, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	options := &chartTimeSeriesOptions{}
	for _, opt := range opts {
		opt(options)
	}

	req := &proto.ChartTimeSeriesRequest{
		Authorization:   a.Authorization(),
		Schema:          &proto.Key{Key: schemaType, Source: a.VendorApp()},
		From:            options.from,
		Until:           options.until,
		Interval:        options.interval,
		Timezone:        options.timezone,
		SeriesProperty:  options.seriesProperty,
		Aggregations:    options.aggregations,
		PropertyFilters: options.propertyFilters,
		FillMissing:     options.fillMissing,
	}

	resp, err := a.connection.ChartTimeSeries(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetSeries(), nil
}

type chartTimeSeriesOptions struct {
	from            *timestamppb.Timestamp
	until           *timestamppb.Timestamp
	interval        string
	timezone        string
	seriesProperty  string
	aggregations    []*proto.PropertyAggregation
	propertyFilters []*proto.PropertyFilter
	fillMissing     bool
}

// ChartTimeSeriesOption is a functional option for the ChartTimeSeries method
type ChartTimeSeriesOption func(*chartTimeSeriesOptions)

// WithChartFrom sets the start timestamp for the time series
func WithChartFrom(from *timestamppb.Timestamp) ChartTimeSeriesOption {
	return func(o *chartTimeSeriesOptions) {
		o.from = from
	}
}

// WithChartUntil sets the end timestamp for the time series
func WithChartUntil(until *timestamppb.Timestamp) ChartTimeSeriesOption {
	return func(o *chartTimeSeriesOptions) {
		o.until = until
	}
}

// WithChartInterval sets the interval for grouping data points (e.g., "1h", "1d", "1w")
func WithChartInterval(interval string) ChartTimeSeriesOption {
	return func(o *chartTimeSeriesOptions) {
		o.interval = interval
	}
}

// WithChartTimezone sets the timezone for the time series
func WithChartTimezone(timezone string) ChartTimeSeriesOption {
	return func(o *chartTimeSeriesOptions) {
		o.timezone = timezone
	}
}

// WithChartSeriesProperty sets the property to use for series grouping
func WithChartSeriesProperty(property string) ChartTimeSeriesOption {
	return func(o *chartTimeSeriesOptions) {
		o.seriesProperty = property
	}
}

// WithChartAggregations sets the property aggregations for the chart
func WithChartAggregations(aggregations ...*proto.PropertyAggregation) ChartTimeSeriesOption {
	return func(o *chartTimeSeriesOptions) {
		o.aggregations = aggregations
	}
}

// WithChartFilters sets the property filters for the time series query
func WithChartFilters(filters ...*proto.PropertyFilter) ChartTimeSeriesOption {
	return func(o *chartTimeSeriesOptions) {
		o.propertyFilters = filters
	}
}

// WithChartFillMissing enables filling missing data points in the time series
func WithChartFillMissing(fill bool) ChartTimeSeriesOption {
	return func(o *chartTimeSeriesOptions) {
		o.fillMissing = fill
	}
}
