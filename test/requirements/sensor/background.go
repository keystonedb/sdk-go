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

func (d *Requirement) scheduleBackground(actor *keystone.Actor, report requirements.Reporter, scheduler *requirements.Scheduler) {
	for _, appScope := range []bool{false, true} {
		scheduler.After(0, func() { d.backgroundWrite(actor, appScope, report, scheduler) })
	}
}

func (d *Requirement) backgroundWrite(actor *keystone.Actor, appScope bool, report requirements.Reporter, scheduler *requirements.Scheduler) {
	name := "integration.background." + uuid.NewString()
	at := time.Now().UTC().Truncate(time.Minute).Add(-time.Minute)
	deadline := time.Now().Add(queryTimeout)
	result := func(err error) {
		report(requirements.NewResult(fmt.Sprintf("WriteSensors background ingestion (app scope=%t)", appScope), err))
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	request := &proto.WriteSensorsRequest{AppScope: appScope, Measurements: []*proto.EntitySensorMeasurement{
		keystone.NewSensorMeasurement(name, 5, keystone.WithSensorTimestamp(timestamppb.New(at)), keystone.WithSensorMeasurementID(uuid.NewString())),
	}}
	if !appScope {
		if d.fixture.entityID == "" {
			result(fmt.Errorf("entity fixture is unavailable"))
			return
		}
		request.Schema = d.schema()
		request.EntityId = d.fixture.entityID
	}
	request.Measurements[0].TtlSeconds = -1
	_, err := actor.WriteSensors(ctx, request, keystone.WithSensorBackground(true))
	if status.Code(err) != codes.InvalidArgument {
		result(fmt.Errorf("background write with negative TTL: want InvalidArgument; got %v", err))
		return
	}
	request.Measurements[0].TtlSeconds = 0
	response, err := actor.WriteSensors(ctx, request, keystone.WithSensorBackground(true))
	if err != nil {
		result(err)
		return
	}
	if !response.GetSuccess() {
		result(fmt.Errorf("background write was not acknowledged"))
		return
	}
	// Queue acknowledgement starts a separate polling phase. Polling yields to
	// other requirements and each RPC retains the original completion deadline.
	var check func()
	check = func() {
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()
		chart, err := actor.ChartSensor(ctx, &proto.ChartSensorRequest{
			Schema: request.Schema, EntityId: request.EntityId, AppScope: appScope, Sensor: name,
			From: timestamppb.New(at), Until: timestamppb.New(at.Add(time.Minute)), IntervalSeconds: 60,
		})
		if err != nil {
			result(err)
			return
		}
		var sum float64
		var samples int64
		for _, bucket := range chart.GetBuckets() {
			sum += bucket.GetSum()
			samples += bucket.GetSamples()
		}
		if samples != 0 {
			if samples != 1 || !closeEnough(sum, 5) {
				result(fmt.Errorf("background ingestion: got sum=%g samples=%d; want sum=5 samples=1", sum, samples))
				return
			}
			result(nil)
			return
		}
		if !time.Now().Before(deadline) {
			result(fmt.Errorf("background measurement did not become readable before its deadline"))
			return
		}
		scheduler.After(250*time.Millisecond, check)
	}
	scheduler.After(0, check)
}
