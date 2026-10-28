package feed

import (
	"testing"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
)

func TestActivityRequirementsAllowDriftButRejectMissingPeriods(t *testing.T) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	want := []summaryBucket{{start, end, []keystone.FeedEntry{
		{Position: position(start.Add(time.Minute))},
		{Position: position(start.Add(2 * time.Minute))},
	}}}
	active := keystone.FeedSummaryBucket{Start: start, End: end, Count: 1, Cursor: "jump"}
	for _, test := range []struct {
		name    string
		buckets []keystone.FeedSummaryBucket
		valid   bool
	}{
		{"low density estimate still exposes activity", []keystone.FeedSummaryBucket{active}, true},
		{"extra historical empty period is allowed", []keystone.FeedSummaryBucket{active, {Start: end, End: end.Add(24 * time.Hour), Count: 1, Cursor: "empty-jump"}}, true},
		{"missing active period", nil, false},
		{"wrong calendar interval", []keystone.FeedSummaryBucket{{Start: start.Add(-time.Hour), End: end, Count: 1, Cursor: "jump"}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := checkActivitySummary(&keystone.FeedSummary{Buckets: test.buckets, Approximate: true}, want)
			if (err == nil) != test.valid {
				t.Fatalf("validation error=%v, want valid=%v", err, test.valid)
			}
		})
	}
}
