package keystone

import (
	"context"
	"testing"
	"time"

	"github.com/keystonedb/sdk-go/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type sensorTestRequest interface {
	protobuf.Message
	GetAuthorization() *proto.Authorization
}

func testSensorRPC[Q sensorTestRequest, R protobuf.Message](t *testing.T,
	request Q, response R,
	setHandler func(*MockServer, func(context.Context, Q) (R, error)),
	connectionCall func(*Connection, context.Context, Q, ...grpc.CallOption) (R, error),
	actorCall func(*Actor, context.Context, Q, ...grpc.CallOption) (R, error),
	authorize func(Q, *proto.Authorization),
) {
	t.Helper()
	conn, mock, listener, server := MockConnection()
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	conn.appID = proto.VendorApp{VendorId: "vendor", AppId: "app"}
	conn.token = "token"
	actor := conn.Actor("workspace", "127.0.0.1", "user", "sdk-test")
	actor.SetTraceID("trace")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	original := protobuf.Clone(request).(Q)
	expected := request
	var serverErr error
	setHandler(mock, func(_ context.Context, got Q) (R, error) {
		require.True(t, protobuf.Equal(expected, got), "request mismatch: %v", got)
		return response, serverErr
	})

	// Exercise the generated client and protobuf serialization, including call options.
	got, err := connectionCall(conn, ctx, request, grpc.MaxCallRecvMsgSize(1024*1024))
	require.NoError(t, err)
	require.True(t, protobuf.Equal(response, got))

	expected = protobuf.Clone(request).(Q)
	authorize(expected, actor.Authorization())
	if withSchema, ok := any(expected).(interface{ GetSchema() *proto.Key }); ok && withSchema.GetSchema() != nil {
		withSchema.GetSchema().Source = actor.VendorApp()
	}
	got, err = actorCall(&actor, ctx, request, grpc.MaxCallRecvMsgSize(1024*1024))
	require.NoError(t, err)
	require.True(t, protobuf.Equal(response, got))
	require.True(t, protobuf.Equal(original, request), "actor mutated caller's request")

	// An explicitly supplied schema source must survive the actor wrapper.
	if withSchema, ok := any(request).(interface{ GetSchema() *proto.Key }); ok && withSchema.GetSchema() != nil {
		withSchema.GetSchema().Source = &proto.VendorApp{VendorId: "other", AppId: "other"}
		expected = protobuf.Clone(request).(Q)
		authorize(expected, actor.Authorization())
		got, err = actorCall(&actor, ctx, request)
		require.NoError(t, err)
		require.True(t, protobuf.Equal(response, got))
	}

	// A restrictive receive limit verifies that call options reach the gRPC client.
	_, err = actorCall(&actor, ctx, request, grpc.MaxCallRecvMsgSize(1))
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
	expected = request
	_, err = connectionCall(conn, ctx, request, grpc.MaxCallRecvMsgSize(1))
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
	expected = protobuf.Clone(request).(Q)
	authorize(expected, actor.Authorization())

	canceledCtx, cancelCall := context.WithCancel(ctx)
	cancelCall()
	_, err = actorCall(&actor, canceledCtx, request)
	require.Equal(t, codes.Canceled, status.Code(err))

	serverErr = status.Error(codes.PermissionDenied, "sensor access denied")
	_, err = actorCall(&actor, ctx, request)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
	expected = request
	_, err = connectionCall(conn, ctx, request)
	require.Equal(t, codes.PermissionDenied, status.Code(err))

	var nilRequest Q
	_, err = actorCall(&actor, ctx, nilRequest)
	require.EqualError(t, err, "sensor request is nil")
	_, err = actorCall(nil, ctx, request)
	require.EqualError(t, err, "actor or connection is nil")
	_, err = actorCall(&Actor{}, ctx, request)
	require.EqualError(t, err, "actor or connection is nil")

	// Unconfigured mock handlers retain the normal Unimplemented behavior.
	setHandler(mock, nil)
	_, err = connectionCall(conn, ctx, request)
	require.Equal(t, codes.Unimplemented, status.Code(err))
}

func TestChartSensorRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.ChartSensorRequest{Authorization: &proto.Authorization{Token: "original"}, Schema: &proto.Key{Key: "device"}, EntityId: "entity", Sensor: "temperature", From: timestamppb.Now(), Until: timestamppb.Now(), IntervalSeconds: 300, ExcludeBackfill: true, TargetPoints: 200},
		&proto.ChartSensorResponse{ResolutionSeconds: 300, Estimated: true},
		func(m *MockServer, handler func(context.Context, *proto.ChartSensorRequest) (*proto.ChartSensorResponse, error)) {
			m.ChartSensorFunc = handler
		},
		(*Connection).ChartSensor, (*Actor).ChartSensor,
		func(req *proto.ChartSensorRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestHeatmapSensorRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.HeatmapSensorRequest{Authorization: &proto.Authorization{Token: "original"}, Schema: &proto.Key{Key: "device"}, EntityId: "entity", Sensor: "temperature", From: timestamppb.Now(), Until: timestamppb.Now(), Layout: proto.SensorHeatmapLayout(1), Timezone: "Europe/London", ExcludeBackfill: true},
		&proto.HeatmapSensorResponse{Timezone: "Europe/London", DisplayIntervalSeconds: 3600},
		func(m *MockServer, handler func(context.Context, *proto.HeatmapSensorRequest) (*proto.HeatmapSensorResponse, error)) {
			m.HeatmapSensorFunc = handler
		},
		(*Connection).HeatmapSensor, (*Actor).HeatmapSensor,
		func(req *proto.HeatmapSensorRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestSensorActivityRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.SensorActivityRequest{Authorization: &proto.Authorization{Token: "original"}, Schema: &proto.Key{Key: "device"}, EntityId: "entity", Sensor: "temperature", LookbackDays: 120, Calendar: proto.SensorActivityCalendar(1), MonthOfYear: 10},
		&proto.SensorActivityResponse{Timezone: "UTC", RankedSlots: []uint32{1, 2}},
		func(m *MockServer, handler func(context.Context, *proto.SensorActivityRequest) (*proto.SensorActivityResponse, error)) {
			m.SensorActivityFunc = handler
		},
		(*Connection).SensorActivity, (*Actor).SensorActivity,
		func(req *proto.SensorActivityRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestPredictSensorRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.PredictSensorRequest{Authorization: &proto.Authorization{Token: "original"}, Schema: &proto.Key{Key: "device"}, EntityId: "entity", Sensor: "temperature", From: timestamppb.Now(), Until: timestamppb.Now(), IntervalSeconds: 300, TrainingDays: 56, Timezone: "Europe/London", Metric: proto.SensorPredictionMetric(1), RangeCoverage: 0.9, MinimumObservations: 4, MissingAsZero: true, DailySeasonality: true, ExcludeBackfill: true},
		&proto.PredictSensorResponse{ResolutionSeconds: 300, RangeCoverage: 0.9, Model: "staged_recent_daily_weekly_v2", Points: []*proto.SensorPredictionPoint{{Baseline: "daily", TrainingSamples: 20, Available: true}}},
		func(m *MockServer, handler func(context.Context, *proto.PredictSensorRequest) (*proto.PredictSensorResponse, error)) {
			m.PredictSensorFunc = handler
		},
		(*Connection).PredictSensor, (*Actor).PredictSensor,
		func(req *proto.PredictSensorRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestWriteSensorsEntityRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.WriteSensorsRequest{Authorization: &proto.Authorization{Token: "original"}, Schema: &proto.Key{Key: "device"}, EntityId: "entity", TtlSeconds: 30, Measurements: []*proto.EntitySensorMeasurement{{Sensor: "hits", Value: 5, At: timestamppb.Now(), TtlSeconds: 60, MeasurementId: "stable", Backfill: true, ResolutionSeconds: 60, ObservedSeconds: 60, ActiveSeconds: 30, Attempts: 2, Successes: 1}}},
		&proto.GenericResponse{Success: true},
		func(m *MockServer, handler func(context.Context, *proto.WriteSensorsRequest) (*proto.GenericResponse, error)) {
			m.WriteSensorsFunc = handler
		},
		(*Connection).WriteSensors, (*Actor).WriteSensors,
		func(req *proto.WriteSensorsRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestWriteSensorsAppRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.WriteSensorsRequest{Authorization: &proto.Authorization{Token: "original"}, AppScope: true, TtlSeconds: 30, Measurements: []*proto.EntitySensorMeasurement{{Sensor: "hits", Value: 5, At: timestamppb.Now(), TtlSeconds: 60, MeasurementId: "stable", Backfill: true, ResolutionSeconds: 60, ObservedSeconds: 60, ActiveSeconds: 30, Attempts: 2, Successes: 1}}},
		&proto.GenericResponse{Success: true},
		func(m *MockServer, handler func(context.Context, *proto.WriteSensorsRequest) (*proto.GenericResponse, error)) {
			m.WriteSensorsFunc = handler
		},
		(*Connection).WriteSensors, (*Actor).WriteSensors,
		func(req *proto.WriteSensorsRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestCreateSensorStreamEntityRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.CreateSensorStreamRequest{Authorization: &proto.Authorization{Token: "original"}, Schema: &proto.Key{Key: "device"}, EntityId: "entity", Sensor: "hits", SessionId: "login", SessionExpiresAt: timestamppb.Now(), Origin: "https://example.com", WindowSeconds: 3600, IntervalSeconds: 60, Metric: proto.SensorPredictionMetric_SensorPredictionHits, ExcludeBackfill: true, IncludePrediction: true},
		&proto.CreateSensorStreamResponse{SnapshotUrl: "https://example.com/snapshot", StreamUrl: "https://example.com/events", ExpiresAt: timestamppb.Now(), ServerTime: timestamppb.Now()},
		func(m *MockServer, handler func(context.Context, *proto.CreateSensorStreamRequest) (*proto.CreateSensorStreamResponse, error)) {
			m.CreateSensorStreamFunc = handler
		},
		(*Connection).CreateSensorStream, (*Actor).CreateSensorStream,
		func(req *proto.CreateSensorStreamRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestCreateSensorStreamAppRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.CreateSensorStreamRequest{Authorization: &proto.Authorization{Token: "original"}, AppScope: true, Sensor: "hits", SessionId: "login", SessionExpiresAt: timestamppb.Now(), Origin: "https://example.com", WindowSeconds: 3600, IntervalSeconds: 60, Metric: proto.SensorPredictionMetric_SensorPredictionHits, ExcludeBackfill: true, IncludePrediction: true},
		&proto.CreateSensorStreamResponse{SnapshotUrl: "https://example.com/snapshot", StreamUrl: "https://example.com/events", ExpiresAt: timestamppb.Now(), ServerTime: timestamppb.Now()},
		func(m *MockServer, handler func(context.Context, *proto.CreateSensorStreamRequest) (*proto.CreateSensorStreamResponse, error)) {
			m.CreateSensorStreamFunc = handler
		},
		(*Connection).CreateSensorStream, (*Actor).CreateSensorStream,
		func(req *proto.CreateSensorStreamRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestRevokeSensorSessionRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.RevokeSensorSessionRequest{Authorization: &proto.Authorization{Token: "original"}, SessionId: "login"},
		&proto.GenericResponse{Success: true},
		func(m *MockServer, handler func(context.Context, *proto.RevokeSensorSessionRequest) (*proto.GenericResponse, error)) {
			m.RevokeSensorSessionFunc = handler
		},
		(*Connection).RevokeSensorSession, (*Actor).RevokeSensorSession,
		func(req *proto.RevokeSensorSessionRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestChartSensorAppScopeRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.ChartSensorRequest{Authorization: &proto.Authorization{Token: "original"}, AppScope: true, Sensor: "hits"},
		&proto.ChartSensorResponse{ResolutionSeconds: 60},
		func(m *MockServer, handler func(context.Context, *proto.ChartSensorRequest) (*proto.ChartSensorResponse, error)) {
			m.ChartSensorFunc = handler
		},
		(*Connection).ChartSensor, (*Actor).ChartSensor,
		func(req *proto.ChartSensorRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestHeatmapSensorAppScopeRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.HeatmapSensorRequest{Authorization: &proto.Authorization{Token: "original"}, AppScope: true, Sensor: "hits"},
		&proto.HeatmapSensorResponse{Timezone: "UTC"},
		func(m *MockServer, handler func(context.Context, *proto.HeatmapSensorRequest) (*proto.HeatmapSensorResponse, error)) {
			m.HeatmapSensorFunc = handler
		},
		(*Connection).HeatmapSensor, (*Actor).HeatmapSensor,
		func(req *proto.HeatmapSensorRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestSensorActivityAppScopeRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.SensorActivityRequest{Authorization: &proto.Authorization{Token: "original"}, AppScope: true, Sensor: "hits"},
		&proto.SensorActivityResponse{Timezone: "UTC"},
		func(m *MockServer, handler func(context.Context, *proto.SensorActivityRequest) (*proto.SensorActivityResponse, error)) {
			m.SensorActivityFunc = handler
		},
		(*Connection).SensorActivity, (*Actor).SensorActivity,
		func(req *proto.SensorActivityRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestPredictSensorAppScopeRPC(t *testing.T) {
	testSensorRPC(t,
		&proto.PredictSensorRequest{Authorization: &proto.Authorization{Token: "original"}, AppScope: true, Sensor: "hits"},
		&proto.PredictSensorResponse{ResolutionSeconds: 60},
		func(m *MockServer, handler func(context.Context, *proto.PredictSensorRequest) (*proto.PredictSensorResponse, error)) {
			m.PredictSensorFunc = handler
		},
		(*Connection).PredictSensor, (*Actor).PredictSensor,
		func(req *proto.PredictSensorRequest, auth *proto.Authorization) { req.Authorization = auth },
	)
}

func TestListAppSensorsRPC(t *testing.T) {
	conn, mock, listener, server := MockConnection()
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	conn.appID = proto.VendorApp{VendorId: "vendor", AppId: "app"}
	conn.token = "token"
	actor := conn.Actor("workspace", "127.0.0.1", "user", "sdk-test")
	actor.SetTraceID("trace")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	auth := actor.Authorization()
	var serverErr error
	mock.ListAppSensorsFunc = func(_ context.Context, got *proto.Authorization) (*proto.ListAppSensorsResponse, error) {
		require.True(t, protobuf.Equal(auth, got))
		return &proto.ListAppSensorsResponse{Sensors: []string{"hits", "volume"}}, serverErr
	}
	for _, call := range []func(context.Context, ...grpc.CallOption) (*proto.ListAppSensorsResponse, error){
		actor.ListAppSensors,
		func(ctx context.Context, options ...grpc.CallOption) (*proto.ListAppSensorsResponse, error) {
			return conn.ListAppSensors(ctx, auth, options...)
		},
	} {
		response, err := call(ctx)
		require.NoError(t, err)
		require.Equal(t, []string{"hits", "volume"}, response.GetSensors())
		_, err = call(ctx, grpc.MaxCallRecvMsgSize(1))
		require.Equal(t, codes.ResourceExhausted, status.Code(err))
		serverErr = status.Error(codes.PermissionDenied, "denied")
		_, err = call(ctx)
		require.Equal(t, codes.PermissionDenied, status.Code(err))
		serverErr = nil
		canceled, cancelCall := context.WithCancel(ctx)
		cancelCall()
		_, err = call(canceled)
		require.Equal(t, codes.Canceled, status.Code(err))
	}
	for _, actor := range []*Actor{nil, {}} {
		_, err := actor.ListAppSensors(ctx)
		require.EqualError(t, err, "actor or connection is nil")
	}
	mock.ListAppSensorsFunc = nil
	_, err := actor.ListAppSensors(ctx)
	require.Equal(t, codes.Unimplemented, status.Code(err))
}
