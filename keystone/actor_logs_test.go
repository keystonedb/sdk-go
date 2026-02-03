package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

func TestActor_Logs_NilActor(t *testing.T) {
	var actor *Actor
	_, err := actor.Logs(context.Background(), "entity-123")
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
}

func TestActor_Logs_NilConnection(t *testing.T) {
	actor := &Actor{}
	_, err := actor.Logs(context.Background(), "entity-123")
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
}

func TestActor_Log_NilActor(t *testing.T) {
	var actor *Actor
	err := actor.Log(context.Background(), "entity-123", proto.LogLevel_Info, "test message")
	if err == nil {
		t.Error("expected error for nil actor, got nil")
	}
}

func TestActor_Log_NilConnection(t *testing.T) {
	actor := &Actor{}
	err := actor.Log(context.Background(), "entity-123", proto.LogLevel_Info, "test message")
	if err == nil {
		t.Error("expected error for nil connection, got nil")
	}
}

func TestLogOptions(t *testing.T) {
	// Test WithLogReference
	opts := &logOptions{}
	WithLogReference("test-ref")(opts)
	if opts.reference != "test-ref" {
		t.Errorf("WithLogReference: expected 'test-ref', got %s", opts.reference)
	}

	// Test WithLogTraceID
	opts = &logOptions{}
	WithLogTraceID("trace-123")(opts)
	if opts.traceID != "trace-123" {
		t.Errorf("WithLogTraceID: expected 'trace-123', got %s", opts.traceID)
	}

	// Test WithLogData
	opts = &logOptions{}
	data := map[string]string{"key": "value"}
	WithLogData(data)(opts)
	if opts.data["key"] != "value" {
		t.Errorf("WithLogData: expected data to be set")
	}

	// Test WithLogAuditUser
	opts = &logOptions{}
	user := &proto.User{UserId: "user-123", Client: "test-client"}
	WithLogAuditUser(user)(opts)
	if opts.auditUser == nil || opts.auditUser.UserId != "user-123" {
		t.Errorf("WithLogAuditUser: expected audit user to be set")
	}
}

func TestLogsOptions(t *testing.T) {
	// Test WithLogsMinLevel
	opts := &logsOptions{}
	WithLogsMinLevel(proto.LogLevel_Warn)(opts)
	if opts.minLevel != proto.LogLevel_Warn {
		t.Errorf("WithLogsMinLevel: expected %v, got %v", proto.LogLevel_Warn, opts.minLevel)
	}

	// Test WithLogsLevels
	opts = &logsOptions{}
	WithLogsLevels(proto.LogLevel_Error, proto.LogLevel_Critical)(opts)
	if len(opts.levels) != 2 {
		t.Errorf("WithLogsLevels: expected 2 levels, got %d", len(opts.levels))
	}
	if opts.levels[0] != proto.LogLevel_Error || opts.levels[1] != proto.LogLevel_Critical {
		t.Errorf("WithLogsLevels: expected [Error, Critical], got %v", opts.levels)
	}

	// Test WithLogsWindow
	opts = &logsOptions{}
	window := &proto.Window{}
	WithLogsWindow(window)(opts)
	if opts.window != window {
		t.Errorf("WithLogsWindow: expected window to be set")
	}
}
