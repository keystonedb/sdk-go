package group_count

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/packaged/logger/v3/logger"
	"go.uber.org/zap"
)

const (
	configTypeOne   = "one"
	configTypeTwo   = "two"
	configTypeThree = "three"
)

type Requirement struct {
	// UseQuery runs the same group-count assertions against Query aggregate rows.
	UseQuery bool
	testID   string
}

func (d *Requirement) Name() string {
	return "Group Count"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Config{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	d.testID = uuid.NewString()
	report(d.create(actor))
	report(d.groupCount(actor))
	if d.UseQuery {
		report(d.groupHaving(actor))
	}
}

func (d *Requirement) create(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Create"}

	p1 := &models.Config{ConfigType: configTypeOne, Name: "Config One", TestID: d.testID}
	p2 := &models.Config{ConfigType: configTypeTwo, Name: "Config Two", TestID: d.testID}
	p2b := &models.Config{ConfigType: configTypeTwo, Name: "Config Two b", TestID: d.testID}
	p3 := &models.Config{ConfigType: configTypeThree, Name: "Config Three", TestID: d.testID}
	p3b := &models.Config{ConfigType: configTypeThree, Name: "Config Three b", TestID: d.testID}
	p3c := &models.Config{ConfigType: configTypeThree, Name: "Config Three c", TestID: d.testID}

	if err := actor.Mutate(context.Background(), p1); err != nil {
		return res.WithError(err)
	}
	if err := actor.Mutate(context.Background(), p2); err != nil {
		return res.WithError(err)
	}
	if err := actor.Mutate(context.Background(), p2b); err != nil {
		return res.WithError(err)
	}
	if err := actor.Mutate(context.Background(), p3); err != nil {
		return res.WithError(err)
	}
	if err := actor.Mutate(context.Background(), p3b); err != nil {
		return res.WithError(err)
	}
	if err := actor.Mutate(context.Background(), p3c); err != nil {
		return res.WithError(err)
	}

	return res
}

func (d *Requirement) groupCount(actor *keystone.Actor) requirements.TestResult {
	res := requirements.TestResult{Name: "Group Count"}

	grouped, err := d.count(actor)
	if err != nil {
		return res.WithError(err)
	}

	if len(grouped) != 3 {
		return res.WithError(fmt.Errorf("expected 3 groups, got %d", len(grouped)))
	}

	if grouped[configTypeOne].GetCount() != 1 {
		return res.WithError(fmt.Errorf("expected 1 for %s, got %d", configTypeOne, grouped[configTypeOne].GetCount()))
	}
	if grouped[configTypeTwo].GetCount() != 2 {
		return res.WithError(fmt.Errorf("expected 2 for %s, got %d", configTypeTwo, grouped[configTypeTwo].GetCount()))
	}
	if grouped[configTypeThree].GetCount() != 3 {
		return res.WithError(fmt.Errorf("expected 3 for %s, got %d", configTypeThree, grouped[configTypeThree].GetCount()))
	}

	if grouped["noConfig"].GetCount() != 0 {
		return res.WithError(fmt.Errorf("expected 0 for noConfig, got %d", grouped["noConfig"].GetCount()))
	}

	logger.I().Info("Grouped Count", zap.Any("grouped", grouped))

	return res
}

func (d *Requirement) count(actor *keystone.Actor) (map[string]*proto.GroupCountResponse_Result, error) {
	if !d.UseQuery {
		return actor.GroupCount(context.Background(), keystone.Type(models.Config{}), []string{"config_type"}, keystone.WhereEquals("test_id", d.testID))
	}
	resp, err := actor.QueryWithRequest(context.Background(), keystone.Type(models.Config{}), &proto.QueryRequest{
		GroupBy:    []string{"config_type"},
		Aggregates: []*proto.QueryAggregate{{Function: proto.QueryAggregate_COUNT, As: "count"}},
	}, keystone.WhereEquals("test_id", d.testID))
	if err != nil {
		return nil, err
	}
	grouped := make(map[string]*proto.GroupCountResponse_Result)
	for _, row := range resp.GetRows() {
		key := row.GetValues()["config_type"].GetText()
		if _, exists := grouped[key]; exists {
			return nil, fmt.Errorf("duplicate group %q", key)
		}
		grouped[key] = &proto.GroupCountResponse_Result{Key: key, Count: int32(row.GetValues()["count"].GetInt())}
	}
	return grouped, nil
}

func (d *Requirement) groupHaving(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Query grouped HAVING, order, total and cursor"}
	request := &proto.QueryRequest{
		GroupBy:    []string{"config_type"},
		Aggregates: []*proto.QueryAggregate{{Function: proto.QueryAggregate_COUNT, As: "count"}},
		Having:     []*proto.PropertyFilter{{Property: "count", Operator: proto.Operator_GreaterThan, Values: []*proto.Value{{Int: 1}}}},
		OrderBy:    []*proto.PropertySort{{Property: "count", Descending: true}},
		Limit:      1, IncludeTotal: true,
	}
	for i, key := range []string{configTypeThree, configTypeTwo} {
		response, err := actor.QueryWithRequest(context.Background(), keystone.Type(models.Config{}), request, keystone.WhereEquals("test_id", d.testID))
		if err != nil {
			return result.WithError(err)
		}
		if len(response.Rows) != 1 || len(response.Entities) != 0 || response.TotalResults == nil || response.GetTotalResults() != 2 {
			return result.WithError(fmt.Errorf("expected one aggregate row and total 2: %v", response))
		}
		values := response.Rows[0].GetValues()
		if values["config_type"].GetText() != key || values["count"].GetInt() != int64(3-i) {
			return result.WithError(fmt.Errorf("unexpected aggregate values: %v", values))
		}
		if response.HasMore != (i == 0) {
			return result.WithError(fmt.Errorf("unexpected has_more on group page %d", i))
		}
		if response.HasMore && response.NextCursor == "" {
			return result.WithError(fmt.Errorf("missing group cursor"))
		}
		request.Cursor = response.NextCursor
	}
	return result
}
