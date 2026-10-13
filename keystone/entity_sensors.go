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

// SensorMeasurementOption configures a sensor measurement before it is queued.
type SensorMeasurementOption func(*proto.EntitySensorMeasurement)

// WithSensorTTL expires the measurement and its summaries after this many seconds
// from server receipt. Zero uses normal retention (or a WriteSensors request TTL).
func WithSensorTTL(seconds int32) SensorMeasurementOption {
	return func(m *proto.EntitySensorMeasurement) { m.TtlSeconds = seconds }
}

// WithSensorMeasurementID sets the stable ID used to deduplicate retries.
func WithSensorMeasurementID(id string) SensorMeasurementOption {
	return func(m *proto.EntitySensorMeasurement) { m.MeasurementId = id }
}

// WithSensorTimestamp sets the measurement's event time.
func WithSensorTimestamp(at *timestamppb.Timestamp) SensorMeasurementOption {
	return func(m *proto.EntitySensorMeasurement) { m.At = at }
}

// WithSensorBackfill marks a historical measurement that should not update live activity.
func WithSensorBackfill(backfill bool) SensorMeasurementOption {
	return func(m *proto.EntitySensorMeasurement) { m.Backfill = backfill }
}

// AddSensorMeasurement queues a measurement with an optional TTL and other metadata.
func (e *EmbeddedSensors) AddSensorMeasurement(sensor string, value float64, options ...SensorMeasurementOption) {
	e.AddSensorMeasurementWithData(sensor, value, nil, options...)
}

// AddSensorMeasurementWithData queues a measurement with data and optional metadata.
func (e *EmbeddedSensors) AddSensorMeasurementWithData(sensor string, value float64, data map[string]string, options ...SensorMeasurementOption) {
	measurement := &proto.EntitySensorMeasurement{Sensor: sensor, Value: value, At: timestamppb.Now(), Data: data}
	for _, option := range options {
		option(measurement)
	}
	e.ksEntitySensorsMeasurements = append(e.ksEntitySensorsMeasurements, measurement)
}
