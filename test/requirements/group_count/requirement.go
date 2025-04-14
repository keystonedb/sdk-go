package group_count

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
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
	testID string
}

func (d *Requirement) Name() string {
	return "Group Count"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.Config{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	d.testID = uuid.NewString()
	return []requirements.TestResult{
		d.create(actor),
		d.groupCount(actor),
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

	grouped, err := actor.GroupCount(context.Background(), keystone.Type(models.Config{}), []string{"config_type"}, keystone.WhereEquals("test_id", d.testID))
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
