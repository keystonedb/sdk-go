package keystone

import (
	"context"
	"fmt"
	"time"

	"github.com/keystonedb/sdk-go/proto"
	gproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FeedSummaryInterval int32

const (
	FeedDaily FeedSummaryInterval = iota
	FeedHourly
	FeedMonthly
)

// FeedSummaryOptions selects a calendar outline. From is inclusive, Until is
// exclusive; both must align to a minute. Interval defaults to daily, Timezone
// to UTC. Ascending controls the direction of the returned jump cursors.
type FeedSummaryOptions struct {
	From, Until time.Time
	Interval    FeedSummaryInterval
	Timezone    string
	Ascending   bool
}

type FeedSummaryBucket struct {
	Start, End time.Time
	Count      int64
	// Deprecated: exact member boundaries are not populated by activity outlines.
	// Use Cursor to navigate to the calendar boundary.
	FirstPosition, LastPosition string
	Cursor                      string
	// Deprecated: activity outlines have no exact projection refresh timestamp.
	UpdatedAt time.Time
}

// FeedSummary is a conservative activity outline, ordered chronologically.
// Approximate outlines may retain empty periods and overcount retries/removals.
// Recorded hours stay visible even if their density counter is zero or missing.
type FeedSummary struct {
	Buckets []FeedSummaryBucket
	// Deprecated: current activity outlines are immediately available. There is
	// no projection to await, and this field is zero on the current server.
	PendingHours int32
	Approximate  bool
}

// Summary loads approximate density and jump cursors without loading feed entries.
// Recorded hours stay visible independently of density counters. To jump,
// pass a bucket's Cursor to Read with the same Ascending setting. A jump includes
// the time boundary and then continues through the feed, beyond that period.
func (f *Feed) Summary(ctx context.Context, options FeedSummaryOptions) (*FeedSummary, error) {
	auth, err := f.authorization()
	if err != nil {
		return nil, err
	}
	if options.From.IsZero() || options.Until.IsZero() || !options.From.Before(options.Until) ||
		!options.From.Equal(options.From.Truncate(time.Minute)) || !options.Until.Equal(options.Until.Truncate(time.Minute)) {
		return nil, fmt.Errorf("feed: From must precede Until; both must be set and aligned to a minute")
	}
	maxRange := 366 * 24 * time.Hour
	switch options.Interval {
	case FeedDaily, FeedMonthly:
	case FeedHourly:
		maxRange = 31 * 24 * time.Hour
	default:
		return nil, fmt.Errorf("feed: unsupported summary interval")
	}
	if options.Until.Sub(options.From) > maxRange {
		return nil, fmt.Errorf("feed: summary range exceeds 31 days for hourly or 366 days for daily/monthly intervals")
	}
	from, until := timestamppb.New(options.From), timestamppb.New(options.Until)
	if from.CheckValid() != nil || until.CheckValid() != nil {
		return nil, fmt.Errorf("feed: summary timestamps are outside the supported range")
	}
	response, err := f.actor.connection.FeedSummary(ctx, &proto.FeedSummaryRequest{
		Authorization: auth, Feed: gproto.Clone(&f.key).(*proto.FeedKey), From: from, Until: until,
		Interval: proto.FeedSummaryInterval(options.Interval), Timezone: options.Timezone, Ascending: options.Ascending,
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("feed: nil summary response")
	}
	summary := &FeedSummary{PendingHours: response.PendingHours, Approximate: response.Approximate, Buckets: make([]FeedSummaryBucket, 0, len(response.Buckets))}
	for _, bucket := range response.Buckets {
		if bucket == nil || bucket.Start == nil || bucket.End == nil || bucket.Start.CheckValid() != nil || bucket.End.CheckValid() != nil {
			return nil, fmt.Errorf("feed: invalid calendar bucket in summary response")
		}
		item := FeedSummaryBucket{Start: bucket.Start.AsTime(), End: bucket.End.AsTime(), Count: bucket.Count,
			FirstPosition: bucket.FirstPosition, LastPosition: bucket.LastPosition, Cursor: bucket.Cursor}
		if bucket.UpdatedAt != nil {
			item.UpdatedAt = bucket.UpdatedAt.AsTime()
		}
		summary.Buckets = append(summary.Buckets, item)
	}
	return summary, nil
}
