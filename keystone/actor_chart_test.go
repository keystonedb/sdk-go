package keystone

import (
	"context"
	"testing"
	"time"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestActor_ChartTimeSeries_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.ChartTimeSeries(context.Background(), "TestSchema")
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
}

func TestActor_ChartTimeSeries_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.ChartTimeSeries(context.Background(), "TestSchema")
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
}

func TestChartOptions(t *testing.T) {
	// Test WithChartFrom
	opts := &chartTimeSeriesOptions{}
	fromTime := timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	WithChartFrom(fromTime)(opts)
	if opts.from != fromTime {
		t.Errorf("WithChartFrom: expected from time to be set")
	}

	// Test WithChartUntil
	opts = &chartTimeSeriesOptions{}
	untilTime := timestamppb.New(time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC))
	WithChartUntil(untilTime)(opts)
	if opts.until != untilTime {
		t.Errorf("WithChartUntil: expected until time to be set")
	}

	// Test WithChartInterval
	opts = &chartTimeSeriesOptions{}
	WithChartInterval("1h")(opts)
	if opts.interval != "1h" {
		t.Errorf("WithChartInterval: expected '1h', got %s", opts.interval)
	}

	// Test WithChartTimezone
	opts = &chartTimeSeriesOptions{}
	WithChartTimezone("America/New_York")(opts)
	if opts.timezone != "America/New_York" {
		t.Errorf("WithChartTimezone: expected 'America/New_York', got %s", opts.timezone)
	}

	// Test WithChartSeriesProperty
	opts = &chartTimeSeriesOptions{}
	WithChartSeriesProperty("category")(opts)
	if opts.seriesProperty != "category" {
		t.Errorf("WithChartSeriesProperty: expected 'category', got %s", opts.seriesProperty)
	}

	// Test WithChartAggregations
	opts = &chartTimeSeriesOptions{}
	aggs := []*proto.PropertyAggregation{
		{Property: "amount", Type: proto.PropertyAggregation_Sum, Alias: "total_amount"},
		{Property: "count", Type: proto.PropertyAggregation_Count, Alias: "total_count"},
	}
	WithChartAggregations(aggs...)(opts)
	if len(opts.aggregations) != 2 {
		t.Errorf("WithChartAggregations: expected 2 aggregations, got %d", len(opts.aggregations))
	}
	if opts.aggregations[0].Property != "amount" {
		t.Errorf("WithChartAggregations: expected first aggregation property 'amount', got %s", opts.aggregations[0].Property)
	}

	// Test WithChartFilters
	opts = &chartTimeSeriesOptions{}
	filters := []*proto.PropertyFilter{
		{Property: "status", Operator: proto.Operator_Equal},
	}
	WithChartFilters(filters...)(opts)
	if len(opts.propertyFilters) != 1 {
		t.Errorf("WithChartFilters: expected 1 filter, got %d", len(opts.propertyFilters))
	}

	// Test WithChartFillMissing
	opts = &chartTimeSeriesOptions{}
	WithChartFillMissing(true)(opts)
	if !opts.fillMissing {
		t.Errorf("WithChartFillMissing: expected true, got false")
	}
}
