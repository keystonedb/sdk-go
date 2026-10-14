package sensor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/requirements"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (d *Requirement) verifyBackground(actor *keystone.Actor, report requirements.Reporter) {
	for _, appScope := range []bool{false, true} {
		err := d.backgroundWrite(actor, appScope)
		report(requirements.NewResult(fmt.Sprintf("WriteSensors background ingestion (app scope=%t)", appScope), err))
	}
}

func (d *Requirement) backgroundWrite(actor *keystone.Actor, appScope bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	name := "integration.background." + uuid.NewString()
	at := time.Now().UTC().Truncate(time.Minute).Add(-time.Minute)
	request := &proto.WriteSensorsRequest{AppScope: appScope, Measurements: []*proto.EntitySensorMeasurement{
		keystone.NewSensorMeasurement(name, 5, keystone.WithSensorTimestamp(timestamppb.New(at)), keystone.WithSensorMeasurementID(uuid.NewString())),
	}}
	if !appScope {
		if d.fixture.entityID == "" {
			return fmt.Errorf("entity fixture is unavailable")
		}
		request.Schema = d.schema()
		request.EntityId = d.fixture.entityID
	}
	// The queue validates before acknowledging: invalid measurements cannot be accepted.
	request.Measurements[0].TtlSeconds = -1
	_, err := actor.WriteSensors(ctx, request, keystone.WithSensorBackground(true))
	if status.Code(err) != codes.InvalidArgument {
		return fmt.Errorf("background write with negative TTL: want InvalidArgument; got %v", err)
	}
	request.Measurements[0].TtlSeconds = 0
	response, err := actor.WriteSensors(ctx, request, keystone.WithSensorBackground(true))
	if err != nil {
		return err
	}
	if !response.GetSuccess() {
		return fmt.Errorf("background write was not acknowledged")
	}
	// Success acknowledges durable queueing, so wait for the measurement to become readable.
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		chart, err := actor.ChartSensor(ctx, &proto.ChartSensorRequest{
			Schema: request.Schema, EntityId: request.EntityId, AppScope: appScope, Sensor: name,
			From: timestamppb.New(at), Until: timestamppb.New(at.Add(time.Minute)), IntervalSeconds: 60,
		})
		if err != nil {
			return err
		}
		var sum float64
		var samples int64
		for _, bucket := range chart.GetBuckets() {
			sum += bucket.GetSum()
			samples += bucket.GetSamples()
		}
		if samples != 0 {
			if samples != 1 || !closeEnough(sum, 5) {
				return fmt.Errorf("background ingestion: got sum=%g samples=%d; want sum=5 samples=1", sum, samples)
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("background measurement did not become readable: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
