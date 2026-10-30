package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SensorMeasurementOption configures a measurement for WriteSensors.
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

// NewSensorMeasurement creates a measurement for the WriteSensors RPC.
// Its event time defaults to now; no entity mutation is involved.
func NewSensorMeasurement(sensor string, value float64, options ...SensorMeasurementOption) *proto.EntitySensorMeasurement {
	measurement := &proto.EntitySensorMeasurement{Sensor: sensor, Value: value, At: timestamppb.Now()}
	for _, option := range options {
		option(measurement)
	}
	return measurement
}

// WithSensorData attaches metadata to a measurement.
func WithSensorData(data map[string]string) SensorMeasurementOption {
	return func(m *proto.EntitySensorMeasurement) { m.Data = data }
}
