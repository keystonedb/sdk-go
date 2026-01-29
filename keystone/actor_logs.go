package keystone

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/proto"
)

// Logs retrieves audit log entries for an entity
func (a *Actor) Logs(ctx context.Context, entityID string, opts ...LogsOption) ([]*proto.EntityLog, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	options := &logsOptions{}
	for _, opt := range opts {
		opt(options)
	}

	req := &proto.LogsRequest{
		Authorization: a.Authorization(),
		EntityId:      entityID,
		Levels:        options.levels,
		MinLevel:      options.minLevel,
		Window:        options.window,
	}

	resp, err := a.connection.Logs(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetLogs(), nil
}

type logsOptions struct {
	levels   []proto.LogLevel
	minLevel proto.LogLevel
	window   *proto.Window
}

// LogsOption is a functional option for the Logs method
type LogsOption func(*logsOptions)

// WithLogsMinLevel sets the minimum log level to retrieve
func WithLogsMinLevel(level proto.LogLevel) LogsOption {
	return func(o *logsOptions) {
		o.minLevel = level
	}
}

// WithLogsLevels sets the specific log levels to retrieve
func WithLogsLevels(levels ...proto.LogLevel) LogsOption {
	return func(o *logsOptions) {
		o.levels = levels
	}
}

// WithLogsWindow sets the time window for log retrieval
func WithLogsWindow(window *proto.Window) LogsOption {
	return func(o *logsOptions) {
		o.window = window
	}
}
