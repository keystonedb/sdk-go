package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

// TestNilActor_NewMethods tests that all new Actor methods return an error when called on a nil actor.
// This is a table-driven test covering methods added in Phase 1:
// - Logs, Log (actor_logs.go)
// - Events (actor_events.go)
// - Lookup, LookupOne (actor_lookup.go)
// - AKVDel (actor_akv.go)
// - ChartTimeSeries (actor_chart.go)
// - DailyEntities (actor_daily.go)
// - SchemaStatistics (actor_schema_stats.go)
func TestNilActor_NewMethods(t *testing.T) {
	var nilActor *Actor
	ctx := context.Background()

	tests := []struct {
		name   string
		method func() error
	}{
		{
			name: "Logs",
			method: func() error {
				_, err := nilActor.Logs(ctx, "entity-123")
				return err
			},
		},
		{
			name: "Log",
			method: func() error {
				return nilActor.Log(ctx, "entity-123", proto.LogLevel_Info, "test message")
			},
		},
		{
			name: "Events",
			method: func() error {
				_, err := nilActor.Events(ctx, "entity-123")
				return err
			},
		},
		{
			name: "Lookup",
			method: func() error {
				_, err := nilActor.Lookup(ctx, "email", "test@example.com")
				return err
			},
		},
		{
			name: "LookupOne",
			method: func() error {
				_, err := nilActor.LookupOne(ctx, "email", "test@example.com")
				return err
			},
		},
		{
			name: "AKVDel",
			method: func() error {
				_, err := nilActor.AKVDel(ctx, "test-key")
				return err
			},
		},
		{
			name: "ChartTimeSeries",
			method: func() error {
				_, err := nilActor.ChartTimeSeries(ctx, "TestSchema")
				return err
			},
		},
		{
			name: "DailyEntities",
			method: func() error {
				_, err := nilActor.DailyEntities(ctx, "TestSchema", &proto.Date{Year: 2024, Month: 1, Day: 15})
				return err
			},
		},
		{
			name: "SchemaStatistics",
			method: func() error {
				_, err := nilActor.SchemaStatistics(ctx, "TestSchema")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.method()
			if err == nil {
				t.Errorf("%s: expected error for nil actor, got nil", tt.name)
			}
			if err != nil && err.Error() != "actor or connection is nil" {
				t.Errorf("%s: expected 'actor or connection is nil', got %q", tt.name, err.Error())
			}
		})
	}
}

// TestNilConnection_NewMethods tests that all new Actor methods return an error when called on an actor with a nil connection.
// This is a table-driven test covering methods added in Phase 1.
func TestNilConnection_NewMethods(t *testing.T) {
	actorWithNilConnection := &Actor{}
	ctx := context.Background()

	tests := []struct {
		name   string
		method func() error
	}{
		{
			name: "Logs",
			method: func() error {
				_, err := actorWithNilConnection.Logs(ctx, "entity-123")
				return err
			},
		},
		{
			name: "Log",
			method: func() error {
				return actorWithNilConnection.Log(ctx, "entity-123", proto.LogLevel_Info, "test message")
			},
		},
		{
			name: "Events",
			method: func() error {
				_, err := actorWithNilConnection.Events(ctx, "entity-123")
				return err
			},
		},
		{
			name: "Lookup",
			method: func() error {
				_, err := actorWithNilConnection.Lookup(ctx, "email", "test@example.com")
				return err
			},
		},
		{
			name: "LookupOne",
			method: func() error {
				_, err := actorWithNilConnection.LookupOne(ctx, "email", "test@example.com")
				return err
			},
		},
		{
			name: "AKVDel",
			method: func() error {
				_, err := actorWithNilConnection.AKVDel(ctx, "test-key")
				return err
			},
		},
		{
			name: "ChartTimeSeries",
			method: func() error {
				_, err := actorWithNilConnection.ChartTimeSeries(ctx, "TestSchema")
				return err
			},
		},
		{
			name: "DailyEntities",
			method: func() error {
				_, err := actorWithNilConnection.DailyEntities(ctx, "TestSchema", &proto.Date{Year: 2024, Month: 1, Day: 15})
				return err
			},
		},
		{
			name: "SchemaStatistics",
			method: func() error {
				_, err := actorWithNilConnection.SchemaStatistics(ctx, "TestSchema")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.method()
			if err == nil {
				t.Errorf("%s: expected error for nil connection, got nil", tt.name)
			}
			if err != nil && err.Error() != "actor or connection is nil" {
				t.Errorf("%s: expected 'actor or connection is nil', got %q", tt.name, err.Error())
			}
		})
	}
}

// TestNilActor_ExistingAKVMethods tests existing AKV methods for nil actor handling.
// These methods (AKVPut, AKVGet) were present before Phase 1 but are included for completeness.
func TestNilActor_ExistingAKVMethods(t *testing.T) {
	var nilActor *Actor
	ctx := context.Background()

	tests := []struct {
		name   string
		method func() error
	}{
		{
			name: "AKVPut",
			method: func() error {
				_, err := nilActor.AKVPut(ctx, AKV("test", "value"))
				return err
			},
		},
		{
			name: "AKVGet",
			method: func() error {
				_, err := nilActor.AKVGet(ctx, "test")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.method()
			if err == nil {
				t.Errorf("%s: expected error for nil actor, got nil", tt.name)
			}
			if err != nil && err.Error() != "actor or connection is nil" {
				t.Errorf("%s: expected 'actor or connection is nil', got %q", tt.name, err.Error())
			}
		})
	}
}

// TestNilConnection_ExistingAKVMethods tests existing AKV methods for nil connection handling.
func TestNilConnection_ExistingAKVMethods(t *testing.T) {
	actorWithNilConnection := &Actor{}
	ctx := context.Background()

	tests := []struct {
		name   string
		method func() error
	}{
		{
			name: "AKVPut",
			method: func() error {
				_, err := actorWithNilConnection.AKVPut(ctx, AKV("test", "value"))
				return err
			},
		},
		{
			name: "AKVGet",
			method: func() error {
				_, err := actorWithNilConnection.AKVGet(ctx, "test")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.method()
			if err == nil {
				t.Errorf("%s: expected error for nil connection, got nil", tt.name)
			}
			if err != nil && err.Error() != "actor or connection is nil" {
				t.Errorf("%s: expected 'actor or connection is nil', got %q", tt.name, err.Error())
			}
		})
	}
}

// TestNilActor_WithOptions verifies that nil actor errors occur before option processing,
// testing that options do not cause panics on nil actors.
func TestNilActor_WithOptions(t *testing.T) {
	var nilActor *Actor
	ctx := context.Background()

	t.Run("Logs with options", func(t *testing.T) {
		_, err := nilActor.Logs(ctx, "entity-123",
			WithLogsMinLevel(proto.LogLevel_Error),
			WithLogsLevels(proto.LogLevel_Critical),
		)
		if err == nil {
			t.Error("expected error for nil actor with options, got nil")
		}
	})

	t.Run("Log with options", func(t *testing.T) {
		err := nilActor.Log(ctx, "entity-123", proto.LogLevel_Info, "test",
			WithLogReference("ref-123"),
			WithLogTraceID("trace-123"),
		)
		if err == nil {
			t.Error("expected error for nil actor with options, got nil")
		}
	})

	t.Run("Events with options", func(t *testing.T) {
		_, err := nilActor.Events(ctx, "entity-123",
			WithEventTypes(&proto.Key{Key: "type1"}),
		)
		if err == nil {
			t.Error("expected error for nil actor with options, got nil")
		}
	})

	t.Run("Lookup with options", func(t *testing.T) {
		_, err := nilActor.Lookup(ctx, "email", "test@example.com",
			WithLookupSchemeID("user"),
		)
		if err == nil {
			t.Error("expected error for nil actor with options, got nil")
		}
	})

	t.Run("LookupOne with options", func(t *testing.T) {
		_, err := nilActor.LookupOne(ctx, "email", "test@example.com",
			WithLookupSchemeID("user"),
		)
		if err == nil {
			t.Error("expected error for nil actor with options, got nil")
		}
	})

	t.Run("ChartTimeSeries with options", func(t *testing.T) {
		_, err := nilActor.ChartTimeSeries(ctx, "TestSchema",
			WithChartInterval("1h"),
			WithChartTimezone("UTC"),
		)
		if err == nil {
			t.Error("expected error for nil actor with options, got nil")
		}
	})

	t.Run("DailyEntities with options", func(t *testing.T) {
		_, err := nilActor.DailyEntities(ctx, "TestSchema", &proto.Date{Year: 2024, Month: 1, Day: 15},
			WithDailyEntitiesLimit(100),
			WithDailyEntitiesReverseOrder(true),
		)
		if err == nil {
			t.Error("expected error for nil actor with options, got nil")
		}
	})

	t.Run("SchemaStatistics with options", func(t *testing.T) {
		_, err := nilActor.SchemaStatistics(ctx, "TestSchema",
			WithStatsIncludeBreakdown(true),
			WithStatsDayLimit(30),
		)
		if err == nil {
			t.Error("expected error for nil actor with options, got nil")
		}
	})
}

// TestNilConnection_WithOptions verifies that nil connection errors occur before option processing,
// testing that options do not cause panics on actors with nil connections.
func TestNilConnection_WithOptions(t *testing.T) {
	actorWithNilConnection := &Actor{}
	ctx := context.Background()

	t.Run("Logs with options", func(t *testing.T) {
		_, err := actorWithNilConnection.Logs(ctx, "entity-123",
			WithLogsMinLevel(proto.LogLevel_Error),
			WithLogsLevels(proto.LogLevel_Critical),
		)
		if err == nil {
			t.Error("expected error for nil connection with options, got nil")
		}
	})

	t.Run("Log with options", func(t *testing.T) {
		err := actorWithNilConnection.Log(ctx, "entity-123", proto.LogLevel_Info, "test",
			WithLogReference("ref-123"),
			WithLogTraceID("trace-123"),
		)
		if err == nil {
			t.Error("expected error for nil connection with options, got nil")
		}
	})

	t.Run("Events with options", func(t *testing.T) {
		_, err := actorWithNilConnection.Events(ctx, "entity-123",
			WithEventTypes(&proto.Key{Key: "type1"}),
		)
		if err == nil {
			t.Error("expected error for nil connection with options, got nil")
		}
	})

	t.Run("Lookup with options", func(t *testing.T) {
		_, err := actorWithNilConnection.Lookup(ctx, "email", "test@example.com",
			WithLookupSchemeID("user"),
		)
		if err == nil {
			t.Error("expected error for nil connection with options, got nil")
		}
	})

	t.Run("LookupOne with options", func(t *testing.T) {
		_, err := actorWithNilConnection.LookupOne(ctx, "email", "test@example.com",
			WithLookupSchemeID("user"),
		)
		if err == nil {
			t.Error("expected error for nil connection with options, got nil")
		}
	})

	t.Run("ChartTimeSeries with options", func(t *testing.T) {
		_, err := actorWithNilConnection.ChartTimeSeries(ctx, "TestSchema",
			WithChartInterval("1h"),
			WithChartTimezone("UTC"),
		)
		if err == nil {
			t.Error("expected error for nil connection with options, got nil")
		}
	})

	t.Run("DailyEntities with options", func(t *testing.T) {
		_, err := actorWithNilConnection.DailyEntities(ctx, "TestSchema", &proto.Date{Year: 2024, Month: 1, Day: 15},
			WithDailyEntitiesLimit(100),
			WithDailyEntitiesReverseOrder(true),
		)
		if err == nil {
			t.Error("expected error for nil connection with options, got nil")
		}
	})

	t.Run("SchemaStatistics with options", func(t *testing.T) {
		_, err := actorWithNilConnection.SchemaStatistics(ctx, "TestSchema",
			WithStatsIncludeBreakdown(true),
			WithStatsDayLimit(30),
		)
		if err == nil {
			t.Error("expected error for nil connection with options, got nil")
		}
	})
}
