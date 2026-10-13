package keystone

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSensorMeasurementOptions(t *testing.T) {
	var sensors EmbeddedSensors
	at := timestamppb.New(time.Unix(1700000000, 0))
	sensors.AddSensorMeasurement("legacy", 1)
	sensors.AddSensorMeasurementWithData("with-data", 2, map[string]string{"currency": "GBP"},
		WithSensorTTL(60), WithSensorMeasurementID("retry-id"), WithSensorTimestamp(at), WithSensorBackfill(true))
	measurements := sensors.GetSensorMeasurements()
	require.Len(t, measurements, 2)
	require.Zero(t, measurements[0].GetTtlSeconds())
	require.NotNil(t, measurements[0].GetAt())
	require.Equal(t, "with-data", measurements[1].GetSensor())
	require.Equal(t, float64(2), measurements[1].GetValue())
	require.Equal(t, int32(60), measurements[1].GetTtlSeconds())
	require.Equal(t, "retry-id", measurements[1].GetMeasurementId())
	require.Equal(t, at, measurements[1].GetAt())
	require.True(t, measurements[1].GetBackfill())
	require.Equal(t, "GBP", measurements[1].GetData()["currency"])
	sensors.AddSensorMeasurement("ttl-only", 3, WithSensorTTL(10))
	require.Equal(t, int32(10), sensors.GetSensorMeasurements()[2].GetTtlSeconds())
	require.NoError(t, sensors.ClearSensorMeasurements())
	require.Empty(t, sensors.GetSensorMeasurements())
}
