package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func TestActor_SchemaStatistics_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.SchemaStatistics(context.Background(), "TestSchema")
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
}

func TestActor_SchemaStatistics_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.SchemaStatistics(context.Background(), "TestSchema")
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
}

func TestSchemaStatisticsOptions(t *testing.T) {
	// Test WithStatsCreatedFrom
	opts := &schemaStatisticsOptions{}
	fromDate := &proto.Date{Year: 2024, Month: 1, Day: 1}
	WithStatsCreatedFrom(fromDate)(opts)
	if opts.createdFrom != fromDate {
		t.Errorf("WithStatsCreatedFrom: expected date to be set")
	}

	// Test WithStatsCreatedUntil
	opts = &schemaStatisticsOptions{}
	untilDate := &proto.Date{Year: 2024, Month: 12, Day: 31}
	WithStatsCreatedUntil(untilDate)(opts)
	if opts.createdUntil != untilDate {
		t.Errorf("WithStatsCreatedUntil: expected date to be set")
	}

	// Test WithStatsIncludeBreakdown
	opts = &schemaStatisticsOptions{}
	WithStatsIncludeBreakdown(true)(opts)
	if !opts.includeBreakdown {
		t.Errorf("WithStatsIncludeBreakdown: expected true, got false")
	}

	// Test WithStatsDayLimit
	opts = &schemaStatisticsOptions{}
	WithStatsDayLimit(30)(opts)
	if opts.dayLimit != 30 {
		t.Errorf("WithStatsDayLimit: expected 30, got %d", opts.dayLimit)
	}
}
