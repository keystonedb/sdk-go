package aggregates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

// Requirement exercises aggregates through collections without summary structs.
type Requirement struct{ runID string }

func (*Requirement) Name() string { return "Aggregates (collections)" }

func (d *Requirement) Register(conn *keystone.Connection) error {
	d.runID = uuid.NewString()
	conn.RegisterTypes(models.AggregateSample{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	report(d.functions(actor))
	report(d.groups(actor, false))
	report(d.groups(actor, true))
	report(d.empty(actor))
	report(d.reset(actor))
}

func (d *Requirement) fixtures(ctx context.Context, actor *keystone.Actor, scope string) ([]*models.AggregateSample, error) {
	samples := make([]*models.AggregateSample, 0, 6)
	for i, amount := range []int64{2, 2, 8, 4, 10, 6} {
		sample := &models.AggregateSample{Run: d.runID + scope, Category: []string{"a", "a", "a", "b", "b", "c"}[i], Position: int64(i + 1), Amount: amount, Weight: float64(amount) * 0.75}
		if err := actor.Mutate(ctx, sample); err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}
	return samples, nil
}

func (d *Requirement) functions(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Collection count, distinct, sum, min, max, average and float aggregates"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := d.fixtures(ctx, actor, "functions"); err != nil {
		return result.WithError(err)
	}
	source := keystone.Collection(models.AggregateSample{}, actor)
	rows, err := source.Aggregate(ctx, keystone.WhereEquals("run", d.runID+"functions"),
		keystone.Count("count"), keystone.CountProperty("amount", "present"), keystone.CountDistinct("amount", "distinct_count"),
		keystone.Sum("amount", "sum"), keystone.Min("amount", "min"), keystone.Max("amount", "max"), keystone.Avg("amount", "avg"),
		keystone.Sum("weight", "float_sum"), keystone.Avg("weight", "float_avg"),
		keystone.Aggregate(proto.QueryAggregate_SUM, "amount", "distinct_sum", true), keystone.Explain(), keystone.IncludeTotal())
	if err != nil {
		return result.WithError(err)
	}
	if len(rows.Items()) != 1 || rows.HasMore() || rows.Response().TotalResults == nil || rows.Response().GetTotalResults() != 1 || rows.Response().Execution == nil {
		return result.WithError(fmt.Errorf("expected one aggregate row, total and execution details: %v", rows.Response()))
	}
	count := 0
	for row := range rows.All() {
		count++
		for alias, expected := range map[string]int64{"count": 6, "present": 6, "distinct_count": 5, "sum": 32, "min": 2, "max": 10, "distinct_sum": 30} {
			if row.Value(alias) == nil || row.Value(alias).IsNull || row.Value(alias).GetInt() != expected {
				return result.WithError(fmt.Errorf("%s: expected %d, got %v", alias, expected, row.Value(alias)))
			}
		}
		for alias, expected := range map[string]float64{"avg": 32.0 / 6, "float_sum": 24, "float_avg": 4} {
			if row.Value(alias) == nil || row.Value(alias).IsNull || math.Abs(row.Value(alias).GetFloat()-expected) > 1e-9 {
				return result.WithError(fmt.Errorf("%s: expected %v, got %v", alias, expected, row.Value(alias)))
			}
		}
		if row.Value("missing") != nil {
			return result.WithError(fmt.Errorf("missing alias should return nil"))
		}
	}
	if count != 1 || len(source.Items()) != 0 || source.Response() != nil {
		return result.WithError(fmt.Errorf("aggregate iteration or source collection state changed unexpectedly"))
	}
	return result
}

func (d *Requirement) groups(actor *keystone.Actor, saved bool) requirements.TestResult {
	result := requirements.TestResult{Name: fmt.Sprintf("Collection grouped HAVING, order, cursor and snapshot (saved=%t)", saved)}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	scope := fmt.Sprintf("groups-%t", saved)
	samples, err := d.fixtures(ctx, actor, scope)
	if err != nil {
		return result.WithError(err)
	}
	options := []keystone.QueryOption{keystone.WhereEquals("run", d.runID+scope), keystone.GroupBy("category"), keystone.Count("count"), keystone.Sum("amount", "sum"),
		keystone.Having(keystone.And(keystone.WhereGreaterThan("count", 1), keystone.Or(keystone.WhereEquals("category", "a"), keystone.WhereEquals("category", "b")))),
		keystone.SortDesc("sum"), keystone.Limit(1), keystone.IncludeTotal(), keystone.Explain()}
	if saved {
		options = append(options, keystone.SaveResultSet())
	}
	source := keystone.Collection((*models.AggregateSample)(nil), actor)
	rows, err := source.Aggregate(ctx, options...)
	if err != nil {
		return result.WithError(err)
	}
	token := rows.Response().ResultSet
	if saved {
		if token == "" || rows.Response().ResultSetExpires == nil {
			return result.WithError(fmt.Errorf("missing saved aggregate token or expiry"))
		}
		samples[0].Amount = 200
		if err := actor.Mutate(ctx, samples[0]); err != nil {
			return result.WithError(err)
		}
	}
	for page := 0; page < 2; page++ {
		response := rows.Response()
		if len(rows.Items()) != page+1 || len(response.Rows) != 1 || response.TotalResults == nil || response.GetTotalResults() != 2 || rows.HasMore() != (page == 0) {
			return result.WithError(fmt.Errorf("unexpected grouped collection page %d: %v", page, response))
		}
		row := rows.Items()[page]
		category, count, sum := "b", int64(2), int64(14)
		if page == 1 {
			category, count, sum = "a", 3, 12
		}
		if row.Value("category").GetText() != category || row.Value("count").GetInt() != count || row.Value("sum").GetInt() != sum {
			return result.WithError(fmt.Errorf("unexpected group %d: %v", page, row))
		}
		if saved && (response.ResultSet != token || (page == 1 && response.GetExecution().GetStrategy() != "result_set")) {
			return result.WithError(fmt.Errorf("saved group page did not reuse result set: %v", response))
		}
		if page == 0 {
			if rows.NextCursor() == "" {
				return result.WithError(fmt.Errorf("missing group cursor"))
			}
			if err := rows.Next(ctx); err != nil {
				return result.WithError(err)
			}
		}
	}
	count := 0
	for range rows.All() {
		count++
	}
	if count != 2 || !errors.Is(rows.Next(ctx), io.EOF) {
		return result.WithError(fmt.Errorf("group collection iteration or exhaustion failed"))
	}
	if saved {
		live, err := source.Aggregate(ctx, keystone.WhereEquals("run", d.runID+scope), keystone.WhereEquals("category", "a"), keystone.Sum("amount", "sum"))
		if err != nil {
			return result.WithError(err)
		}
		if len(live.Items()) != 1 || live.Items()[0].Value("sum").GetInt() != 210 {
			return result.WithError(fmt.Errorf("live aggregate did not reflect mutation: %v", live.Items()))
		}
		// A fresh collection can reopen the same snapshot without result structs.
		replay := keystone.AggregateCollection(keystone.Type(models.AggregateSample{}), actor)
		if err := replay.Query(ctx, keystone.ResultSet(token), keystone.Limit(10), keystone.IncludeTotal()); err != nil {
			return result.WithError(err)
		}
		if len(replay.Items()) != 2 || replay.Items()[1].Value("sum").GetInt() != 12 || replay.Response().GetTotalResults() != 2 || replay.HasMore() {
			return result.WithError(fmt.Errorf("saved aggregate replay changed: %v", replay.Items()))
		}
	}
	return result
}

func (d *Requirement) empty(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Collection empty aggregates, SQL NULL and saved empty groups"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	source := keystone.Collection(models.AggregateSample{}, actor)
	rows, err := source.Aggregate(ctx, keystone.WithEntityIDs(nil), keystone.Count("count"), keystone.Sum("amount", "sum"), keystone.Avg("amount", "avg"), keystone.Min("amount", "min"), keystone.Max("amount", "max"))
	if err != nil {
		return result.WithError(err)
	}
	if len(rows.Items()) != 1 || rows.Items()[0].Value("count") == nil || rows.Items()[0].Value("count").GetInt() != 0 {
		return result.WithError(fmt.Errorf("expected one empty aggregate with zero count: %v", rows.Items()))
	}
	for _, alias := range []string{"sum", "avg", "min", "max"} {
		if value := rows.Items()[0].Value(alias); value == nil || !value.IsNull {
			return result.WithError(fmt.Errorf("empty %s must remain SQL NULL: %v", alias, value))
		}
	}
	if err := rows.Query(ctx, keystone.WithEntityIDs(nil), keystone.GroupBy("category"), keystone.Count("count"), keystone.SaveResultSet(), keystone.IncludeTotal()); err != nil {
		return result.WithError(err)
	}
	if len(rows.Items()) != 0 || rows.HasMore() || rows.NextCursor() != "" || rows.Response().ResultSet == "" || rows.Response().TotalResults == nil || rows.Response().GetTotalResults() != 0 || !errors.Is(rows.Next(ctx), io.EOF) {
		return result.WithError(fmt.Errorf("unexpected saved empty groups: %v", rows.Response()))
	}
	for range rows.All() {
		return result.WithError(fmt.Errorf("empty collection yielded a row"))
	}
	return result
}

func (d *Requirement) reset(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Collection replaces results and preserves state on failed aggregate query"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := d.fixtures(ctx, actor, "reset"); err != nil {
		return result.WithError(err)
	}
	rows := keystone.AggregateCollection(keystone.Type(models.AggregateSample{}), actor)
	options := []keystone.QueryOption{keystone.WhereEquals("run", d.runID+"reset"), keystone.GroupBy("category"), keystone.Count("count"), keystone.SortAsc("category"), keystone.Limit(1)}
	if err := rows.Query(ctx, options...); err != nil {
		return result.WithError(err)
	}
	original, cursor := rows.Response(), rows.NextCursor()
	if err := rows.Query(ctx, keystone.Count("invalid alias!")); err == nil {
		return result.WithError(fmt.Errorf("invalid aggregate alias was accepted"))
	}
	if rows.Response() != original || rows.NextCursor() != cursor || len(rows.Items()) != 1 {
		return result.WithError(fmt.Errorf("failed query changed collection state"))
	}
	if err := rows.Next(ctx); err != nil {
		return result.WithError(err)
	}
	if len(rows.Items()) != 2 || rows.Items()[1].Value("category").GetText() != "b" {
		return result.WithError(fmt.Errorf("failed query changed continuation: %v", rows.Items()))
	}
	if err := rows.Query(ctx, keystone.WhereEquals("run", d.runID+"reset"), keystone.Count("count")); err != nil {
		return result.WithError(err)
	}
	if len(rows.Items()) != 1 || rows.Items()[0].Value("count").GetInt() != 6 || rows.HasMore() {
		return result.WithError(fmt.Errorf("new query did not replace prior rows: %v", rows.Items()))
	}
	return result
}
