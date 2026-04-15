package akv_timeline

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct{}

func (d *Requirement) Name() string {
	return "AKV Time Track"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	report(d.putAndGetLatest(actor))
	report(d.getAt(actor))
	report(d.earliestAfter(actor))
	report(d.rangeQuery(actor))
	report(d.rangeWithLimit(actor))
	report(d.rangeGroupedHourly(actor))
	report(d.rangeGroupedDaily(actor))
	report(d.rangeGroupedPickFirst(actor))
	report(d.batchPut(actor))
	report(d.deleteAll(actor))
	report(d.deleteExact(actor))
	report(d.deleteExactMultiple(actor))
	report(d.deleteRange(actor))
	report(d.globalVsWorkspace(actor))
	report(d.multipleKeys(actor))
	report(d.ttlPut(actor))
}

// seedRates writes 24 hourly exchange rate entries for the past 24 hours under the given key.
// Returns the base time (now truncated to second) and the entries written.
func (d *Requirement) seedRates(actor *keystone.Actor, key string) (time.Time, []keystone.AKVTimeEntry, error) {
	now := time.Now().UTC().Truncate(time.Second)
	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), key, "", keystone.TimeDeleteAll)
	entries := make([]keystone.AKVTimeEntry, 24)
	for i := 0; i < 24; i++ {
		rate := 1.2500 + float64(i)*0.0025
		entries[i] = keystone.AKVTimeEntry{
			Key:       key,
			Timestamp: now.Add(-time.Duration(23-i) * time.Hour),
			Data:      []byte(strconv.FormatFloat(rate, 'f', 4, 64)),
		}
	}

	resp, err := actor.AKVTimePut(context.Background(), entries...)
	if err != nil {
		return now, nil, err
	}
	if !resp.Success {
		return now, nil, fmt.Errorf("seed put failed: %s", resp.ErrorMessage)
	}
	return now, entries, nil
}

func (d *Requirement) putAndGetLatest(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Put 24 entries and GetLatest"}

	now, entries, err := d.seedRates(actor, "rate-usd-gbp")
	if err != nil {
		return res.WithError(err)
	}
	_ = now

	// Latest should be the last entry written (index 23)
	val, err := actor.AKVTimeGetLatest(context.Background(), "rate-usd-gbp", "")
	if err != nil {
		return res.WithError(err)
	}
	if val == nil {
		return res.WithError(errors.New("no result returned"))
	}
	expected := entries[23].Data
	if !bytes.Equal(val.Data, expected) {
		return res.WithError(fmt.Errorf("expected %s, got %s", expected, val.Data))
	}

	return res
}

func (d *Requirement) getAt(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "GetAt historical points"}

	now := time.Now().UTC().Truncate(time.Second)

	// Query at 12.5 hours ago — should get the entry at 13 hours ago (index 10, 13h ago = 23-13=10)
	pivot := now.Add(-12*time.Hour - 30*time.Minute)
	val, err := actor.AKVTimeGetAt(context.Background(), "rate-usd-gbp", "", pivot)
	if err != nil {
		return res.WithError(err)
	}
	if val == nil {
		return res.WithError(errors.New("no result returned for 12.5h ago"))
	}
	// Entry at index 10 = 23-10=13 hours ago, rate = 1.2500 + 10*0.0025 = 1.2750
	expected := "1.2750"
	if string(val.Data) != expected {
		return res.WithError(fmt.Errorf("at 12.5h ago: expected %s, got %s", expected, val.Data))
	}

	// Query at 0.5 hours ago — should get the entry at 1 hour ago (index 22)
	pivot = now.Add(-30 * time.Minute)
	val, err = actor.AKVTimeGetAt(context.Background(), "rate-usd-gbp", "", pivot)
	if err != nil {
		return res.WithError(err)
	}
	if val == nil {
		return res.WithError(errors.New("no result returned for 0.5h ago"))
	}
	// Entry at index 22 = 1 hour ago, rate = 1.2500 + 22*0.0025 = 1.3050
	expected = "1.3050"
	if string(val.Data) != expected {
		return res.WithError(fmt.Errorf("at 0.5h ago: expected %s, got %s", expected, val.Data))
	}

	// Query far before any data — should return nil
	val, err = actor.AKVTimeGetAt(context.Background(), "rate-usd-gbp", "", now.Add(-48*time.Hour))
	if err != nil {
		return res.WithError(err)
	}
	if val != nil {
		return res.WithError(fmt.Errorf("expected nil for query before all data, got %s", val.Data))
	}

	return res
}

func (d *Requirement) earliestAfter(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "EarliestAfter"}

	now := time.Now().UTC().Truncate(time.Second)

	// Query for earliest after 12.5 hours ago — should get the entry at 12 hours ago (index 11)
	pivot := now.Add(-12*time.Hour - 30*time.Minute)
	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:   "rate-usd-gbp",
		Mode:  keystone.TimeQueryEarliestAfter,
		Pivot: pivot,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 1 {
		return res.WithError(fmt.Errorf("expected 1 result, got %d", len(results)))
	}
	// Entry at index 11 = 12 hours ago, rate = 1.2500 + 11*0.0025 = 1.2775
	expected := "1.2775"
	if string(results[0].Data) != expected {
		return res.WithError(fmt.Errorf("expected %s, got %s", expected, results[0].Data))
	}

	return res
}

func (d *Requirement) rangeQuery(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Range query (all 24)"}

	now := time.Now().UTC().Truncate(time.Second)

	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "rate-usd-gbp",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-24 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 24 {
		return res.WithError(fmt.Errorf("expected 24 results, got %d", len(results)))
	}

	// First result should be the most recent (DESC order)
	expectedFirst := "1.3075" // index 23: 1.2500 + 23*0.0025
	if string(results[0].Data) != expectedFirst {
		return res.WithError(fmt.Errorf("first result should be latest (%s): got %s", expectedFirst, results[0].Data))
	}

	// Last result should be the oldest
	expectedLast := "1.2500" // index 0
	if string(results[23].Data) != expectedLast {
		return res.WithError(fmt.Errorf("last result should be oldest (%s): got %s", expectedLast, results[23].Data))
	}

	// Partial range: last 5.5 hours (should include entries at 0h, 1h, 2h, 3h, 4h, 5h = 6 entries)
	results, err = actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "rate-usd-gbp",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-5*time.Hour - 30*time.Minute),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 6 {
		return res.WithError(fmt.Errorf("expected 6 results for last 5.5h, got %d", len(results)))
	}

	return res
}

func (d *Requirement) rangeWithLimit(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Range query with limit"}

	now := time.Now().UTC().Truncate(time.Second)

	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "rate-usd-gbp",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-24 * time.Hour),
		RangeEnd:   now,
		Limit:      5,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 5 {
		return res.WithError(fmt.Errorf("expected 5 results with limit, got %d", len(results)))
	}

	return res
}

func (d *Requirement) rangeGroupedHourly(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Range grouped minutely (pick last)"}

	now := time.Now().UTC().Truncate(time.Second)

	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), "grp-minutely", "", keystone.TimeDeleteAll)

	// Anchor to a known minute boundary so we can predict which clock-minute each entry falls in.
	// Truncate to the current minute, then place entries at known offsets within distinct minutes.
	base := now.Truncate(time.Minute)
	// Minute -3: base-3m+10s, base-3m+20s, base-3m+30s
	// Minute -2: base-2m+10s, base-2m+20s, base-2m+30s
	// Minute -1: base-1m+10s, base-1m+20s, base-1m+30s
	var entries []keystone.AKVTimeEntry
	for m := 3; m >= 1; m-- {
		for s := 1; s <= 3; s++ {
			ts := base.Add(-time.Duration(m)*time.Minute + time.Duration(s*10)*time.Second)
			val := fmt.Sprintf("m%d-s%d", 3-m, s-1)
			entries = append(entries, keystone.AKVTimeEntry{
				Key:       "grp-minutely",
				Timestamp: ts,
				Data:      []byte(val),
			})
		}
	}
	_, err := actor.AKVTimePut(context.Background(), entries...)
	if err != nil {
		return res.WithError(err)
	}

	// Without grouping, should return all 9 (range well outside the entries)
	ungrouped, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "grp-minutely",
		Mode:       keystone.TimeQueryRange,
		RangeStart: base.Add(-4 * time.Minute),
		RangeEnd:   base,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(ungrouped) != 9 {
		return res.WithError(fmt.Errorf("expected 9 ungrouped results, got %d", len(ungrouped)))
	}

	// With minutely grouping + pick last, should reduce to 3 buckets (one per minute)
	grouped, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:           "grp-minutely",
		Mode:          keystone.TimeQueryRange,
		RangeStart:    base.Add(-4 * time.Minute),
		RangeEnd:      base,
		GroupInterval: keystone.TimeGroupMinutely,
		GroupPick:     keystone.TimePickLast,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(grouped) != 3 {
		return res.WithError(fmt.Errorf("expected 3 minutely grouped results, got %d", len(grouped)))
	}

	// Cleanup
	_, _ = actor.AKVTimeDel(context.Background(), "grp-minutely", "", keystone.TimeDeleteAll)
	return res
}

func (d *Requirement) rangeGroupedDaily(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Range grouped daily"}

	now := time.Now().UTC().Truncate(time.Second)

	// The 24 hourly entries for rate-usd-gbp span ~24 hours, which should be 1-2 daily buckets
	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:           "rate-usd-gbp",
		Mode:          keystone.TimeQueryRange,
		RangeStart:    now.Add(-25 * time.Hour),
		RangeEnd:      now,
		GroupInterval: keystone.TimeGroupDaily,
		GroupPick:     keystone.TimePickLast,
	})
	if err != nil {
		return res.WithError(err)
	}
	// Should be 1 or 2 daily buckets depending on UTC day boundaries
	if len(results) < 1 || len(results) > 2 {
		return res.WithError(fmt.Errorf("expected 1-2 daily buckets, got %d", len(results)))
	}

	return res
}

func (d *Requirement) rangeGroupedPickFirst(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Range grouped pick first vs last"}

	now := time.Now().UTC().Truncate(time.Second)

	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), "grp-pick", "", keystone.TimeDeleteAll)

	// Place 10 entries within a single clock minute to guarantee one minutely bucket.
	// Anchor to the start of the current minute, then write at +10s through +19s.
	base := now.Truncate(time.Minute)
	var entries []keystone.AKVTimeEntry
	for i := 0; i < 10; i++ {
		entries = append(entries, keystone.AKVTimeEntry{
			Key:       "grp-pick",
			Timestamp: base.Add(time.Duration(10+i) * time.Second),
			Data:      []byte(fmt.Sprintf("v%d", i)),
		})
	}
	_, err := actor.AKVTimePut(context.Background(), entries...)
	if err != nil {
		return res.WithError(err)
	}

	// Without grouping, should return all 10
	ungrouped, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "grp-pick",
		Mode:       keystone.TimeQueryRange,
		RangeStart: base,
		RangeEnd:   base.Add(1 * time.Minute),
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(ungrouped) != 10 {
		return res.WithError(fmt.Errorf("expected 10 ungrouped results, got %d", len(ungrouped)))
	}

	// Pick first — minutely grouping so all entries in one bucket, returns earliest
	resultsFirst, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:           "grp-pick",
		Mode:          keystone.TimeQueryRange,
		RangeStart:    base,
		RangeEnd:      base.Add(1 * time.Minute),
		GroupInterval: keystone.TimeGroupMinutely,
		GroupPick:     keystone.TimePickFirst,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(resultsFirst) != 1 {
		return res.WithError(fmt.Errorf("expected 1 grouped result (first), got %d", len(resultsFirst)))
	}
	if string(resultsFirst[0].Data) != "v0" {
		return res.WithError(fmt.Errorf("pick first: expected v0, got %s", resultsFirst[0].Data))
	}

	// Pick last — same bucket, returns most recent
	resultsLast, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:           "grp-pick",
		Mode:          keystone.TimeQueryRange,
		RangeStart:    base,
		RangeEnd:      base.Add(1 * time.Minute),
		GroupInterval: keystone.TimeGroupMinutely,
		GroupPick:     keystone.TimePickLast,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(resultsLast) != 1 {
		return res.WithError(fmt.Errorf("expected 1 grouped result (last), got %d", len(resultsLast)))
	}
	if string(resultsLast[0].Data) != "v9" {
		return res.WithError(fmt.Errorf("pick last: expected v9, got %s", resultsLast[0].Data))
	}

	// Cleanup
	_, _ = actor.AKVTimeDel(context.Background(), "grp-pick", "", keystone.TimeDeleteAll)
	return res
}

func (d *Requirement) batchPut(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Batch put 50 entries"}

	now := time.Now().UTC().Truncate(time.Second)

	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), "batch-test", "", keystone.TimeDeleteAll)

	// Write 50 entries at 1-minute intervals
	entries := make([]keystone.AKVTimeEntry, 50)
	for i := 0; i < 50; i++ {
		entries[i] = keystone.AKVTimeEntry{
			Key:       "batch-test",
			Timestamp: now.Add(-time.Duration(49-i) * time.Minute),
			Data:      []byte(fmt.Sprintf("entry-%02d", i)),
		}
	}

	resp, err := actor.AKVTimePut(context.Background(), entries...)
	if err != nil {
		return res.WithError(err)
	}
	if !resp.Success {
		return res.WithError(fmt.Errorf("batch put failed: %s", resp.ErrorMessage))
	}

	// Verify all 50 are retrievable
	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "batch-test",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-1 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 50 {
		return res.WithError(fmt.Errorf("expected 50 results, got %d", len(results)))
	}

	// Verify latest
	val, err := actor.AKVTimeGetLatest(context.Background(), "batch-test", "")
	if err != nil {
		return res.WithError(err)
	}
	if val == nil || string(val.Data) != "entry-49" {
		got := "<nil>"
		if val != nil {
			got = string(val.Data)
		}
		return res.WithError(fmt.Errorf("expected entry-49 as latest, got %s", got))
	}

	// Cleanup
	_, _ = actor.AKVTimeDel(context.Background(), "batch-test", "", keystone.TimeDeleteAll)
	return res
}

func (d *Requirement) deleteAll(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Delete All (24 entries)"}

	delResp, err := actor.AKVTimeDel(context.Background(), "rate-usd-gbp", "", keystone.TimeDeleteAll)
	if err != nil {
		return res.WithError(err)
	}
	if !delResp.Success {
		return res.WithError(fmt.Errorf("delete failed: %s", delResp.ErrorMessage))
	}

	val, err := actor.AKVTimeGetLatest(context.Background(), "rate-usd-gbp", "")
	if err != nil {
		return res.WithError(err)
	}
	if val != nil {
		return res.WithError(errors.New("expected no results after delete all"))
	}

	return res
}

func (d *Requirement) deleteExact(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Delete Exact (single)"}

	now := time.Now().UTC().Truncate(time.Second)

	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), "del-exact", "", keystone.TimeDeleteAll)

	// Write 10 entries
	var entries []keystone.AKVTimeEntry
	for i := 0; i < 10; i++ {
		entries = append(entries, keystone.AKVTimeEntry{
			Key:       "del-exact",
			Timestamp: now.Add(-time.Duration(9-i) * time.Hour),
			Data:      []byte(fmt.Sprintf("v%d", i)),
		})
	}
	_, err := actor.AKVTimePut(context.Background(), entries...)
	if err != nil {
		return res.WithError(err)
	}

	// Delete entry at index 5 (4 hours ago)
	delResp, err := actor.AKVTimeDel(context.Background(), "del-exact", "", keystone.TimeDeleteExact,
		keystone.WithTimestamps(now.Add(-4*time.Hour)))
	if err != nil {
		return res.WithError(err)
	}
	if !delResp.Success {
		return res.WithError(fmt.Errorf("delete failed: %s", delResp.ErrorMessage))
	}

	// Should have 9 entries remaining
	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "del-exact",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-10 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 9 {
		return res.WithError(fmt.Errorf("expected 9 results after deleting 1, got %d", len(results)))
	}

	// Cleanup
	_, _ = actor.AKVTimeDel(context.Background(), "del-exact", "", keystone.TimeDeleteAll)
	return res
}

func (d *Requirement) deleteExactMultiple(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Delete Exact (multiple timestamps)"}

	now := time.Now().UTC().Truncate(time.Second)

	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), "del-exact-multi", "", keystone.TimeDeleteAll)

	// Write 10 entries
	var entries []keystone.AKVTimeEntry
	for i := 0; i < 10; i++ {
		entries = append(entries, keystone.AKVTimeEntry{
			Key:       "del-exact-multi",
			Timestamp: now.Add(-time.Duration(9-i) * time.Hour),
			Data:      []byte(fmt.Sprintf("v%d", i)),
		})
	}
	_, err := actor.AKVTimePut(context.Background(), entries...)
	if err != nil {
		return res.WithError(err)
	}

	// Delete 3 specific timestamps (indices 2, 5, 8 => 7h, 4h, 1h ago)
	delResp, err := actor.AKVTimeDel(context.Background(), "del-exact-multi", "", keystone.TimeDeleteExact,
		keystone.WithTimestamps(
			now.Add(-7*time.Hour),
			now.Add(-4*time.Hour),
			now.Add(-1*time.Hour),
		))
	if err != nil {
		return res.WithError(err)
	}
	if !delResp.Success {
		return res.WithError(fmt.Errorf("delete failed: %s", delResp.ErrorMessage))
	}

	// Should have 7 entries remaining
	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "del-exact-multi",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-10 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 7 {
		return res.WithError(fmt.Errorf("expected 7 results after deleting 3, got %d", len(results)))
	}

	// Cleanup
	_, _ = actor.AKVTimeDel(context.Background(), "del-exact-multi", "", keystone.TimeDeleteAll)
	return res
}

func (d *Requirement) deleteRange(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Delete Time Range"}

	now := time.Now().UTC().Truncate(time.Second)

	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), "del-range", "", keystone.TimeDeleteAll)

	// Write 10 entries at hourly intervals
	var entries []keystone.AKVTimeEntry
	for i := 0; i < 10; i++ {
		entries = append(entries, keystone.AKVTimeEntry{
			Key:       "del-range",
			Timestamp: now.Add(-time.Duration(9-i) * time.Hour),
			Data:      []byte(fmt.Sprintf("v%d", i)),
		})
	}
	_, err := actor.AKVTimePut(context.Background(), entries...)
	if err != nil {
		return res.WithError(err)
	}

	// Delete entries from 6.5h ago to 3.5h ago (inclusive range catches entries at 6h, 5h, 4h ago = 3 entries)
	delResp, err := actor.AKVTimeDel(context.Background(), "del-range", "",
		keystone.TimeDeleteTimeRange,
		keystone.WithTimeRange(now.Add(-390*time.Minute), now.Add(-210*time.Minute)),
	)
	if err != nil {
		return res.WithError(err)
	}
	if !delResp.Success {
		return res.WithError(fmt.Errorf("delete range failed: %s", delResp.ErrorMessage))
	}

	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "del-range",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-10 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 7 {
		return res.WithError(fmt.Errorf("expected 7 results after range delete, got %d", len(results)))
	}

	// Cleanup
	_, _ = actor.AKVTimeDel(context.Background(), "del-range", "", keystone.TimeDeleteAll)
	return res
}

func (d *Requirement) globalVsWorkspace(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Global vs Workspace scoping"}

	now := time.Now().UTC().Truncate(time.Second)

	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), "scope-test", "", keystone.TimeDeleteAll)
	_, _ = actor.AKVTimeDel(context.Background(), "scope-test", "ws1", keystone.TimeDeleteAll)
	_, _ = actor.AKVTimeDel(context.Background(), "scope-test", "ws2", keystone.TimeDeleteAll)

	// Write 5 global entries
	var globalEntries []keystone.AKVTimeEntry
	for i := 0; i < 5; i++ {
		globalEntries = append(globalEntries, keystone.AKVTimeEntry{
			Key:       "scope-test",
			Workspace: "",
			Timestamp: now.Add(-time.Duration(4-i) * time.Hour),
			Data:      []byte(fmt.Sprintf("global-%d", i)),
		})
	}
	_, err := actor.AKVTimePut(context.Background(), globalEntries...)
	if err != nil {
		return res.WithError(err)
	}

	// Write 5 entries to workspace "ws1"
	var ws1Entries []keystone.AKVTimeEntry
	for i := 0; i < 5; i++ {
		ws1Entries = append(ws1Entries, keystone.AKVTimeEntry{
			Key:       "scope-test",
			Workspace: "ws1",
			Timestamp: now.Add(-time.Duration(4-i) * time.Hour),
			Data:      []byte(fmt.Sprintf("ws1-%d", i)),
		})
	}
	_, err = actor.AKVTimePut(context.Background(), ws1Entries...)
	if err != nil {
		return res.WithError(err)
	}

	// Write 3 entries to workspace "ws2"
	var ws2Entries []keystone.AKVTimeEntry
	for i := 0; i < 3; i++ {
		ws2Entries = append(ws2Entries, keystone.AKVTimeEntry{
			Key:       "scope-test",
			Workspace: "ws2",
			Timestamp: now.Add(-time.Duration(2-i) * time.Hour),
			Data:      []byte(fmt.Sprintf("ws2-%d", i)),
		})
	}
	_, err = actor.AKVTimePut(context.Background(), ws2Entries...)
	if err != nil {
		return res.WithError(err)
	}

	// Global latest should be "global-4"
	val, err := actor.AKVTimeGetLatest(context.Background(), "scope-test", "")
	if err != nil {
		return res.WithError(err)
	}
	if val == nil || string(val.Data) != "global-4" {
		got := "<nil>"
		if val != nil {
			got = string(val.Data)
		}
		return res.WithError(fmt.Errorf("global latest: expected global-4, got %s", got))
	}

	// ws1 latest should be "ws1-4"
	val, err = actor.AKVTimeGetLatest(context.Background(), "scope-test", "ws1")
	if err != nil {
		return res.WithError(err)
	}
	if val == nil || string(val.Data) != "ws1-4" {
		got := "<nil>"
		if val != nil {
			got = string(val.Data)
		}
		return res.WithError(fmt.Errorf("ws1 latest: expected ws1-4, got %s", got))
	}

	// ws2 latest should be "ws2-2"
	val, err = actor.AKVTimeGetLatest(context.Background(), "scope-test", "ws2")
	if err != nil {
		return res.WithError(err)
	}
	if val == nil || string(val.Data) != "ws2-2" {
		got := "<nil>"
		if val != nil {
			got = string(val.Data)
		}
		return res.WithError(fmt.Errorf("ws2 latest: expected ws2-2, got %s", got))
	}

	// Global range should return 5
	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "scope-test",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-5 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 5 {
		return res.WithError(fmt.Errorf("global range: expected 5 results, got %d", len(results)))
	}

	// ws2 range should return 3
	results, err = actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "scope-test",
		Workspace:  "ws2",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-5 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 3 {
		return res.WithError(fmt.Errorf("ws2 range: expected 3 results, got %d", len(results)))
	}

	// Cleanup
	_, _ = actor.AKVTimeDel(context.Background(), "scope-test", "", keystone.TimeDeleteAll)
	_, _ = actor.AKVTimeDel(context.Background(), "scope-test", "ws1", keystone.TimeDeleteAll)
	_, _ = actor.AKVTimeDel(context.Background(), "scope-test", "ws2", keystone.TimeDeleteAll)
	return res
}

func (d *Requirement) multipleKeys(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Multiple keys isolation"}

	now := time.Now().UTC().Truncate(time.Second)

	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), "key-alpha", "", keystone.TimeDeleteAll)
	_, _ = actor.AKVTimeDel(context.Background(), "key-beta", "", keystone.TimeDeleteAll)

	// Write 8 entries across 2 different keys
	var entries []keystone.AKVTimeEntry
	for i := 0; i < 8; i++ {
		entries = append(entries, keystone.AKVTimeEntry{
			Key:       "key-alpha",
			Timestamp: now.Add(-time.Duration(7-i) * time.Hour),
			Data:      []byte(fmt.Sprintf("alpha-%d", i)),
		})
	}
	_, err := actor.AKVTimePut(context.Background(), entries...)
	if err != nil {
		return res.WithError(err)
	}

	entries = entries[:0]
	for i := 0; i < 5; i++ {
		entries = append(entries, keystone.AKVTimeEntry{
			Key:       "key-beta",
			Timestamp: now.Add(-time.Duration(4-i) * time.Hour),
			Data:      []byte(fmt.Sprintf("beta-%d", i)),
		})
	}
	_, err = actor.AKVTimePut(context.Background(), entries...)
	if err != nil {
		return res.WithError(err)
	}

	// Alpha should have 8, beta should have 5
	alphaResults, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "key-alpha",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-8 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(alphaResults) != 8 {
		return res.WithError(fmt.Errorf("alpha: expected 8, got %d", len(alphaResults)))
	}

	betaResults, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "key-beta",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-5 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(betaResults) != 5 {
		return res.WithError(fmt.Errorf("beta: expected 5, got %d", len(betaResults)))
	}

	// Deleting alpha should not affect beta
	_, _ = actor.AKVTimeDel(context.Background(), "key-alpha", "", keystone.TimeDeleteAll)

	betaVal, err := actor.AKVTimeGetLatest(context.Background(), "key-beta", "")
	if err != nil {
		return res.WithError(err)
	}
	if betaVal == nil || string(betaVal.Data) != "beta-4" {
		return res.WithError(errors.New("beta should be unaffected after deleting alpha"))
	}

	// Cleanup
	_, _ = actor.AKVTimeDel(context.Background(), "key-beta", "", keystone.TimeDeleteAll)
	return res
}

func (d *Requirement) ttlPut(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Put with TTL"}

	now := time.Now().UTC().Truncate(time.Second)

	// Clean any leftover data from previous runs
	_, _ = actor.AKVTimeDel(context.Background(), "ttl-test", "", keystone.TimeDeleteAll)

	// Write entries with TTL — just verify the put succeeds (TTL expiry is server-side)
	resp, err := actor.AKVTimePut(context.Background(),
		keystone.AKVTimeEntry{
			Key:       "ttl-test",
			Timestamp: now,
			Data:      []byte("expires-soon"),
			TTL:       1 * time.Hour,
		},
		keystone.AKVTimeEntry{
			Key:       "ttl-test",
			Timestamp: now.Add(-1 * time.Hour),
			Data:      []byte("expires-later"),
			TTL:       24 * time.Hour,
		},
	)
	if err != nil {
		return res.WithError(err)
	}
	if !resp.Success {
		return res.WithError(fmt.Errorf("put with TTL failed: %s", resp.ErrorMessage))
	}

	// Verify both are readable immediately
	results, err := actor.AKVTimeGet(context.Background(), keystone.AKVTimeQuery{
		Key:        "ttl-test",
		Mode:       keystone.TimeQueryRange,
		RangeStart: now.Add(-2 * time.Hour),
		RangeEnd:   now,
	})
	if err != nil {
		return res.WithError(err)
	}
	if len(results) != 2 {
		return res.WithError(fmt.Errorf("expected 2 TTL entries, got %d", len(results)))
	}

	// Cleanup
	_, _ = actor.AKVTimeDel(context.Background(), "ttl-test", "", keystone.TimeDeleteAll)
	return res
}
