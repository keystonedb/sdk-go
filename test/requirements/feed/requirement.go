package feed

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/kubex/k4id"
)

type Requirement struct{}

var _ requirements.Requirement = (*Requirement)(nil)

func (*Requirement) Name() string { return "Ordered Feeds" }

func (*Requirement) Register(conn *keystone.Connection) error {
	// Only the entity-reference scenario needs a schema. Payload feeds do not.
	conn.RegisterTypes(models.User{})
	return nil
}

func (*Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	for _, scenario := range []struct {
		name string
		run  func(context.Context, *fixture) error
	}{
		{"Empty feed and end of pagination", empty},
		{"Payloads, generated positions, retry and update", put},
		{"UUID timestamp ordering and exclusive anchors", ordering},
		{"Default page size, both directions and saved cursors", pagination},
		{"Sparse history and removed entries", sparse},
		{"Entity references and membership-only removal", references},
		{"Feed type, key, workspace, app and vendor isolation", isolation},
		{"Input validation and size boundaries", validation},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		f := newFixture(actor)
		err := scenario.run(ctx, f)
		cancel()
		report(requirements.NewResult(scenario.name, err))
		if err := f.cleanup(); err != nil {
			report(requirements.NewResult(scenario.name+" cleanup", err))
		}
	}
}

type membership struct {
	feed     *keystone.Feed
	position string
}

type fixture struct {
	actor *keystone.Actor
	key   string
	rows  []membership
}

func newFixture(actor *keystone.Actor) *fixture {
	return &fixture{actor: actor, key: k4id.New().String()}
}

func (f *fixture) feed() *keystone.Feed { return f.actor.Feed("requirement-activity", f.key) }

func (f *fixture) put(ctx context.Context, feed *keystone.Feed, entry *keystone.FeedEntry) error {
	err := feed.Put(ctx, entry)
	// Put retains positions even after an ambiguous RPC failure.
	if entry.Position != "" {
		f.rows = append(f.rows, membership{feed, entry.Position})
	}
	return err
}

func (f *fixture) remove(ctx context.Context, feed *keystone.Feed, position string) error {
	if err := feed.Remove(ctx, position); err != nil {
		return err
	}
	f.rows = slices.DeleteFunc(f.rows, func(row membership) bool {
		return row.feed == feed && row.position == position
	})
	return nil
}

func (f *fixture) cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var errs []error
	seen := make(map[membership]bool)
	for _, row := range f.rows {
		if !seen[row] {
			errs = append(errs, row.feed.Remove(ctx, row.position))
			seen[row] = true
		}
	}
	return errors.Join(errs...)
}

func empty(ctx context.Context, f *fixture) error {
	for _, ascending := range []bool{false, true} {
		if err := expect(ctx, f.feed(), keystone.FeedReadOptions{Ascending: ascending}, nil); err != nil {
			return err
		}
	}
	return nil
}

func put(ctx context.Context, f *fixture) error {
	feed := f.feed()
	entry := keystone.FeedEntry{Data: []byte{0, 1, 255, '\n', 0}}
	if err := f.put(ctx, feed, &entry); err != nil {
		return err
	}
	id, err := uuid.Parse(entry.Position)
	if err != nil || id.Version() != 1 || id.Variant() != uuid.RFC4122 {
		return fmt.Errorf("generated position %q is not an RFC 4122 UUID v1", entry.Position)
	}
	position := entry.Position
	if err := feed.Put(ctx, &entry); err != nil {
		return fmt.Errorf("retry identical entry: %w", err)
	}
	if entry.Position != position {
		return fmt.Errorf("retry changed position")
	}
	if err := expect(ctx, feed, keystone.FeedReadOptions{}, []keystone.FeedEntry{entry}); err != nil {
		return fmt.Errorf("retry must not duplicate membership: %w", err)
	}
	entry.Data = []byte("updated payload")
	if err := feed.Put(ctx, &entry); err != nil {
		return err
	}
	if err := expect(ctx, feed, keystone.FeedReadOptions{}, []keystone.FeedEntry{entry}); err != nil {
		return fmt.Errorf("update at same position: %w", err)
	}
	if err := f.remove(ctx, feed, position); err != nil {
		return err
	}
	return expect(ctx, feed, keystone.FeedReadOptions{}, nil)
}

// position constructs a v1 UUID at an explicit timestamp. Spacing entries over
// minutes crosses time_low rollovers, catching accidental lexical UUID ordering.
func position(at time.Time) string {
	ticks := uint64(at.UnixNano()/100) + 0x01b21dd213814000
	var id uuid.UUID
	binary.BigEndian.PutUint32(id[0:4], uint32(ticks))
	binary.BigEndian.PutUint16(id[4:6], uint16(ticks>>32))
	binary.BigEndian.PutUint16(id[6:8], uint16(ticks>>48)&0x0fff|0x1000)
	id[8], id[9], id[15] = 0x80, 1, 1
	return id.String()
}

func seed(ctx context.Context, f *fixture, feed *keystone.Feed, count int, spacing time.Duration) ([]keystone.FeedEntry, error) {
	entries := make([]keystone.FeedEntry, count)
	start := time.Now().UTC().Add(-time.Duration(count) * spacing)
	for i := range entries {
		entries[i] = keystone.FeedEntry{Position: position(start.Add(time.Duration(i) * spacing)), Data: []byte(fmt.Sprintf("entry-%03d", i))}
	}
	// Commit newest first so commit order cannot stand in for timestamp order.
	for i := len(entries) - 1; i >= 0; i-- {
		if err := f.put(ctx, feed, &entries[i]); err != nil {
			return nil, fmt.Errorf("seed entry %d: %w", i, err)
		}
	}
	return entries, nil
}

func ordering(ctx context.Context, f *fixture) error {
	feed := f.feed()
	entries, err := seed(ctx, f, feed, 7, 3*time.Minute)
	if err != nil {
		return err
	}
	for _, ascending := range []bool{false, true} {
		want := slices.Clone(entries)
		if !ascending {
			slices.Reverse(want)
		}
		if err := expect(ctx, feed, keystone.FeedReadOptions{Ascending: ascending, Limit: 2}, want); err != nil {
			return err
		}
		for _, anchor := range []int{0, 3, 6} {
			want := slices.Clone(entries[:anchor])
			if ascending {
				want = slices.Clone(entries[anchor+1:])
			} else {
				slices.Reverse(want)
			}
			options := keystone.FeedReadOptions{Ascending: ascending, Position: entries[anchor].Position, Limit: 2}
			if err := expect(ctx, feed, options, want); err != nil {
				return fmt.Errorf("anchor %d, ascending %t: %w", anchor, ascending, err)
			}
		}
	}
	return nil
}

func pagination(ctx context.Context, f *fixture) error {
	feed := f.feed()
	entries, err := seed(ctx, f, feed, 55, time.Second)
	if err != nil {
		return err
	}
	page, err := feed.Read(ctx, keystone.FeedReadOptions{})
	if err != nil {
		return err
	}
	if len(page.Entries()) != 50 || !page.HasMore() || page.NextCursor() == "" {
		return fmt.Errorf("default page: entries=%d, hasMore=%t, cursor=%q; want 50 and continuation", len(page.Entries()), page.HasMore(), page.NextCursor())
	}
	for _, ascending := range []bool{false, true} {
		want := slices.Clone(entries)
		if !ascending {
			slices.Reverse(want)
		}
		for _, limit := range []int32{1, 7, 1000} {
			options := keystone.FeedReadOptions{Ascending: ascending, Limit: limit}
			if err := expect(ctx, feed, options, want); err != nil {
				return fmt.Errorf("limit %d, ascending %t: %w", limit, ascending, err)
			}
		}
		options := keystone.FeedReadOptions{Ascending: ascending, Limit: 7}
		page, err := feed.Read(ctx, options)
		if err != nil {
			return err
		}
		if !page.HasMore() || page.NextCursor() == "" {
			return fmt.Errorf("missing saved cursor")
		}
		consumed := len(page.Entries())
		options.Cursor = page.NextCursor()
		// Use a new handle so resume does not depend on the previous pager.
		if err := expect(ctx, f.feed(), options, want[consumed:]); err != nil {
			return fmt.Errorf("resume saved cursor: %w", err)
		}
		wrongDirection := options
		wrongDirection.Ascending = !ascending
		if _, err := feed.Read(ctx, wrongDirection); err == nil {
			return fmt.Errorf("cursor accepted with a different direction")
		}
		otherWorkspace := f.otherWorkspaceActor()
		otherUser := f.actor.Connection().Actor(f.actor.WorkspaceID(), f.actor.RemoteIP(), f.key, f.actor.UserAgent())
		for i, other := range []*keystone.Feed{
			f.actor.Feed("requirement-other", f.key),
			f.actor.Feed("requirement-activity", f.key+"-other"),
			otherWorkspace.Feed("requirement-activity", f.key),
			otherUser.Feed("requirement-activity", f.key),
			f.sourceActor(f.actor.VendorID(), f.actor.AppID()+"-feed-other").Feed("requirement-activity", f.key),
			f.sourceActor(f.actor.VendorID()+"-feed-other", f.actor.AppID()).Feed("requirement-activity", f.key),
		} {
			if _, err := other.Read(ctx, options); err == nil {
				return fmt.Errorf("cursor accepted in different scope %d", i)
			}
		}
		if _, err := feed.Read(ctx, keystone.FeedReadOptions{Cursor: "invalid-cursor"}); err == nil {
			return fmt.Errorf("malformed cursor accepted")
		}
	}
	return nil
}

func sparse(ctx context.Context, f *fixture) error {
	feed := f.feed()
	entries, err := seed(ctx, f, feed, 5, 40*24*time.Hour)
	if err != nil {
		return err
	}
	if err := f.remove(ctx, feed, entries[2].Position); err != nil {
		return err
	}
	want := append(slices.Clone(entries[:2]), entries[3:]...)
	for _, ascending := range []bool{true, false} {
		if !ascending {
			slices.Reverse(want)
		}
		if err := expect(ctx, feed, keystone.FeedReadOptions{Limit: 1, Ascending: ascending}, want); err != nil {
			return fmt.Errorf("sparse history: %w", err)
		}
	}
	return nil
}

func references(ctx context.Context, f *fixture) error {
	user := &models.User{ExternalID: f.key, Validate: "feed entity must remain unchanged"}
	if err := f.actor.Mutate(ctx, user); err != nil {
		return err
	}
	feed := f.feed()
	entries := []keystone.FeedEntry{
		{EntityID: user.GetKeystoneID(), EntityType: keystone.Type(user)},
		{EntityID: user.GetKeystoneID(), EntityType: keystone.Type(user), Data: []byte("preview")},
	}
	for i := range entries {
		if err := f.put(ctx, feed, &entries[i]); err != nil {
			return err
		}
	}
	if entries[0].Position == entries[1].Position {
		return fmt.Errorf("separate memberships for the same entity share a position")
	}
	if err := expect(ctx, feed, keystone.FeedReadOptions{Ascending: true}, entries); err != nil {
		return err
	}
	if err := f.remove(ctx, feed, entries[0].Position); err != nil {
		return err
	}
	if err := expect(ctx, feed, keystone.FeedReadOptions{}, entries[1:]); err != nil {
		return err
	}
	loaded := &models.User{}
	if err := f.actor.GetByID(ctx, user.GetKeystoneID(), loaded, keystone.WithProperties("external_id", "validate")); err != nil {
		return fmt.Errorf("load entity after removing membership: %w", err)
	}
	if loaded.ExternalID != user.ExternalID || loaded.Validate != user.Validate {
		return fmt.Errorf("feed operations changed the referenced entity: %+v", loaded)
	}
	return nil
}

func isolation(ctx context.Context, f *fixture) error {
	other := f.otherWorkspaceActor()
	feeds := []*keystone.Feed{
		f.feed(),
		f.actor.Feed("requirement-other", f.key),
		f.actor.Feed("requirement-activity", f.key+"-other"),
		other.Feed("requirement-activity", f.key),
		f.sourceActor(f.actor.VendorID(), f.actor.AppID()+"-feed-other").Feed("requirement-activity", f.key),
		f.sourceActor(f.actor.VendorID()+"-feed-other", f.actor.AppID()).Feed("requirement-activity", f.key),
	}
	entries := make([]keystone.FeedEntry, len(feeds))
	sharedPosition := position(time.Now())
	for i, feed := range feeds {
		if err := expect(ctx, feed, keystone.FeedReadOptions{}, nil); err != nil {
			return err
		}
		entries[i] = keystone.FeedEntry{Position: sharedPosition, Data: []byte(fmt.Sprintf("scope-%d", i))}
		if err := f.put(ctx, feed, &entries[i]); err != nil {
			return err
		}
	}
	for i, feed := range feeds {
		if err := expect(ctx, feed, keystone.FeedReadOptions{}, entries[i:i+1]); err != nil {
			return fmt.Errorf("scope %d: %w", i, err)
		}
	}
	if err := f.remove(ctx, feeds[0], sharedPosition); err != nil {
		return err
	}
	if err := expect(ctx, feeds[0], keystone.FeedReadOptions{}, nil); err != nil {
		return err
	}
	for i := 1; i < len(feeds); i++ {
		if err := expect(ctx, feeds[i], keystone.FeedReadOptions{}, entries[i:i+1]); err != nil {
			return fmt.Errorf("removal leaked into scope %d: %w", i, err)
		}
	}
	return nil
}

func (f *fixture) otherWorkspaceActor() *keystone.Actor {
	// Workspace IDs must be 2–14 bytes, independently of the public feed key.
	for {
		workspaceID := "fw" + k4id.New().String()[:12]
		if workspaceID == f.actor.WorkspaceID() {
			continue
		}
		actor := f.actor.Connection().Actor(workspaceID, f.actor.RemoteIP(), f.actor.UserID(), f.actor.UserAgent())
		return &actor
	}
}

func (f *fixture) sourceActor(vendorID, appID string) *keystone.Actor {
	// Reuse the transport and credentials, with no schema registration in the
	// secondary source: payload feeds must work independently of entity schemas.
	conn := keystone.NewConnection(f.actor.Connection().DirectClient(), vendorID, appID, f.actor.Authorization().GetToken())
	actor := conn.Actor(f.actor.WorkspaceID(), f.actor.RemoteIP(), f.actor.UserID(), f.actor.UserAgent())
	return &actor
}

func validation(ctx context.Context, f *fixture) error {
	feed := f.feed()
	for _, invalid := range []*keystone.FeedEntry{nil, {}, {Data: []byte("x"), Position: "bad"}, {Data: []byte("x"), Position: uuid.NewString()}} {
		if err := feed.Put(ctx, invalid); err == nil {
			return fmt.Errorf("accepted invalid entry: %+v", invalid)
		}
	}
	anchor := position(time.Now())
	for _, options := range []keystone.FeedReadOptions{{Limit: -1}, {Limit: 1001}, {Position: "bad"}, {Position: uuid.NewString()}, {Position: anchor, Cursor: "cursor"}} {
		if _, err := feed.Read(ctx, options); err == nil {
			return fmt.Errorf("accepted invalid read options: %+v", options)
		}
	}
	for _, invalid := range []string{"", "bad", uuid.NewString()} {
		if err := feed.Remove(ctx, invalid); err == nil {
			return fmt.Errorf("accepted invalid removal position %q", invalid)
		}
	}
	for _, pair := range [][2]string{{"", f.key}, {"type", ""}, {strings.Repeat("t", 129), f.key}, {"type", strings.Repeat("k", 513)}, {strings.Repeat("é", 65), f.key}} {
		invalid := f.actor.Feed(pair[0], pair[1])
		if err := invalid.Put(ctx, &keystone.FeedEntry{Data: []byte("x")}); err == nil {
			return fmt.Errorf("accepted invalid feed in Put")
		}
		if _, err := invalid.Read(ctx, keystone.FeedReadOptions{}); err == nil {
			return fmt.Errorf("accepted invalid feed in Read")
		}
		if err := invalid.Remove(ctx, anchor); err == nil {
			return fmt.Errorf("accepted invalid feed in Remove")
		}
	}
	boundary := f.actor.Feed(strings.Repeat("t", 128), f.key+strings.Repeat("k", 512-len(f.key)))
	entry := keystone.FeedEntry{Data: bytes.Repeat([]byte{0xff}, 16*1024)}
	if err := f.put(ctx, boundary, &entry); err != nil {
		return fmt.Errorf("maximum type, key and payload sizes: %w", err)
	}
	if err := expect(ctx, boundary, keystone.FeedReadOptions{Limit: 1000}, []keystone.FeedEntry{entry}); err != nil {
		return err
	}
	oversized := keystone.FeedEntry{Data: make([]byte, 16*1024+1)}
	if err := feed.Put(ctx, &oversized); err == nil {
		f.rows = append(f.rows, membership{feed, oversized.Position})
		return fmt.Errorf("accepted payload larger than 16 KiB")
	}
	// A rejected write must leave the collection empty.
	return expect(ctx, feed, keystone.FeedReadOptions{}, nil)
}

// expect follows HasMore even for empty pages. It compares each page in place
// so an SDK that appends instead of replacing entries cannot pass unnoticed.
func expect(ctx context.Context, feed *keystone.Feed, options keystone.FeedReadOptions, want []keystone.FeedEntry) error {
	page, err := feed.Read(ctx, options)
	if err != nil {
		return err
	}
	limit := int(options.Limit)
	if limit == 0 {
		limit = 50
	}
	seenCursors := make(map[string]bool)
	index := 0
	for pages := 0; ; pages++ {
		if pages >= 1024 {
			return fmt.Errorf("pagination did not terminate within 1024 pages")
		}
		if len(page.Entries()) > limit {
			return fmt.Errorf("page contains %d entries, limit %d", len(page.Entries()), limit)
		}
		for _, got := range page.Entries() {
			if index >= len(want) {
				return fmt.Errorf("unexpected extra entry at %d: %q", index, got.Position)
			}
			expected := want[index]
			if got.Position != expected.Position || got.EntityID != expected.EntityID || got.EntityType != expected.EntityType || !bytes.Equal(got.Data, expected.Data) {
				return fmt.Errorf("entry %d: position=%q entity=%q type=%q payload bytes=%d; want position=%q entity=%q type=%q payload bytes=%d", index, got.Position, got.EntityID, got.EntityType, len(got.Data), expected.Position, expected.EntityID, expected.EntityType, len(expected.Data))
			}
			index++
		}
		if !page.HasMore() {
			break
		}
		cursor := page.NextCursor()
		if cursor == "" || seenCursors[cursor] {
			return fmt.Errorf("missing or repeated continuation cursor")
		}
		seenCursors[cursor] = true
		if err := page.Next(ctx); err != nil {
			return fmt.Errorf("next page: %w", err)
		}
	}
	if index != len(want) {
		return fmt.Errorf("read %d entries, want %d", index, len(want))
	}
	if err := page.Next(ctx); !errors.Is(err, io.EOF) {
		return fmt.Errorf("Next at end = %v, want io.EOF", err)
	}
	return nil
}
