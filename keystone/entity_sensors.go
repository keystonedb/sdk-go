package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SensorProvider is an interface for entities that can provide Sensors
type SensorProvider interface {
	ClearSensorMeasurements() error
	GetSensorMeasurements() []*proto.EntitySensorMeasurement
}

// EmbeddedSensors is a struct that implements SensorProvider
type EmbeddedSensors struct {
	ksEntitySensorsMeasurements []*proto.EntitySensorMeasurement
}

// ClearSensorMeasurements clears the Sensors
func (e *EmbeddedSensors) ClearSensorMeasurements() error {
	e.ksEntitySensorsMeasurements = []*proto.EntitySensorMeasurement{}
	return nil
}

// GetSensorMeasurements returns the Sensors measurements
func (e *EmbeddedSensors) GetSensorMeasurements() []*proto.EntitySensorMeasurement {
	return e.ksEntitySensorsMeasurements
}

// AddSensorMeasurement adds a Sensor measurement
func (e *EmbeddedSensors) AddSensorMeasurement(sensor string, value float64) {
	e.ksEntitySensorsMeasurements = append(e.ksEntitySensorsMeasurements, &proto.EntitySensorMeasurement{
		Sensor: sensor,
		Value:  value,
		At:     timestamppb.Now(),
	})
}

// AddSensorMeasurementWithData adds a Sensor measurement
func (e *EmbeddedSensors) AddSensorMeasurementWithData(sensor string, value float64, data map[string]string) {
	e.ksEntitySensorsMeasurements = append(e.ksEntitySensorsMeasurements, &proto.EntitySensorMeasurement{
		Sensor: sensor,
		Value:  value,
		At:     timestamppb.Now(),
		Data:   data,
	})
}
