package sensor

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	querySensor      = "integration.query"
	predictionSensor = "integration.prediction"
	queryTimeout     = 30 * time.Second
)

type sensorFixture struct {
	appScope              bool
	suffix                string
	entityID              string
	from, until, forecast time.Time
}

func (d *Requirement) schema() *proto.Key {
	if d.fixture.appScope {
		return nil
	}
	return &proto.Key{Key: keystone.Type(models.Person{})}
}

func (d *Requirement) sensorName(name string) string { return name + d.fixture.suffix }

func (d *Requirement) storeQueryMeasurements(actor *keystone.Actor) requirements.TestResult {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	now := time.Now().UTC().Truncate(time.Minute)
	from := now.Truncate(15 * time.Minute).Add(-15 * time.Minute)
	d.fixture = sensorFixture{appScope: d.fixture.appScope, suffix: "." + uuid.NewString(),
		from:     from,
		until:    from.Add(15 * time.Minute),
		forecast: now.Truncate(time.Hour).Add(time.Hour),
	}
	var measurements []*proto.EntitySensorMeasurement
	add := func(sensor string, at time.Time, value float64, backfill bool) {
		measurements = append(measurements, keystone.NewSensorMeasurement(d.sensorName(sensor), value, keystone.WithSensorTimestamp(timestamppb.New(at)), keystone.WithSensorBackfill(backfill), keystone.WithSensorMeasurementID(uuid.NewString())))
	}
	// Use one completed UTC quarter-hour so weekly heatmap aggregation
	// does not clip a partial boundary interval.
	add(querySensor, from.Add(time.Minute), 2, false)
	add(querySensor, from.Add(2*time.Minute), 3, false)
	add(querySensor, from.Add(3*time.Minute), 7, true)
	// Two complete historical intervals at the forecast's UTC hour provide
	// deterministic daily seasonal peers without treating missing data as zero.
	for days := 1; days <= 2; days++ {
		add(predictionSensor, d.fixture.forecast.AddDate(0, 0, -days).Add(time.Minute), 10, true)
	}
	var err error
	if !d.fixture.appScope {
		// Direct writes require a registered schema and stable ID, but do not mutate the entity.
		entity := &models.Person{Name: "Sensor Query Integration"}
		err = actor.Mutate(ctx, entity, keystone.WithMutationComment("Sensor query integration fixture"))
		if err != nil {
			return requirements.NewResult("Create Sensor Entity", err)
		}
		d.fixture.entityID = entity.GetKeystoneID().String()
		if d.fixture.entityID == "" {
			return requirements.NewResult("Create Sensor Entity", fmt.Errorf("mutation returned no entity ID"))
		}
	}
	request := &proto.WriteSensorsRequest{
		Schema: d.schema(), EntityId: d.fixture.entityID, AppScope: d.fixture.appScope,
		Measurements: measurements,
	}
	response, err := actor.WriteSensors(ctx, request)
	if err == nil && !response.GetSuccess() {
		err = fmt.Errorf("WriteSensors did not succeed")
	}

	if err == nil {
		response, err = actor.WriteSensors(ctx, request)
		if err == nil && !response.GetSuccess() {
			err = fmt.Errorf("WriteSensors retry did not succeed")
		}
	}
	return requirements.NewResult("WriteSensors (measurements and idempotent retry)", err)
}

func (d *Requirement) chart(actor *keystone.Actor, excludeBackfill bool) requirements.TestResult {
	result := requirements.NewResult(fmt.Sprintf("ChartSensor (exclude backfill=%t)", excludeBackfill), nil)
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	response, err := actor.ChartSensor(ctx, &proto.ChartSensorRequest{
		Schema: d.schema(), EntityId: d.fixture.entityID, Sensor: d.sensorName(querySensor), AppScope: d.fixture.appScope,
		From: timestamppb.New(d.fixture.from), Until: timestamppb.New(d.fixture.until),
		IntervalSeconds: 60, ExcludeBackfill: excludeBackfill,
	})
	if err != nil {
		return result.WithError(err)
	}
	var sum float64
	var samples int64
	for _, bucket := range response.GetBuckets() {
		if bucket.GetAt() == nil || bucket.GetAt().CheckValid() != nil {
			return result.WithError(fmt.Errorf("chart bucket has an invalid timestamp"))
		}
		sum += bucket.GetSum()
		samples += bucket.GetSamples()
	}
	wantSum, wantSamples := expectedTotals(excludeBackfill)
	if !closeEnough(sum, wantSum) || samples != wantSamples {
		return result.WithError(fmt.Errorf("got sum=%g samples=%d; want sum=%g samples=%d", sum, samples, wantSum, wantSamples))
	}
	if response.GetResolutionSeconds() == 0 {
		return result.WithError(fmt.Errorf("chart returned no resolution"))
	}
	return result
}

func (d *Requirement) heatmap(actor *keystone.Actor, layout proto.SensorHeatmapLayout, excludeBackfill bool) requirements.TestResult {
	result := requirements.NewResult(fmt.Sprintf("HeatmapSensor (%s, exclude backfill=%t)", layout, excludeBackfill), nil)
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	response, err := actor.HeatmapSensor(ctx, &proto.HeatmapSensorRequest{
		Schema: d.schema(), EntityId: d.fixture.entityID, Sensor: d.sensorName(querySensor), AppScope: d.fixture.appScope,
		From: timestamppb.New(d.fixture.from), Until: timestamppb.New(d.fixture.until),
		Layout: layout, Timezone: "UTC", ExcludeBackfill: excludeBackfill,
	})
	if err != nil {
		return result.WithError(err)
	}
	wantCells := 1440
	if layout == proto.SensorHeatmapLayout_SensorHeatmapWeek {
		wantCells = 672
	}
	if len(response.GetCells()) != wantCells || response.GetTimezone() != "UTC" {
		return result.WithError(fmt.Errorf("got %d cells timezone=%q; want %d cells in UTC", len(response.GetCells()), response.GetTimezone(), wantCells))
	}
	var sum, samples float64
	for _, cell := range response.GetCells() {
		sum += cell.GetSum()
		samples += cell.GetSamples()
	}
	wantSum, wantSamples := expectedTotals(excludeBackfill)
	if !closeEnough(sum, wantSum) || !closeEnough(samples, float64(wantSamples)) {
		return result.WithError(fmt.Errorf("got sum=%g samples=%g; want sum=%g samples=%d", sum, samples, wantSum, wantSamples))
	}
	return result
}

func (d *Requirement) activity(actor *keystone.Actor, calendar proto.SensorActivityCalendar) requirements.TestResult {
	result := requirements.NewResult(fmt.Sprintf("SensorActivity (%s)", calendar), nil)
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	response, err := actor.SensorActivity(ctx, &proto.SensorActivityRequest{
		Schema: d.schema(), EntityId: d.fixture.entityID, Sensor: d.sensorName(querySensor), AppScope: d.fixture.appScope,
		Calendar: calendar, LookbackDays: 7,
	})
	if err != nil {
		return result.WithError(err)
	}
	wantSlots := 168
	if calendar == proto.SensorActivityCalendar_SensorActivityMonth {
		wantSlots = 744
	}
	if len(response.GetSlots()) != wantSlots || response.GetTimezone() != "UTC" {
		return result.WithError(fmt.Errorf("got %d slots timezone=%q; want %d slots in UTC", len(response.GetSlots()), response.GetTimezone(), wantSlots))
	}
	var sum float64
	var samples int64
	for _, slot := range response.GetSlots() {
		sum += slot.GetSum()
		samples += slot.GetSamples()
	}
	if !closeEnough(sum, 12) || samples != 3 {
		return result.WithError(fmt.Errorf("got sum=%g samples=%d; want sum=12 samples=3", sum, samples))
	}
	if len(response.GetRankedSlots()) == 0 {
		return result.WithError(fmt.Errorf("activity returned no ranked slots for recorded measurements"))
	}
	for _, index := range response.GetRankedSlots() {
		if int(index) >= len(response.GetSlots()) || response.GetSlots()[index].GetSamples() == 0 {
			return result.WithError(fmt.Errorf("invalid ranked activity slot %d", index))
		}
	}
	return result
}

func (d *Requirement) predict(actor *keystone.Actor) requirements.TestResult {
	result := requirements.NewResult("PredictSensor (historical daily peers)", nil)
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	response, err := actor.PredictSensor(ctx, &proto.PredictSensorRequest{
		Schema: d.schema(), EntityId: d.fixture.entityID, Sensor: d.sensorName(predictionSensor), AppScope: d.fixture.appScope,
		From: timestamppb.New(d.fixture.forecast), Until: timestamppb.New(d.fixture.forecast.Add(time.Hour)),
		IntervalSeconds: 3600, TrainingDays: 7, Timezone: "UTC",
		Metric:              proto.SensorPredictionMetric_SensorPredictionSum,
		MinimumObservations: 2, DailySeasonality: true, RangeCoverage: 0.9,
	})
	if err != nil {
		return result.WithError(err)
	}
	if len(response.GetPoints()) != 1 || response.GetResolutionSeconds() != 3600 || response.GetTimezone() != "UTC" {
		return result.WithError(fmt.Errorf("got %d points resolution=%d timezone=%q; want one hourly point in UTC", len(response.GetPoints()), response.GetResolutionSeconds(), response.GetTimezone()))
	}
	point := response.GetPoints()[0]
	if point.GetAt() == nil || point.GetAt().CheckValid() != nil || !point.GetAt().AsTime().Equal(d.fixture.forecast) {
		return result.WithError(fmt.Errorf("prediction timestamp does not match requested forecast"))
	}
	if !point.GetAvailable() || point.GetObservations() != 2 || !closeEnough(point.GetExpected(), 10) || !closeEnough(point.GetLower(), 10) || !closeEnough(point.GetUpper(), 10) {
		return result.WithError(fmt.Errorf("expected available prediction of 10 with range [10,10] from two observations; got %v", point))
	}
	if point.GetBaseline() != "daily" || point.GetTrainingSamples() != 2 {
		return result.WithError(fmt.Errorf("want daily baseline with two training samples; got baseline=%q samples=%d", point.GetBaseline(), point.GetTrainingSamples()))
	}
	if point.GetActualAvailable() || point.GetAnomaly() {
		return result.WithError(fmt.Errorf("future prediction unexpectedly has actual data or an anomaly"))
	}
	return result
}

func expectedTotals(excludeBackfill bool) (float64, int64) {
	if excludeBackfill {
		return 5, 2
	}
	return 12, 3
}

func closeEnough(got, want float64) bool {
	return !math.IsNaN(got) && math.Abs(got-want) < 1e-9
}

func (d *Requirement) predictWithoutBackfill(actor *keystone.Actor) requirements.TestResult {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	response, err := actor.PredictSensor(ctx, &proto.PredictSensorRequest{
		Schema: d.schema(), EntityId: d.fixture.entityID, AppScope: d.fixture.appScope,
		Sensor: d.sensorName(predictionSensor), From: timestamppb.New(d.fixture.forecast),
		Until: timestamppb.New(d.fixture.forecast.Add(time.Hour)), IntervalSeconds: 3600,
		TrainingDays: 7, MinimumObservations: 2, DailySeasonality: true, ExcludeBackfill: true,
	})
	if err == nil {
		if len(response.GetPoints()) != 1 {
			err = fmt.Errorf("want one learning prediction point; got %d", len(response.GetPoints()))
		} else if point := response.GetPoints()[0]; point.GetAvailable() || point.GetTrainingSamples() != 0 || point.GetBaseline() != "" {
			err = fmt.Errorf("excluded backfill supplied prediction evidence: %v", point)
		}
	}
	return requirements.NewResult("PredictSensor (excluded backfill remains learning)", err)
}
