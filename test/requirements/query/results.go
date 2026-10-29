package query

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/requirements"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Separate fixtures keep snapshot mutations isolated from the legacy query cases.
type querySample struct {
	keystone.BaseEntity
	Run      string  `keystone:",indexed"`
	Category string  `keystone:",indexed"`
	Position int64   `keystone:",indexed"`
	Amount   int64   `keystone:",indexed"`
	Weight   float64 `keystone:",indexed"`
	Label    string
}

func (d *Requirement) resultFixtures(ctx context.Context, actor *keystone.Actor, scope string) ([]*querySample, error) {
	samples := make([]*querySample, 0, 6)
	for i, amount := range []int64{2, 2, 8, 4, 10, 6} {
		sample := &querySample{Run: d.runID + scope, Category: []string{"a", "a", "a", "b", "b", "c"}[i], Position: int64(i + 1), Amount: amount, Weight: float64(amount) * 0.75, Label: fmt.Sprintf("row-%d", i+1)}
		if err := actor.Mutate(ctx, sample); err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func (d *Requirement) readSavedEntityPages(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Query saved entity membership, fresh projections and token replay"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	samples, err := d.resultFixtures(ctx, actor, "saved-entities")
	if err != nil {
		return result.WithError(err)
	}
	pages, err := actor.QueryPages(ctx, keystone.Type(querySample{}), keystone.WhereEquals("run", d.runID+"saved-entities"), keystone.Select("label", "position"), keystone.SortAsc("position"), keystone.Limit(2), keystone.SaveResultSet(), keystone.IncludeTotal(), keystone.Explain())
	if err != nil {
		return result.WithError(err)
	}
	first := pages.Response()
	expectedExpiry := time.Now().Add(24 * time.Hour)
	if first.ResultSet == "" || first.ResultSetExpires == nil || first.ResultSetExpires.CheckValid() != nil || first.ResultSetExpires.AsTime().Before(expectedExpiry.Add(-time.Minute)) || first.ResultSetExpires.AsTime().After(expectedExpiry.Add(time.Minute)) {
		return result.WithError(fmt.Errorf("missing or invalid saved token expiry: %v", first))
	}
	samples[5].Label = "updated-after-save"
	if err := actor.Mutate(ctx, samples[5]); err != nil {
		return result.WithError(err)
	}
	// This would sort first in a live query, but cannot join the saved membership.
	if err := actor.Mutate(ctx, &querySample{Run: d.runID + "saved-entities", Category: "new", Position: -1, Amount: 1}); err != nil {
		return result.WithError(err)
	}
	seen := map[keystone.ID]bool{}
	for page := 0; page < 3; page++ {
		response := pages.Response()
		if len(response.Entities) != 2 || len(response.Rows) != 0 || response.ResultSet != first.ResultSet || response.TotalResults == nil || response.GetTotalResults() != 6 || pages.HasMore() != (page < 2) {
			return result.WithError(fmt.Errorf("unexpected saved entity page %d: %v", page, response))
		}
		if page > 0 && response.GetExecution().GetStrategy() != "result_set" {
			return result.WithError(fmt.Errorf("expected saved page execution strategy"))
		}
		entities, err := keystone.AsSlice[querySample](response.Entities...)
		if err != nil {
			return result.WithError(err)
		}
		for i, entity := range entities {
			expected := samples[page*2+i]
			if entity.GetKeystoneID() != expected.GetKeystoneID() || seen[entity.GetKeystoneID()] || entity.Position != expected.Position || entity.Label != expected.Label || entity.Run != "" || entity.Amount != 0 {
				return result.WithError(fmt.Errorf("unexpected saved entity: %+v, expected ID %s and label %q", entity, expected.GetKeystoneID(), expected.Label))
			}
			seen[entity.GetKeystoneID()] = true
		}
		if page < 2 {
			if err := pages.Next(ctx); err != nil {
				return result.WithError(err)
			}
		}
	}
	if !errors.Is(pages.Next(ctx), io.EOF) {
		return result.WithError(fmt.Errorf("exhausted saved pager must return EOF"))
	}
	// Omit the cursor to rewind and change page size without resending the query.
	replay, err := actor.Query(ctx, "", keystone.ResultSet(first.ResultSet), keystone.Limit(3), keystone.IncludeTotal())
	if err != nil {
		return result.WithError(err)
	}
	if len(replay.Entities) != 3 || replay.Entities[0].GetEntity().GetEntityId() != samples[0].GetKeystoneID().String() || replay.GetTotalResults() != 6 || !replay.HasMore {
		return result.WithError(fmt.Errorf("unexpected saved replay: %v", replay))
	}
	live, err := actor.Query(ctx, keystone.Type(querySample{}), keystone.WhereEquals("run", d.runID+"saved-entities"), keystone.IncludeTotal())
	if err != nil {
		return result.WithError(err)
	}
	if live.GetTotalResults() != 7 {
		return result.WithError(fmt.Errorf("live query should include the inserted entity: %v", live))
	}
	return result
}

func (d *Requirement) readSavedCollection(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Typed collection saved result cursoring and exhaustion"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	samples, err := d.resultFixtures(ctx, actor, "collection")
	if err != nil {
		return result.WithError(err)
	}
	collection := keystone.Collection((*querySample)(nil), actor)
	if err := collection.Query(ctx, keystone.WhereEquals("run", d.runID+"collection"), keystone.Select("label", "position"), keystone.SortAsc("position"), keystone.SaveResultSet(), keystone.Limit(2)); err != nil {
		return result.WithError(err)
	}
	token := collection.Response().ResultSet
	if token == "" {
		return result.WithError(fmt.Errorf("collection did not save its results"))
	}
	for page := 0; page < 3; page++ {
		if len(collection.Items()) != (page+1)*2 || collection.HasMore() != (page < 2) || collection.Response().ResultSet != token {
			return result.WithError(fmt.Errorf("unexpected collection state on page %d", page))
		}
		if page < 2 {
			if err := collection.Next(ctx); err != nil {
				return result.WithError(err)
			}
		}
	}
	for i, entity := range collection.Items() {
		if entity.GetKeystoneID() != samples[i].GetKeystoneID() || entity.Label != samples[i].Label || entity.Position != samples[i].Position {
			return result.WithError(fmt.Errorf("unexpected collection item %d: %+v", i, entity))
		}
	}
	if !errors.Is(collection.Next(ctx), io.EOF) {
		return result.WithError(fmt.Errorf("exhausted collection must return EOF"))
	}
	return result
}

func (d *Requirement) readEmptyResults(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Query empty aggregates and saved results"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	response, err := actor.Query(ctx, keystone.Type(querySample{}), keystone.WithEntityIDs(nil), keystone.Count("count"), keystone.Sum("amount", "sum"), keystone.Avg("amount", "avg"), keystone.Min("amount", "min"), keystone.Max("amount", "max"))
	if err != nil {
		return result.WithError(err)
	}
	if len(response.Rows) != 1 || response.Rows[0].Values["count"] == nil || response.Rows[0].Values["count"].GetInt() != 0 {
		return result.WithError(fmt.Errorf("empty aggregate must return one zero count: %v", response))
	}
	for _, alias := range []string{"sum", "avg", "min", "max"} {
		if value := response.Rows[0].Values[alias]; value == nil || !value.IsNull {
			return result.WithError(fmt.Errorf("empty %s must be SQL NULL: %v", alias, value))
		}
	}
	for _, grouped := range []bool{false, true} {
		options := []keystone.QueryOption{keystone.WithEntityIDs(nil), keystone.SaveResultSet(), keystone.IncludeTotal()}
		if grouped {
			options = append(options, keystone.GroupBy("category"), keystone.Count("count"))
		}
		pages, err := actor.QueryPages(ctx, keystone.Type(querySample{}), options...)
		if err != nil {
			return result.WithError(err)
		}
		response := pages.Response()
		if response.ResultSet == "" || response.ResultSetExpires == nil || response.TotalResults == nil || response.GetTotalResults() != 0 || len(response.Entities)+len(response.Rows) != 0 || pages.HasMore() || pages.NextCursor() != "" {
			return result.WithError(fmt.Errorf("unexpected empty saved result (grouped=%t): %v", grouped, response))
		}
		replay, err := actor.Query(ctx, "", keystone.ResultSet(response.ResultSet), keystone.IncludeTotal())
		if err != nil {
			return result.WithError(err)
		}
		if replay.TotalResults == nil || replay.GetTotalResults() != 0 || len(replay.Entities)+len(replay.Rows) != 0 || replay.HasMore {
			return result.WithError(fmt.Errorf("empty saved replay returned results: %v", replay))
		}
		if !errors.Is(pages.Next(ctx), io.EOF) {
			return result.WithError(fmt.Errorf("empty pager must return EOF"))
		}
	}
	return result
}

func (d *Requirement) readInvalidResultSets(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Query result token validation and cursor isolation"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := d.resultFixtures(ctx, actor, "invalid"); err != nil {
		return result.WithError(err)
	}
	options := []keystone.QueryOption{keystone.WhereEquals("run", d.runID+"invalid"), keystone.SortAsc("position"), keystone.SaveResultSet(), keystone.Limit(1)}
	first, err := actor.Query(ctx, keystone.Type(querySample{}), options...)
	if err != nil {
		return result.WithError(err)
	}
	second, err := actor.Query(ctx, keystone.Type(querySample{}), options...)
	if err != nil {
		return result.WithError(err)
	}
	if first.ResultSet == "" || second.ResultSet == "" || first.ResultSet == second.ResultSet || first.NextCursor == "" {
		return result.WithError(fmt.Errorf("expected independent saved queries with cursors"))
	}
	cases := []struct {
		name    string
		token   string
		options []keystone.QueryOption
		code    codes.Code
	}{
		{"unknown token", "missing-" + d.runID, nil, codes.NotFound},
		{"filter", first.ResultSet, []keystone.QueryOption{keystone.WhereEquals("run", d.runID)}, codes.InvalidArgument},
		{"projection", first.ResultSet, []keystone.QueryOption{keystone.Select("label")}, codes.InvalidArgument},
		{"sort", first.ResultSet, []keystone.QueryOption{keystone.SortAsc("position")}, codes.InvalidArgument},
		{"save again", first.ResultSet, []keystone.QueryOption{keystone.SaveResultSet()}, codes.InvalidArgument},
		{"negative limit", first.ResultSet, []keystone.QueryOption{keystone.Limit(-1)}, codes.InvalidArgument},
		{"large limit", first.ResultSet, []keystone.QueryOption{keystone.Limit(1001)}, codes.InvalidArgument},
		{"corrupt cursor", first.ResultSet, []keystone.QueryOption{keystone.Cursor("corrupted")}, codes.InvalidArgument},
		{"other token cursor", second.ResultSet, []keystone.QueryOption{keystone.Cursor(first.NextCursor)}, codes.InvalidArgument},
	}
	for _, test := range cases {
		opts := append([]keystone.QueryOption{keystone.ResultSet(test.token)}, test.options...)
		_, err := actor.Query(ctx, "", opts...)
		if status.Code(err) != test.code {
			return result.WithError(fmt.Errorf("%s: expected %s, got %v", test.name, test.code, err))
		}
	}
	for _, scope := range []struct{ workspace, user string }{
		{actor.WorkspaceID() + "-other", actor.UserID()},
		{actor.WorkspaceID(), actor.UserID() + "-other"},
	} {
		other := actor.Connection().Actor(scope.workspace, actor.RemoteIP(), scope.user, actor.UserAgent())
		other.SetClientName(actor.Client())
		_, err := other.Query(ctx, "", keystone.ResultSet(first.ResultSet))
		if status.Code(err) != codes.NotFound {
			return result.WithError(fmt.Errorf("saved token outside original authorization: expected NotFound, got %v", err))
		}
	}
	// A rejected read must leave the saved token usable.
	if _, err := actor.Query(ctx, "", keystone.ResultSet(first.ResultSet), keystone.Cursor(first.NextCursor)); err != nil {
		return result.WithError(err)
	}
	_, err = actor.Query(ctx, keystone.Type(querySample{}), keystone.WhereEquals("run", d.runID+"invalid"), keystone.SaveResultSet(), keystone.Cursor(first.NextCursor))
	if status.Code(err) != codes.InvalidArgument {
		return result.WithError(fmt.Errorf("save with cursor: expected InvalidArgument, got %v", err))
	}
	return result
}
