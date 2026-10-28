package keystone

import (
	"context"
	"testing"
	"time"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	gproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestFeedSummaryCalendarAndJump(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	feed := actor.Feed("channel-messages", "channel-123")
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	until := from.AddDate(0, 1, 0)
	mock.FeedSummaryFunc = func(_ context.Context, req *proto.FeedSummaryRequest) (*proto.FeedSummaryResponse, error) {
		if !gproto.Equal(req.Authorization, actor.Authorization()) || req.Feed.Type != "channel-messages" || req.Feed.Key != "channel-123" || !req.From.AsTime().Equal(from) || !req.Until.AsTime().Equal(until) || req.Interval != proto.FeedSummaryInterval_FEED_SUMMARY_DAILY || req.Timezone != "Europe/London" || !req.Ascending {
			t.Errorf("lost summary options: %v", req)
		}
		return &proto.FeedSummaryResponse{PendingHours: 2, Approximate: true, Buckets: []*proto.FeedSummaryBucket{{Start: timestamppb.New(from), End: timestamppb.New(from.Add(24 * time.Hour)), Count: 12, FirstPosition: "first", LastPosition: "last", Cursor: "jump", UpdatedAt: timestamppb.New(until)}}}, nil
	}
	summary, err := feed.Summary(context.Background(), FeedSummaryOptions{From: from, Until: until, Timezone: "Europe/London", Ascending: true})
	if err != nil {
		t.Fatal(err)
	}
	if !summary.Approximate || summary.PendingHours != 2 || len(summary.Buckets) != 1 || summary.Buckets[0].Count != 12 || !summary.Buckets[0].UpdatedAt.Equal(until) || summary.Buckets[0].FirstPosition != "first" || summary.Buckets[0].LastPosition != "last" {
		t.Fatalf("lost summary fields: %v", summary)
	}
	mock.FeedReadFunc = func(_ context.Context, req *proto.FeedReadRequest) (*proto.FeedReadResponse, error) {
		if req.Cursor != "jump" || !req.Ascending || req.Position != "" {
			t.Errorf("jump request: %v", req)
		}
		return &proto.FeedReadResponse{Entries: []*proto.FeedEntry{{EntityId: "message"}}}, nil
	}
	page, err := feed.Read(context.Background(), FeedReadOptions{Cursor: summary.Buckets[0].Cursor, Ascending: true})
	if err != nil || len(page.Entries()) != 1 {
		t.Fatalf("jump: %v %v", page, err)
	}
	mock.FeedSummaryFunc = func(context.Context, *proto.FeedSummaryRequest) (*proto.FeedSummaryResponse, error) {
		return nil, status.Error(codes.Unavailable, "retry")
	}
	if _, err := feed.Summary(context.Background(), FeedSummaryOptions{From: from, Until: until}); status.Code(err) != codes.Unavailable {
		t.Fatalf("lost RPC error: %v", err)
	}
}

func TestFeedSummaryInvalidOptions(t *testing.T) {
	actor, _, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	feed := actor.Feed("activity", "user")
	from := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for _, options := range []FeedSummaryOptions{
		{}, {From: from, Until: from}, {From: from.Add(time.Second), Until: from.Add(time.Hour)},
		{From: from, Until: from.Add(time.Hour), Interval: FeedSummaryInterval(99)},
		{From: from, Until: from.Add(32 * 24 * time.Hour), Interval: FeedHourly},
		{From: from, Until: from.Add(367 * 24 * time.Hour)},
	} {
		if _, err := feed.Summary(context.Background(), options); err == nil {
			t.Fatalf("invalid summary options accepted: %v", options)
		}
	}
	var empty *Feed
	if _, err := empty.Summary(context.Background(), FeedSummaryOptions{From: from, Until: from.Add(time.Hour)}); err == nil {
		t.Fatal("nil feed accepted")
	}
}
