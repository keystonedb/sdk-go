package keystone

import (
	"context"
	"errors"
	"time"

	"github.com/keystonedb/sdk-go/proto"
)

// AKVTimeEntry represents a single timestamped entry for AKV Time Track
type AKVTimeEntry struct {
	Key       string
	Workspace string // empty string for global
	Timestamp time.Time
	Data      []byte
	TTL       time.Duration // 0 = no expiry
}

// AKVTimePut writes one or more timestamped entries to the timeline store.
// Each entry requires a key, timestamp, and data. Workspace is optional (empty = global).
// Maximum 100 entries per call.
func (a *Actor) AKVTimePut(ctx context.Context, entries ...AKVTimeEntry) (*proto.GenericResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	req := &proto.AKVTimePutRequest{
		Authorization: a.Authorization(),
	}

	for _, e := range entries {
		entry := &proto.AKVTimeEntry{
			Key:           e.Key,
			Workspace:     e.Workspace,
			TimestampUnix: e.Timestamp.Unix(),
			Data:          e.Data,
		}
		if e.TTL > 0 {
			entry.TtlSeconds = int32(e.TTL.Seconds())
		}
		req.Entries = append(req.Entries, entry)
	}

	return a.connection.AKVTimePut(ctx, req)
}

// AKVTimeQueryMode mirrors proto.AKVTimeQueryMode for ergonomic use
type AKVTimeQueryMode = proto.AKVTimeQueryMode

const (
	TimeQueryLatestBefore  = proto.AKVTimeQueryMode_AKV_TIME_QUERY_LATEST_BEFORE
	TimeQueryEarliestAfter = proto.AKVTimeQueryMode_AKV_TIME_QUERY_EARLIEST_AFTER
	TimeQueryRange         = proto.AKVTimeQueryMode_AKV_TIME_QUERY_RANGE
	TimeQueryLatest        = proto.AKVTimeQueryMode_AKV_TIME_QUERY_LATEST
)

// AKVTimeGroupInterval mirrors proto.AKVTimeGroupInterval
type AKVTimeGroupInterval = proto.AKVTimeGroupInterval

const (
	TimeGroupNone     = proto.AKVTimeGroupInterval_AKV_TIME_GROUP_NONE
	TimeGroupMinutely = proto.AKVTimeGroupInterval_AKV_TIME_GROUP_MINUTELY
	TimeGroupHourly   = proto.AKVTimeGroupInterval_AKV_TIME_GROUP_HOURLY
	TimeGroupDaily    = proto.AKVTimeGroupInterval_AKV_TIME_GROUP_DAILY
	TimeGroupWeekly   = proto.AKVTimeGroupInterval_AKV_TIME_GROUP_WEEKLY
	TimeGroupMonthly  = proto.AKVTimeGroupInterval_AKV_TIME_GROUP_MONTHLY
)

// AKVTimeGroupPick mirrors proto.AKVTimeGroupPick
type AKVTimeGroupPick = proto.AKVTimeGroupPick

const (
	TimePickFirst = proto.AKVTimeGroupPick_AKV_TIME_PICK_FIRST
	TimePickLast  = proto.AKVTimeGroupPick_AKV_TIME_PICK_LAST
)

// AKVTimeQuery configures a time-series query
type AKVTimeQuery struct {
	Key       string
	Workspace string // empty = global
	Mode      AKVTimeQueryMode

	// Used by LatestBefore / EarliestAfter
	Pivot time.Time

	// Used by Range
	RangeStart time.Time
	RangeEnd   time.Time

	// Grouping (Range mode only)
	GroupInterval AKVTimeGroupInterval
	GroupPick     AKVTimeGroupPick

	// Max results (Range mode only, max 1000)
	Limit int32
}

// AKVTimeValue is a single result from a time-series query
type AKVTimeValue struct {
	Timestamp time.Time
	Data      []byte
}

// AKVTimeGet queries the timeline store and returns results.
func (a *Actor) AKVTimeGet(ctx context.Context, query AKVTimeQuery) ([]AKVTimeValue, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	req := &proto.AKVTimeGetRequest{
		Authorization: a.Authorization(),
		Key:           query.Key,
		Workspace:     query.Workspace,
		Mode:          query.Mode,
		GroupInterval: query.GroupInterval,
		GroupPick:     query.GroupPick,
		Limit:         query.Limit,
	}

	if !query.Pivot.IsZero() {
		req.PivotTimestamp = query.Pivot.Unix()
	}
	if !query.RangeStart.IsZero() {
		req.RangeStart = query.RangeStart.Unix()
	}
	if !query.RangeEnd.IsZero() {
		req.RangeEnd = query.RangeEnd.Unix()
	}

	resp, err := a.connection.AKVTimeGet(ctx, req)
	if err != nil {
		return nil, err
	}

	if !resp.GetSummary().GetSuccess() {
		return nil, &Error{
			ErrorCode:    resp.GetSummary().GetErrorCode(),
			ErrorMessage: resp.GetSummary().GetErrorMessage(),
		}
	}

	results := make([]AKVTimeValue, 0, len(resp.GetResults()))
	for _, r := range resp.GetResults() {
		results = append(results, AKVTimeValue{
			Timestamp: time.Unix(r.GetTimestampUnix(), 0),
			Data:      r.GetData(),
		})
	}

	return results, nil
}

// AKVTimeGetLatest returns the most recent value for a key.
// Returns nil, nil if no value exists.
func (a *Actor) AKVTimeGetLatest(ctx context.Context, key, workspace string) (*AKVTimeValue, error) {
	results, err := a.AKVTimeGet(ctx, AKVTimeQuery{
		Key:       key,
		Workspace: workspace,
		Mode:      TimeQueryLatest,
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// AKVTimeGetAt returns the most recent value at or before the given timestamp.
// This is the primary method for historical lookups (e.g., "what was the exchange rate at this time?").
// Returns nil, nil if no value exists at or before the timestamp.
func (a *Actor) AKVTimeGetAt(ctx context.Context, key, workspace string, at time.Time) (*AKVTimeValue, error) {
	results, err := a.AKVTimeGet(ctx, AKVTimeQuery{
		Key:       key,
		Workspace: workspace,
		Mode:      TimeQueryLatestBefore,
		Pivot:     at,
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// AKVTimeDeleteMode mirrors proto.AKVTimeDeleteMode
type AKVTimeDeleteMode = proto.AKVTimeDeleteMode

const (
	TimeDeleteAll       = proto.AKVTimeDeleteMode_AKV_TIME_DELETE_ALL
	TimeDeleteExact     = proto.AKVTimeDeleteMode_AKV_TIME_DELETE_EXACT
	TimeDeleteTimeRange = proto.AKVTimeDeleteMode_AKV_TIME_DELETE_TIME_RANGE
)

// AKVTimeDel deletes timeline entries for a key.
//   - TimeDeleteAll: deletes all entries for the key
//   - TimeDeleteExact: deletes specific timestamps (pass timestamps parameter)
//   - TimeDeleteTimeRange: deletes entries between rangeStart and rangeEnd (inclusive)
func (a *Actor) AKVTimeDel(ctx context.Context, key, workspace string, mode AKVTimeDeleteMode, opts ...AKVTimeDelOption) (*proto.GenericResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}

	req := &proto.AKVTimeDelRequest{
		Authorization: a.Authorization(),
		Key:           key,
		Workspace:     workspace,
		Mode:          mode,
	}

	for _, opt := range opts {
		opt(req)
	}

	return a.connection.AKVTimeDel(ctx, req)
}

// AKVTimeDelOption configures a delete request
type AKVTimeDelOption func(*proto.AKVTimeDelRequest)

// WithTimestamps sets the timestamps for EXACT delete mode
func WithTimestamps(timestamps ...time.Time) AKVTimeDelOption {
	return func(req *proto.AKVTimeDelRequest) {
		for _, t := range timestamps {
			req.Timestamps = append(req.Timestamps, t.Unix())
		}
	}
}

// WithTimeRange sets the range for TIME_RANGE delete mode
func WithTimeRange(start, end time.Time) AKVTimeDelOption {
	return func(req *proto.AKVTimeDelRequest) {
		req.RangeStart = start.Unix()
		req.RangeEnd = end.Unix()
	}
}
