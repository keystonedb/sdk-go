package feed

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
)

// SummaryRequirement exercises immediate activity outlines and navigation cursors.
// No entity schema or SQL index is needed for any of these scenarios.
type SummaryRequirement struct{}

type summaryScenario struct {
	name string
	run  func(context.Context, *fixture) error
}

var _ requirements.Requirement = (*SummaryRequirement)(nil)

func (*SummaryRequirement) Name() string                        { return "Feed Summaries" }
func (*SummaryRequirement) Register(*keystone.Connection) error { return nil }

func (*SummaryRequirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	scenarios := []summaryScenario{
		{"Empty summary and UTC/daily defaults", summaryEmpty},
		{"Clipped inclusive/exclusive range and jump continuation", summaryClipped},
		{"Conservative activity after retry, update, insert and removal", summaryLifecycle},
		{"Summary and jump cursor scope isolation", summaryIsolation},
		{"Compact feed keys and namespace boundaries", summaryCompactKeys},
		{"Signed UTC hour boundaries across the Unix epoch", summaryEpoch},
		{"Summary validation and maximum ranges", summaryValidation},
	}
	for _, calendar := range []struct {
		name, zone, start string
		interval          keystone.FeedSummaryInterval
	}{
		{"UTC hourly", "UTC", "2025-06-01T00:00:00Z", keystone.FeedHourly},
		{"UTC daily defaults", "", "2025-06-01T00:00:00Z", keystone.FeedDaily},
		{"UTC monthly and year rollover", "UTC", "2024-12-01T00:00:00Z", keystone.FeedMonthly},
		{"Monthly leap-year February", "UTC", "2024-01-01T00:00:00Z", keystone.FeedMonthly},
		{"London spring DST days", "Europe/London", "2025-03-29T00:00:00Z", keystone.FeedDaily},
		{"London autumn DST days", "Europe/London", "2025-10-25T00:00:00+01:00", keystone.FeedDaily},
		{"London missing spring hour", "Europe/London", "2025-03-30T00:00:00Z", keystone.FeedHourly},
		{"London repeated autumn hour", "Europe/London", "2025-10-26T00:00:00+01:00", keystone.FeedHourly},
		{"Kolkata half-hour offset", "Asia/Kolkata", "2025-06-01T00:00:00+05:30", keystone.FeedHourly},
		{"Kathmandu quarter-hour offset", "Asia/Kathmandu", "2025-06-01T00:00:00+05:45", keystone.FeedHourly},
		{"Kolkata local calendar days", "Asia/Kolkata", "2025-06-01T00:00:00+05:30", keystone.FeedDaily},
		{"Kathmandu local calendar days", "Asia/Kathmandu", "2025-06-01T00:00:00+05:45", keystone.FeedDaily},
	} {
		scenarios = append(scenarios, summaryScenario{calendar.name, func(ctx context.Context, f *fixture) error {
			return summaryCalendar(ctx, f, calendar.zone, calendar.start, calendar.interval)
		}})
	}
	for _, scenario := range scenarios {
		log.Printf("Running feed summary scenario: %s", scenario.name)
		// Activity outlines must be available immediately after acknowledged writes.
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

type summaryBucket struct {
	start, end time.Time
	entries    []keystone.FeedEntry // Chronological membership order.
}

func summaryEmpty(ctx context.Context, f *fixture) error {
	from := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	for _, interval := range []keystone.FeedSummaryInterval{keystone.FeedDaily, keystone.FeedHourly, keystone.FeedMonthly} {
		options := keystone.FeedSummaryOptions{From: from, Until: from.Add(24 * time.Hour), Interval: interval}
		summary, err := readActivitySummary(ctx, f.feed(), options, nil)
		if err != nil {
			return err
		}
		if len(summary.Buckets) != 0 {
			return fmt.Errorf("unwritten feed returned activity markers")
		}
	}
	return nil
}

func summaryEntries(ctx context.Context, f *fixture, feed *keystone.Feed, times ...time.Time) ([]keystone.FeedEntry, error) {
	entries := make([]keystone.FeedEntry, len(times))
	for i, at := range times {
		entries[i] = keystone.FeedEntry{Position: position(at), Data: []byte(at.Format(time.RFC3339Nano))}
	}
	for i := len(entries) - 1; i >= 0; i-- {
		if err := f.put(ctx, feed, &entries[i]); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func summaryCalendar(ctx context.Context, f *fixture, zone, start string, interval keystone.FeedSummaryInterval) error {
	zoneName := zone
	if zoneName == "" {
		zoneName = "UTC"
	}
	loc, err := time.LoadLocation(zoneName)
	if err != nil {
		return err
	}
	from, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return err
	}
	from = from.In(loc)
	next := func(at time.Time) time.Time {
		switch interval {
		case keystone.FeedHourly:
			return at.Add(time.Hour)
		case keystone.FeedMonthly:
			return at.AddDate(0, 1, 0)
		default:
			return at.AddDate(0, 0, 1)
		}
	}
	feed := f.feed()
	var want []summaryBucket
	var all []keystone.FeedEntry
	until := from
	for i := 0; i < 5; i++ {
		end := next(until)
		// Leave an interior period with no writes. Conservative outlines may
		// retain it when an occupied UTC hour overlaps its boundary. The DST
		// transition itself remains populated on both calendar-day scenarios.
		if i != 3 {
			entries, err := summaryEntries(ctx, f, feed, until, end.Add(-time.Minute))
			if err != nil {
				return err
			}
			want = append(want, summaryBucket{until, end, entries})
			all = append(all, entries...)
		}
		until = end
	}
	// Keep entries outside both range ends to prove jumps continue through the
	// feed rather than stopping at a summary or bucket boundary.
	outside, err := summaryEntries(ctx, f, feed, from.Add(-time.Minute), until)
	if err != nil {
		return err
	}
	all = append([]keystone.FeedEntry{outside[0]}, append(all, outside[1])...)
	options := keystone.FeedSummaryOptions{From: from, Until: until, Interval: interval, Timezone: zone}
	return summaryDirections(ctx, feed, options, want, all)
}

func summaryClipped(ctx context.Context, f *fixture) error {
	day := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	from, until := day.Add(15*time.Minute), day.Add(75*time.Minute)
	feed := f.feed()
	all, err := summaryEntries(ctx, f, feed, from.Add(-100*time.Nanosecond), from, from.Add(time.Minute), until.Add(-100*time.Nanosecond), until)
	if err != nil {
		return err
	}
	want := []summaryBucket{{day, day.Add(24 * time.Hour), all[1:4]}}
	options := keystone.FeedSummaryOptions{From: from, Until: until}
	// Counts may conservatively include the whole occupied hour. Jump cursors
	// must honor the clipped request and include its boundary members.
	return summaryDirections(ctx, feed, options, want, all)
}

func summaryDirections(ctx context.Context, feed *keystone.Feed, options keystone.FeedSummaryOptions, want []summaryBucket, all []keystone.FeedEntry) error {
	for _, ascending := range []bool{false, true} {
		options.Ascending = ascending
		summary, err := readActivitySummary(ctx, feed, options, want)
		if err != nil {
			return err
		}
		if err := checkActivityJumps(ctx, feed, options, summary, all); err != nil {
			return err
		}
	}
	return nil
}

// Activity visibility comes from the durable directory. A missing period after
// an acknowledged write is a failure, not a reason to poll a background worker.
func readActivitySummary(ctx context.Context, feed *keystone.Feed, options keystone.FeedSummaryOptions, want []summaryBucket) (*keystone.FeedSummary, error) {
	summary, err := feed.Summary(ctx, options)
	if err != nil {
		return nil, err
	}
	if !summary.Approximate {
		return nil, fmt.Errorf("expected a conservative activity outline")
	}
	if err := checkActivitySummary(summary, want); err != nil {
		return nil, err
	}
	return summary, nil
}

func summaryLifecycle(ctx context.Context, f *fixture) error {
	from := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	feed := f.feed()
	entries, err := summaryEntries(ctx, f, feed, from, from.Add(time.Minute))
	if err != nil {
		return err
	}
	options := keystone.FeedSummaryOptions{From: from, Until: from.Add(24 * time.Hour)}
	want := []summaryBucket{{from, options.Until, entries}}
	if _, err := readActivitySummary(ctx, feed, options, want); err != nil {
		return err
	}
	entries[0].Data = []byte("updated without adding membership")
	for i := 0; i < 2; i++ {
		if err := feed.Put(ctx, &entries[0]); err != nil {
			return err
		}
	}
	// A new member must become navigable immediately, regardless of density
	// counter delivery or how many times the earlier membership was retried.
	extra, err := summaryEntries(ctx, f, feed, from.Add(2*time.Minute))
	if err != nil {
		return err
	}
	entries = append(entries, extra...)
	want[0].entries = entries
	if _, err := readActivitySummary(ctx, feed, options, want); err != nil {
		return err
	}
	if err := f.remove(ctx, feed, entries[0].Position); err != nil {
		return err
	}
	want[0].entries = entries[1:]
	if _, err := readActivitySummary(ctx, feed, options, want); err != nil {
		return err
	}
	for _, entry := range entries[1:] {
		if err := f.remove(ctx, feed, entry.Position); err != nil {
			return err
		}
	}
	_, err = readActivitySummary(ctx, feed, options, nil)
	return err
}

func summaryIsolation(ctx context.Context, f *fixture) error {
	from := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	options := keystone.FeedSummaryOptions{From: from, Until: from.Add(24 * time.Hour)}
	workspace := f.otherWorkspaceActor()
	feeds := []*keystone.Feed{f.feed(), f.actor.Feed("requirement-other", f.key), f.actor.Feed("requirement-activity", f.key+"-other"), workspace.Feed("requirement-activity", f.key),
		f.sourceActor(f.actor.VendorID(), f.actor.AppID()+"-feed-other").Feed("requirement-activity", f.key),
		f.sourceActor(f.actor.VendorID()+"-feed-other", f.actor.AppID()).Feed("requirement-activity", f.key)}
	for i, feed := range feeds {
		scopeFrom := from.AddDate(0, 0, i)
		scopeOptions := keystone.FeedSummaryOptions{From: scopeFrom, Until: scopeFrom.Add(24 * time.Hour)}
		times := make([]time.Time, i+1)
		for j := range times {
			times[j] = scopeFrom.Add(time.Duration(j) * time.Minute)
		}
		entries, err := summaryEntries(ctx, f, feed, times...)
		if err != nil {
			return err
		}
		summary, err := readActivitySummary(ctx, feed, scopeOptions, []summaryBucket{{scopeFrom, scopeOptions.Until, entries}})
		if err != nil {
			return fmt.Errorf("scope %d: %w", i, err)
		}
		otherUser := f.actor.Connection().Actor(f.actor.WorkspaceID(), f.actor.RemoteIP(), f.key, f.actor.UserAgent())
		if _, err := otherUser.Feed("requirement-activity", f.key).Read(ctx, keystone.FeedReadOptions{Cursor: summary.Buckets[0].Cursor}); err == nil {
			return fmt.Errorf("scope %d jump cursor accepted by a different user", i)
		}
		for j, other := range feeds {
			if i == j {
				continue
			}
			if _, err := other.Read(ctx, keystone.FeedReadOptions{Cursor: summary.Buckets[0].Cursor}); err == nil {
				return fmt.Errorf("scope %d jump cursor accepted in scope %d", i, j)
			}
		}
	}
	// The populated primary scope must remain unchanged after all other writes.
	entry := keystone.FeedEntry{Position: position(from), Data: []byte(from.Format(time.RFC3339Nano))}
	_, err := readActivitySummary(ctx, feeds[0], options, []summaryBucket{{from, options.Until, []keystone.FeedEntry{entry}}})
	return err
}

func summaryValidation(ctx context.Context, f *fixture) error {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	feed := f.feed()
	for i, options := range []keystone.FeedSummaryOptions{
		{}, {From: from}, {Until: from}, {From: from, Until: from}, {From: from, Until: from.Add(-time.Minute)},
		{From: from.Add(time.Second), Until: from.Add(time.Hour)}, {From: from, Until: from.Add(time.Hour + time.Nanosecond)},
		{From: from, Until: from.Add(time.Hour), Interval: -1}, {From: from, Until: from.Add(time.Hour), Interval: 99},
		{From: from, Until: from.Add(31*24*time.Hour + time.Minute), Interval: keystone.FeedHourly},
		{From: from, Until: from.Add(366*24*time.Hour + time.Minute)},
		{From: from, Until: from.Add(366*24*time.Hour + time.Minute), Interval: keystone.FeedMonthly},
		{From: time.Date(-1, 1, 1, 0, 0, 0, 0, time.UTC), Until: time.Date(-1, 1, 2, 0, 0, 0, 0, time.UTC)},
		{From: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC), Until: time.Date(10000, 1, 2, 0, 0, 0, 0, time.UTC)},
		{From: from, Until: from.Add(time.Hour), Timezone: "Invalid/Timezone"},
	} {
		if _, err := feed.Summary(ctx, options); err == nil {
			return fmt.Errorf("accepted invalid summary options %d: %+v", i, options)
		}
	}
	valid := keystone.FeedSummaryOptions{From: from, Until: from.Add(time.Hour)}
	var nilActor *keystone.Actor
	for _, invalid := range []*keystone.Feed{nil, nilActor.Feed("type", "key"), (&keystone.Actor{}).Feed("type", "key"), f.actor.Feed("", f.key), f.actor.Feed("type", ""), f.actor.Feed(strings.Repeat("t", 129), f.key), f.actor.Feed("type", strings.Repeat("k", 513)), f.actor.Feed(strings.Repeat("é", 65), f.key)} {
		if _, err := invalid.Summary(ctx, valid); err == nil {
			return fmt.Errorf("accepted invalid feed for summary")
		}
	}
	boundary := f.actor.Feed(strings.Repeat("t", 128), f.key+strings.Repeat("k", 512-len(f.key)))
	if _, err := readActivitySummary(ctx, boundary, valid, nil); err != nil {
		return fmt.Errorf("valid maximum feed type and key lengths: %w", err)
	}
	valid.Until = from.Add(time.Minute)
	if _, err := readActivitySummary(ctx, feed, valid, nil); err != nil {
		return fmt.Errorf("valid one-minute range: %w", err)
	}
	for _, interval := range []keystone.FeedSummaryInterval{keystone.FeedHourly, keystone.FeedDaily, keystone.FeedMonthly} {
		days := 366
		if interval == keystone.FeedHourly {
			days = 31
		}
		options := keystone.FeedSummaryOptions{From: from, Until: from.Add(time.Duration(days) * 24 * time.Hour), Interval: interval}
		if _, err := readActivitySummary(ctx, feed, options, nil); err != nil {
			return fmt.Errorf("valid maximum range for interval %d: %w", interval, err)
		}
	}
	return nil
}

// Approximate summaries promise coverage of active periods, not exact density
// or removal of empty periods. These assertions deliberately preserve that
// stronger navigation guarantee without waiting for background maintenance.
func checkActivitySummary(summary *keystone.FeedSummary, want []summaryBucket) error {
	for i, bucket := range summary.Buckets {
		if bucket.Count < 1 || bucket.Cursor == "" || !bucket.Start.Before(bucket.End) || (i > 0 && bucket.Start.Before(summary.Buckets[i-1].End)) {
			return fmt.Errorf("invalid activity marker: %v", bucket)
		}
	}
	for _, expected := range want {
		for _, entry := range expected.entries {
			at, err := summaryEntryTime(entry)
			if err != nil {
				return err
			}
			found := false
			for _, bucket := range summary.Buckets {
				if !at.Before(bucket.Start) && at.Before(bucket.End) {
					if !bucket.Start.Equal(expected.start) || !bucket.End.Equal(expected.end) {
						return fmt.Errorf("activity period [%s,%s) has wrong calendar boundaries, want [%s,%s)", bucket.Start, bucket.End, expected.start, expected.end)
					}
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("active period containing %s was hidden", at)
			}
		}
	}
	return nil
}

func summaryEntryTime(entry keystone.FeedEntry) (time.Time, error) {
	id, err := uuid.Parse(entry.Position)
	if err != nil {
		return time.Time{}, err
	}
	seconds, nanos := id.Time().UnixTime()
	return time.Unix(seconds, nanos), nil
}

func checkActivityJumps(ctx context.Context, feed *keystone.Feed, options keystone.FeedSummaryOptions, summary *keystone.FeedSummary, all []keystone.FeedEntry) error {
	for _, bucket := range summary.Buckets {
		start, end := bucket.Start, bucket.End
		if start.Before(options.From) {
			start = options.From
		}
		if end.After(options.Until) {
			end = options.Until
		}
		var want []keystone.FeedEntry
		for _, entry := range all {
			at, err := summaryEntryTime(entry)
			if err != nil {
				return err
			}
			if (options.Ascending && !at.Before(start)) || (!options.Ascending && at.Before(end)) {
				want = append(want, entry)
			}
		}
		if !options.Ascending {
			slices.Reverse(want)
		}
		if err := expect(ctx, feed, keystone.FeedReadOptions{Cursor: bucket.Cursor, Ascending: options.Ascending, Limit: 1}, want); err != nil {
			return fmt.Errorf("activity jump: %w", err)
		}
		if _, err := feed.Read(ctx, keystone.FeedReadOptions{Cursor: bucket.Cursor, Ascending: !options.Ascending}); err == nil {
			return fmt.Errorf("activity jump cursor accepted opposite direction")
		}
	}
	return nil
}
