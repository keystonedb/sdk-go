package keystone

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewSensorMeasurement(t *testing.T) {
	defaultMeasurement := NewSensorMeasurement("requests", 1)
	require.Equal(t, "requests", defaultMeasurement.GetSensor())
	require.Equal(t, float64(1), defaultMeasurement.GetValue())
	require.Zero(t, defaultMeasurement.GetTtlSeconds())
	require.NoError(t, defaultMeasurement.GetAt().CheckValid())
	require.WithinDuration(t, time.Now(), defaultMeasurement.GetAt().AsTime(), time.Second)

	at := timestamppb.New(time.Unix(1700000000, 0))
	measurement := NewSensorMeasurement("orders", 2, WithSensorTTL(60),
		WithSensorMeasurementID("retry-id"), WithSensorTimestamp(at), WithSensorBackfill(true),
		WithSensorData(map[string]string{"currency": "GBP"}))
	require.Equal(t, "orders", measurement.GetSensor())
	require.Equal(t, float64(2), measurement.GetValue())
	require.Equal(t, int32(60), measurement.GetTtlSeconds())
	require.Equal(t, "retry-id", measurement.GetMeasurementId())
	require.Equal(t, at, measurement.GetAt())
	require.True(t, measurement.GetBackfill())
	require.Equal(t, "GBP", measurement.GetData()["currency"])
	require.Zero(t, defaultMeasurement.GetTtlSeconds())
	require.Empty(t, defaultMeasurement.GetData())
}
