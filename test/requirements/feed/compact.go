package feed

import (
	"context"
	"fmt"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
)

// Verify public type/key values remain independent after the server compacts
// their entire scope. Each feed deliberately uses the same ordering position.
func summaryCompactKeys(ctx context.Context, f *fixture) error {
	cases := [][2]string{
		{"channel-messages", f.key},
		{"scope:" + f.key, "item"}, {"scope", f.key + ":item"},
		{"concat" + f.key, "x"}, {"concat", f.key + "x"},
		{"notes:" + f.key, "日本語"}, {"notes:" + f.key + "日", "本語"},
		{"Case:" + f.key, "item"}, {"case:" + f.key, "item"},
		{"null" + f.key + "\x00", "item"}, {"null" + f.key, "\x00item"},
	}
	from := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	options := keystone.FeedSummaryOptions{From: from, Until: from.Add(24 * time.Hour)}
	feeds := make([]*keystone.Feed, len(cases))
	entries := make([]keystone.FeedEntry, len(cases))
	for i, pair := range cases {
		feeds[i] = f.actor.Feed(pair[0], pair[1])
		entries[i] = keystone.FeedEntry{Position: position(from), Data: []byte(fmt.Sprintf("namespace-%d", i))}
		if err := f.put(ctx, feeds[i], &entries[i]); err != nil {
			return err
		}
	}
	for i, feed := range feeds {
		if err := expect(ctx, feed, keystone.FeedReadOptions{}, entries[i:i+1]); err != nil {
			return fmt.Errorf("namespace %d aliased another feed: %w", i, err)
		}
		if _, err := readActivitySummary(ctx, feed, options, []summaryBucket{{from, options.Until, entries[i : i+1]}}); err != nil {
			return fmt.Errorf("namespace %d summary: %w", i, err)
		}
	}
	if err := f.remove(ctx, feeds[0], entries[0].Position); err != nil {
		return err
	}
	if err := expect(ctx, feeds[0], keystone.FeedReadOptions{}, nil); err != nil {
		return err
	}
	for i := 1; i < len(feeds); i++ {
		if err := expect(ctx, feeds[i], keystone.FeedReadOptions{}, entries[i:i+1]); err != nil {
			return fmt.Errorf("removal leaked into namespace %d: %w", i, err)
		}
	}
	return nil
}

// Exercise the signed int32 storage hour through the public SDK. Positions and
// cursors stay UUIDs/opaque strings; the client never sends an internal hour ID.
func summaryEpoch(ctx context.Context, f *fixture) error {
	from := time.Date(1969, 12, 31, 23, 0, 0, 0, time.UTC)
	feed := f.feed()
	var want []summaryBucket
	var all []keystone.FeedEntry
	for hour := 0; hour < 3; hour++ {
		start := from.Add(time.Duration(hour) * time.Hour)
		end := start.Add(time.Hour)
		entries, err := summaryEntries(ctx, f, feed, start, end.Add(-100*time.Nanosecond))
		if err != nil {
			return err
		}
		want = append(want, summaryBucket{start, end, entries})
		all = append(all, entries...)
	}
	options := keystone.FeedSummaryOptions{From: from, Until: from.Add(3 * time.Hour), Interval: keystone.FeedHourly}
	return summaryDirections(ctx, feed, options, want, all)
}
