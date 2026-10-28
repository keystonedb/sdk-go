package keystone

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/proto"
	gproto "google.golang.org/protobuf/proto"
)

// FeedEntry is an entity reference, a small opaque payload, or both. Position is
// a UUID v1 ordering key, separate from EntityID. Put fills it in when empty.
type FeedEntry struct {
	Position   string
	EntityID   ID
	EntityType string
	Data       []byte
}

// FeedReadOptions controls a page. Zero values read the newest 50 entries.
// Position is an exclusive anchor: before it by default, after it when Ascending.
// Cursor resumes a previous page and cannot be combined with Position.
type FeedReadOptions struct {
	Limit     int32
	Ascending bool
	Position  string
	Cursor    string
}

// Feed is a named ordered collection scoped to an actor's app and workspace.
// It uses the feed RPCs directly, without SQL or schema registration.
type Feed struct {
	actor *Actor
	key   proto.FeedKey
}

// Feed selects a type/key collection without making a network request.
// Keys may be entity IDs, k4id strings or ordinary names. The server derives
// its compact storage identity from the actor scope and these original values.
func (a *Actor) Feed(feedType, key string) *Feed {
	return &Feed{actor: a, key: proto.FeedKey{Type: feedType, Key: key}}
}

func (f *Feed) authorization() (*proto.Authorization, error) {
	if f == nil || f.actor == nil || f.actor.connection == nil {
		return nil, fmt.Errorf("feed: actor must have a connection")
	}
	if f.key.Type == "" || len(f.key.Type) > 128 || f.key.Key == "" || len(f.key.Key) > 512 {
		return nil, fmt.Errorf("feed: type (1–128 bytes) and key (1–512 bytes) are required")
	}
	return gproto.Clone(f.actor.Authorization()).(*proto.Authorization), nil
}

// Put adds or updates an entry. It assigns Position before the RPC, preserving it
// even on failure: retry Put with the same entry to avoid duplicate membership.
// Put does not create or update the referenced entity. Use entries sequentially;
// concurrent calls sharing the same entry require synchronization.
func (f *Feed) Put(ctx context.Context, entry *FeedEntry) error {
	auth, err := f.authorization()
	if err != nil {
		return err
	}
	if entry == nil || (entry.EntityID == "" && len(entry.Data) == 0) {
		return fmt.Errorf("feed: entry requires EntityID or Data")
	}
	if entry.Position == "" {
		position, err := uuid.NewUUID()
		if err != nil {
			return fmt.Errorf("feed: create position: %w", err)
		}
		entry.Position = position.String()
	} else if err := validateFeedPosition(entry.Position); err != nil {
		return err
	}
	raw := &proto.FeedEntry{Position: entry.Position, EntityId: string(entry.EntityID), Data: append([]byte(nil), entry.Data...)}
	if entry.EntityType != "" {
		raw.Schema = &proto.Key{Key: entry.EntityType, Source: auth.Source}
	}
	response, err := f.actor.connection.FeedPut(ctx, &proto.FeedPutRequest{
		Authorization: auth, Feed: gproto.Clone(&f.key).(*proto.FeedKey), Entry: raw,
	})
	if err != nil {
		return err
	}
	if response.GetEntry().GetPosition() == "" {
		return fmt.Errorf("feed: put response has no entry position")
	}
	return nil
}

// Remove deletes a membership by position, without deleting its referenced entity.
func (f *Feed) Remove(ctx context.Context, position string) error {
	auth, err := f.authorization()
	if err != nil {
		return err
	}
	if err := validateFeedPosition(position); err != nil {
		return err
	}
	response, err := f.actor.connection.FeedRemove(ctx, &proto.FeedRemoveRequest{
		Authorization: auth, Feed: gproto.Clone(&f.key).(*proto.FeedKey), Position: position,
	})
	if err != nil {
		return err
	}
	if response == nil {
		return fmt.Errorf("feed: nil remove response")
	}
	if !response.Success {
		if response.ErrorMessage != "" {
			return fmt.Errorf("feed: %s", response.ErrorMessage)
		}
		return fmt.Errorf("feed: remove failed")
	}
	return nil
}

func validateFeedPosition(position string) error {
	id, err := uuid.Parse(position)
	if err != nil || len(position) != 36 || id.Version() != 1 || id.Variant() != uuid.RFC4122 {
		return fmt.Errorf("feed: position must be an RFC 4122 UUID v1")
	}
	return nil
}

// FeedPage retains a single page. Next replaces its entries, keeping memory
// bounded. Use sequentially and use HasMore, not entry count, to detect the end:
// sparse feeds can return short or empty pages with a continuation cursor.
type FeedPage struct {
	connection *Connection
	request    *proto.FeedReadRequest
	response   *proto.FeedReadResponse
	entries    []FeedEntry
}

// Read fetches the first page with the supplied options. Entity references are
// returned as references; use Actor.GetByID to load the entity's properties.
func (f *Feed) Read(ctx context.Context, options FeedReadOptions) (*FeedPage, error) {
	auth, err := f.authorization()
	if err != nil {
		return nil, err
	}
	if options.Limit < 0 || options.Limit > 1000 {
		return nil, fmt.Errorf("feed: limit must be 0–1000")
	}
	if options.Cursor != "" && options.Position != "" {
		return nil, fmt.Errorf("feed: position and cursor cannot be combined")
	}
	if options.Position != "" {
		if err := validateFeedPosition(options.Position); err != nil {
			return nil, err
		}
	}
	req := &proto.FeedReadRequest{Authorization: auth, Feed: gproto.Clone(&f.key).(*proto.FeedKey), Limit: options.Limit,
		Ascending: options.Ascending, Position: options.Position, Cursor: options.Cursor}
	page := &FeedPage{connection: f.actor.connection}
	if err := page.read(ctx, req); err != nil {
		return nil, err
	}
	return page, nil
}

// Entries returns the current page's entries. Next replaces the slice.
func (p *FeedPage) Entries() []FeedEntry { return p.entries }
func (p *FeedPage) HasMore() bool        { return p.response.GetHasMore() }
func (p *FeedPage) NextCursor() string   { return p.response.GetNextCursor() }

// Next fetches the next page, returning io.EOF at the end. A failure preserves
// the current page and cursor for retry. The original authorization is retained.
func (p *FeedPage) Next(ctx context.Context) error {
	if !p.HasMore() {
		return io.EOF
	}
	if p.NextCursor() == "" {
		return fmt.Errorf("feed: response has more entries but no next cursor")
	}
	req := gproto.Clone(p.request).(*proto.FeedReadRequest)
	req.Position = ""
	req.Cursor = p.NextCursor()
	return p.read(ctx, req)
}

func (p *FeedPage) read(ctx context.Context, req *proto.FeedReadRequest) error {
	response, err := p.connection.FeedRead(ctx, gproto.Clone(req).(*proto.FeedReadRequest))
	if err != nil {
		return err
	}
	if response == nil {
		return fmt.Errorf("feed: nil read response")
	}
	entries := make([]FeedEntry, 0, len(response.Entries))
	for _, entry := range response.Entries {
		if entry == nil {
			return fmt.Errorf("feed: nil entry in read response")
		}
		entries = append(entries, FeedEntry{Position: entry.Position, EntityID: ID(entry.EntityId), EntityType: entry.GetSchema().GetKey(), Data: append([]byte(nil), entry.Data...)})
	}
	p.request, p.response, p.entries = req, response, entries
	return nil
}
