package keystone

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/kubex/k4id"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	gproto "google.golang.org/protobuf/proto"
)

func TestFeedPreservesPublicK4Key(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	key := k4id.New().String()
	feed := actor.Feed("channel-messages", key)
	check := func(got *proto.FeedKey, auth *proto.Authorization) {
		if got.GetType() != "channel-messages" || got.GetKey() != key || !gproto.Equal(auth, actor.Authorization()) {
			t.Errorf("public key or actor scope changed: %v %v", got, auth)
		}
	}
	mock.FeedPutFunc = func(_ context.Context, req *proto.FeedPutRequest) (*proto.FeedPutResponse, error) {
		check(req.Feed, req.Authorization)
		if req.Entry.Position == key {
			t.Error("feed identity replaced the independent ordering position")
		}
		return &proto.FeedPutResponse{Entry: req.Entry}, nil
	}
	entry := &FeedEntry{Data: []byte("message")}
	if err := feed.Put(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	mock.FeedReadFunc = func(_ context.Context, req *proto.FeedReadRequest) (*proto.FeedReadResponse, error) {
		check(req.Feed, req.Authorization)
		return &proto.FeedReadResponse{Entries: []*proto.FeedEntry{{Position: entry.Position, Data: entry.Data}}}, nil
	}
	page, err := feed.Read(context.Background(), FeedReadOptions{})
	if err != nil || len(page.Entries()) != 1 || page.Entries()[0].Position != entry.Position {
		t.Fatalf("read position was not preserved: %v %v", page, err)
	}
	mock.FeedRemoveFunc = func(_ context.Context, req *proto.FeedRemoveRequest) (*proto.GenericResponse, error) {
		check(req.Feed, req.Authorization)
		if req.Position != entry.Position {
			t.Error("removal lost the membership position")
		}
		return &proto.GenericResponse{Success: true}, nil
	}
	if err := feed.Remove(context.Background(), entry.Position); err != nil {
		t.Fatal(err)
	}
}

func TestFeedPutPreservesPositionForRetry(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	feed := actor.Feed("channel-messages", "channel-123")
	entry := &FeedEntry{EntityID: "message-123", EntityType: "messages", Data: []byte("preview")}
	var first *proto.FeedPutRequest
	calls := 0
	mock.FeedPutFunc = func(_ context.Context, req *proto.FeedPutRequest) (*proto.FeedPutResponse, error) {
		calls++
		if !gproto.Equal(req.Authorization, actor.Authorization()) || req.Feed.Type != "channel-messages" || req.Feed.Key != "channel-123" {
			t.Errorf("lost actor or feed: %v", req)
		}
		id, err := uuid.Parse(req.Entry.Position)
		if err != nil || id.Version() != 1 || id.Variant() != uuid.RFC4122 {
			t.Errorf("invalid generated position: %q", req.Entry.Position)
		}
		if req.Entry.EntityId != "message-123" || req.Entry.GetSchema().GetKey() != "messages" || !gproto.Equal(req.Entry.GetSchema().GetSource(), actor.VendorApp()) || string(req.Entry.Data) != "preview" {
			t.Errorf("lost entry: %v", req.Entry)
		}
		if calls == 1 {
			first = gproto.Clone(req).(*proto.FeedPutRequest)
			return nil, status.Error(codes.Unavailable, "retry")
		}
		if !gproto.Equal(req, first) {
			t.Errorf("retry changed request: got %v want %v", req, first)
		}
		return &proto.FeedPutResponse{Entry: req.Entry}, nil
	}
	if err := feed.Put(context.Background(), entry); status.Code(err) != codes.Unavailable {
		t.Fatalf("lost RPC error: %v", err)
	}
	if entry.Position == "" {
		t.Fatal("position not retained after failure")
	}
	position := entry.Position
	if err := feed.Put(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	if entry.Position != position || calls != 2 {
		t.Fatal("retry allocated another position")
	}
}

func TestFeedReadPaginationAndRetry(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	feed := actor.Feed("message-replies", "message-123")
	anchor, _ := uuid.NewUUID()
	options := FeedReadOptions{Limit: 10, Ascending: true, Position: anchor.String()}
	wantAuth := gproto.Clone(actor.Authorization())
	calls := 0
	mock.FeedReadFunc = func(_ context.Context, req *proto.FeedReadRequest) (*proto.FeedReadResponse, error) {
		calls++
		if !gproto.Equal(req.Authorization, wantAuth) || req.Feed.Type != "message-replies" || req.Feed.Key != "message-123" || !req.Ascending || req.Limit != 10 {
			t.Errorf("lost read settings: %v", req)
		}
		if calls == 1 {
			if req.Position != anchor.String() || req.Cursor != "" {
				t.Errorf("lost anchor: %v", req)
			}
			return &proto.FeedReadResponse{Entries: []*proto.FeedEntry{{Position: "entry-1", EntityId: "reply-1", Schema: &proto.Key{Key: "messages"}, Data: []byte("one")}}, HasMore: true, NextCursor: "next-1"}, nil
		}
		if req.Position != "" || req.Cursor != "next-1" {
			t.Errorf("anchor was not cleared on continuation: %v", req)
		}
		if calls == 2 {
			return nil, status.Error(codes.Unavailable, "retry")
		}
		return &proto.FeedReadResponse{Entries: []*proto.FeedEntry{{Position: "entry-2", EntityId: "reply-2"}}}, nil
	}
	page, err := feed.Read(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Entries()) != 1 || page.Entries()[0].EntityID != "reply-1" || page.Entries()[0].EntityType != "messages" || string(page.Entries()[0].Data) != "one" {
		t.Fatalf("entry decode: %v", page.Entries())
	}
	// A continuation must retain the original identity and feed even if the
	// caller changes the actor or reuses its options variable.
	actor.SetTraceID("changed")
	actor.user.UserId = "other-user"
	options.Limit = 99
	if err := page.Next(context.Background()); status.Code(err) != codes.Unavailable {
		t.Fatalf("lost next error: %v", err)
	}
	if len(page.Entries()) != 1 || page.Entries()[0].EntityID != "reply-1" || page.NextCursor() != "next-1" {
		t.Fatal("failure discarded the current page")
	}
	if err := page.Next(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(page.Entries()) != 1 || page.Entries()[0].EntityID != "reply-2" || page.HasMore() {
		t.Fatal("next did not replace current page")
	}
	if err := page.Next(context.Background()); !errors.Is(err, io.EOF) {
		t.Fatalf("end: %v", err)
	}
	if calls != 3 {
		t.Fatalf("unexpected RPCs: %d", calls)
	}
}

func TestFeedEmptyPagesAndResume(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	feed := actor.Feed("user-activity", "user-1")
	calls := 0
	mock.FeedReadFunc = func(_ context.Context, req *proto.FeedReadRequest) (*proto.FeedReadResponse, error) {
		calls++
		if req.Ascending || req.Limit != 0 || req.Position != "" {
			t.Errorf("unexpected defaults: %v", req)
		}
		if calls == 1 {
			if req.Cursor != "saved" {
				t.Errorf("lost resume cursor: %v", req)
			}
			return &proto.FeedReadResponse{HasMore: true, NextCursor: "next"}, nil
		}
		if req.Cursor != "next" {
			t.Errorf("lost empty-page cursor: %v", req)
		}
		return &proto.FeedReadResponse{Entries: []*proto.FeedEntry{{Data: []byte("activity")}}}, nil
	}
	page, err := feed.Read(context.Background(), FeedReadOptions{Cursor: "saved"})
	if err != nil || len(page.Entries()) != 0 || !page.HasMore() {
		t.Fatalf("empty page: %v %v", page, err)
	}
	if err := page.Next(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(page.Entries()) != 1 || string(page.Entries()[0].Data) != "activity" || page.HasMore() {
		t.Fatalf("lost payload-only entry: %v", page.Entries())
	}
}

func TestFeedRemoveAndValidation(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	feed := actor.Feed("channel-messages", "channel-123")
	position, _ := uuid.NewUUID()
	mock.FeedRemoveFunc = func(_ context.Context, req *proto.FeedRemoveRequest) (*proto.GenericResponse, error) {
		if req.Position != position.String() || req.Feed.Key != "channel-123" || !gproto.Equal(req.Authorization, actor.Authorization()) {
			t.Errorf("bad removal: %v", req)
		}
		return &proto.GenericResponse{Success: true}, nil
	}
	if err := feed.Remove(context.Background(), position.String()); err != nil {
		t.Fatal(err)
	}
	mock.FeedRemoveFunc = func(context.Context, *proto.FeedRemoveRequest) (*proto.GenericResponse, error) {
		return &proto.GenericResponse{ErrorMessage: "denied"}, nil
	}
	if err := feed.Remove(context.Background(), position.String()); err == nil || err.Error() != "feed: denied" {
		t.Fatalf("ignored removal failure: %v", err)
	}
	for _, options := range []FeedReadOptions{{Limit: -1}, {Limit: 1001}, {Position: "bad"}, {Position: position.String(), Cursor: "cursor"}} {
		if _, err := feed.Read(context.Background(), options); err == nil {
			t.Fatalf("invalid read options accepted: %v", options)
		}
	}
	for _, entry := range []*FeedEntry{nil, {}, {EntityID: "message", Position: "bad"}, {EntityID: "message", Position: uuid.NewString()}} {
		if err := feed.Put(context.Background(), entry); err == nil {
			t.Fatalf("invalid entry accepted: %v", entry)
		}
	}
	if err := feed.Remove(context.Background(), "bad"); err == nil {
		t.Fatal("invalid removal accepted")
	}
	var nilActor *Actor
	for _, f := range []*Feed{nil, nilActor.Feed("type", "key"), (&Actor{}).Feed("type", "key"), actor.Feed("", "key"), actor.Feed("type", "")} {
		if _, err := f.Read(context.Background(), FeedReadOptions{}); err == nil {
			t.Fatal("invalid feed accepted")
		}
		if err := f.Put(context.Background(), &FeedEntry{Data: []byte("x")}); err == nil {
			t.Fatal("invalid feed write accepted")
		}
		if err := f.Remove(context.Background(), position.String()); err == nil {
			t.Fatal("invalid feed removal accepted")
		}
	}
}

func TestFeedMalformedContinuationRetainsPage(t *testing.T) {
	actor, mock, cleanup := newQueryIndexTestActor(t)
	defer cleanup()
	feed := actor.Feed("activity", "user-1")
	mock.FeedReadFunc = func(context.Context, *proto.FeedReadRequest) (*proto.FeedReadResponse, error) {
		return &proto.FeedReadResponse{HasMore: true}, nil
	}
	page, err := feed.Read(context.Background(), FeedReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Next(context.Background()); err == nil {
		t.Fatal("missing cursor accepted")
	}
}
