package sensor

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (d *Requirement) listAppSensors(actor *keystone.Actor) requirements.TestResult {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	response, err := actor.ListAppSensors(ctx)
	if err == nil {
		for _, sensor := range []string{querySensor, predictionSensor} {
			if !slices.Contains(response.GetSensors(), d.sensorName(sensor)) {
				err = fmt.Errorf("ListAppSensors did not discover %q", d.sensorName(sensor))
				break
			}
		}
	}
	return requirements.NewResult("ListAppSensors", err)
}

func (d *Requirement) verifyTTL(actor *keystone.Actor, report requirements.Reporter) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	name := "integration.ttl." + uuid.NewString()
	at := time.Now().UTC().Truncate(time.Minute).Add(-time.Minute)
	measurement := func(id string, value float64, ttl int32) *proto.EntitySensorMeasurement {
		return &proto.EntitySensorMeasurement{Sensor: name, At: timestamppb.New(at),
			MeasurementId: id, Value: value, TtlSeconds: ttl}
	}
	write := func(request *proto.WriteSensorsRequest) error {
		response, err := actor.WriteSensors(ctx, request)
		if err == nil && !response.GetSuccess() {
			err = fmt.Errorf("WriteSensors did not succeed")
		}
		return err
	}
	// Zero TTL retains a measurement normally; a per-measurement TTL overrides
	// the request default. Every event is older than the short receipt-based TTL.
	err := write(&proto.WriteSensorsRequest{AppScope: true, Measurements: []*proto.EntitySensorMeasurement{measurement("retained", 2, 0)}})
	if err == nil {
		err = write(&proto.WriteSensorsRequest{AppScope: true, TtlSeconds: 5,
			Measurements: []*proto.EntitySensorMeasurement{measurement("default", 1, 0), measurement("override", 3, 60)}})
	}
	report(requirements.NewResult("WriteSensors (TTL default and override)", err))
	if err != nil {
		return
	}
	read := func() (float64, int64, error) {
		response, err := actor.ChartSensor(ctx, &proto.ChartSensorRequest{AppScope: true, Sensor: name,
			From: timestamppb.New(at), Until: timestamppb.New(at.Add(time.Minute)), IntervalSeconds: 60})
		var sum float64
		var samples int64
		for _, bucket := range response.GetBuckets() {
			sum += bucket.GetSum()
			samples += bucket.GetSamples()
		}
		return sum, samples, err
	}
	sum, samples, err := read()
	if err == nil && (!closeEnough(sum, 6) || samples != 3) {
		err = fmt.Errorf("before TTL expiry: got sum=%g samples=%d; want sum=6 samples=3", sum, samples)
	}
	report(requirements.NewResult("Sensor TTL begins at receipt, not event time", err))
	if err != nil {
		return
	}
	// Invalid TTLs must reject the whole batch, including the valid first entry.
	for _, request := range []*proto.WriteSensorsRequest{
		{AppScope: true, TtlSeconds: -1, Measurements: []*proto.EntitySensorMeasurement{measurement("bad-request", 100, 0)}},
		{AppScope: true, Measurements: []*proto.EntitySensorMeasurement{measurement("valid-in-invalid-batch", 100, 0), measurement("bad-measurement", 100, -1)}},
	} {
		_, err := actor.WriteSensors(ctx, request)
		if status.Code(err) != codes.InvalidArgument {
			report(requirements.NewResult("WriteSensors rejects negative TTLs atomically", fmt.Errorf("want InvalidArgument; got %v", err)))
			return
		}
	}
	report(requirements.NewResult("WriteSensors rejects negative TTLs", nil))
	// Poll boundedly rather than relying on a fixed sleep or expiry worker timing.
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(20 * time.Second)
	defer deadline.Stop()
	for {
		sum, samples, err = read()
		if err != nil {
			break
		}
		if closeEnough(sum, 5) && samples == 2 {
			break
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-deadline.C:
			err = fmt.Errorf("after TTL expiry: got sum=%g samples=%d; want sum=5 samples=2 (override and retained entries)", sum, samples)
		case <-ticker.C:
			continue
		}
		break
	}
	report(requirements.NewResult("Sensor TTL expires chart contributions and preserves overrides", err))
	if err != nil {
		return
	}
	activity, err := actor.SensorActivity(ctx, &proto.SensorActivityRequest{AppScope: true, Sensor: name, LookbackDays: 7})
	if err == nil {
		sum, samples = 0, 0
		for _, slot := range activity.GetSlots() {
			sum += slot.GetSum()
			samples += slot.GetSamples()
		}
		if !closeEnough(sum, 5) || samples != 2 {
			err = fmt.Errorf("expired or invalid measurements remain in activity: sum=%g samples=%d", sum, samples)
		}
	}
	report(requirements.NewResult("Sensor TTL expires activity contributions", err))
}
