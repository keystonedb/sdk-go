package sensor

import (
	"context"
	"os"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

type Requirement struct {
	fixture sensorFixture
}

func (d *Requirement) Name() string {
	return "Sensor"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	os.Setenv("KS_SENSOR_TEST_ORIGIN", "https://l.kscdn.co/r/_sensors")
	conn.RegisterTypes(models.Person{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor, report requirements.Reporter) {
	report(d.measure(actor))
	d.verifyQueries(actor, func(result requirements.TestResult) { result.Name = "Entity " + result.Name; report(result) })
	app := &Requirement{fixture: sensorFixture{appScope: true}}
	app.verifyQueries(actor, func(result requirements.TestResult) { result.Name = "App " + result.Name; report(result) })
	d.verifyTTL(actor, report)
}

func (d *Requirement) verifyQueries(actor *keystone.Actor, report requirements.Reporter) {
	stored := d.storeQueryMeasurements(actor)
	report(stored)
	if stored.Error != nil {
		return
	}
	report(d.chart(actor, false))
	report(d.chart(actor, true))
	report(d.heatmap(actor, proto.SensorHeatmapLayout_SensorHeatmapDay, false))
	report(d.heatmap(actor, proto.SensorHeatmapLayout_SensorHeatmapWeek, true))
	report(d.activity(actor, proto.SensorActivityCalendar_SensorActivityWeek))
	report(d.activity(actor, proto.SensorActivityCalendar_SensorActivityMonth))
	report(d.predict(actor))
	report(d.predictWithoutBackfill(actor))
	if d.fixture.appScope {
		report(d.listAppSensors(actor))
	}
}

func (d *Requirement) measure(actor *keystone.Actor) requirements.TestResult {
	psn := &models.Person{}
	psn.Name = "John Smith"
	psn.AddSensorMeasurement("page.view", 1)
	for range 100 {
		psn.AddSensorMeasurement("mouse.move", float64(time.Now().Second()))
		time.Sleep(time.Millisecond)

	}
	psn.AddSensorMeasurementWithData("order", 100.12, map[string]string{
		"currency": "GBP",
	})
	updateErr := actor.Mutate(context.Background(), psn, keystone.WithMutationComment("Write Measurement"))

	return requirements.TestResult{
		Name:  "Store Measurement",
		Error: updateErr,
	}
}
