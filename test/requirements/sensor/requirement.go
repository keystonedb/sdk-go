package sensor

import (
	"context"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"time"
)

type Requirement struct {
}

func (d *Requirement) Name() string {
	return "Sensor"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	conn.RegisterTypes(models.User{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.measure(actor),
	}
}

func (d *Requirement) measure(actor *keystone.Actor) requirements.TestResult {
	psn := &models.Person{}
	psn.Name = "John Smith"
	psn.AddSensorMeasurement("page.view", 1)
	for i := 0; i < 100; i++ {
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
