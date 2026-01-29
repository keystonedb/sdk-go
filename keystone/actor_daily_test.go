package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func TestActor_DailyEntities_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.DailyEntities(context.Background(), "test-schema", &proto.Date{Year: 2024, Month: 1, Day: 15})
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
}

func TestActor_DailyEntities_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.DailyEntities(context.Background(), "test-schema", &proto.Date{Year: 2024, Month: 1, Day: 15})
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
}

func TestDailyEntitiesOptions(t *testing.T) {
	// Test WithDailyEntitiesAfterID
	opts := &dailyEntitiesOptions{}
	WithDailyEntitiesAfterID("entity-123")(opts)
	if opts.afterID != "entity-123" {
		t.Errorf("WithDailyEntitiesAfterID: expected 'entity-123', got %s", opts.afterID)
	}

	// Test WithDailyEntitiesReverseOrder
	opts = &dailyEntitiesOptions{}
	WithDailyEntitiesReverseOrder(true)(opts)
	if opts.reverseOrder != true {
		t.Errorf("WithDailyEntitiesReverseOrder: expected true, got %v", opts.reverseOrder)
	}

	// Test WithDailyEntitiesLimit
	opts = &dailyEntitiesOptions{}
	WithDailyEntitiesLimit(50)(opts)
	if opts.limit != 50 {
		t.Errorf("WithDailyEntitiesLimit: expected 50, got %d", opts.limit)
	}
}
