package keystone

import (
	"context"
	"errors"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
	protobuf "google.golang.org/protobuf/proto"
)

// ChartSensor reads native entity sensor measurements using the actor's authorization.
// The request is cloned; a missing schema source defaults to the actor's app.
func (a *Actor) ChartSensor(ctx context.Context, request *proto.ChartSensorRequest, options ...grpc.CallOption) (*proto.ChartSensorResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	if request == nil {
		return nil, errors.New("sensor request is nil")
	}
	req := protobuf.Clone(request).(*proto.ChartSensorRequest)
	req.Authorization = a.Authorization()
	if req.Schema != nil && req.Schema.Source == nil {
		req.Schema.Source = a.VendorApp()
	}
	return a.connection.ChartSensor(ctx, req, options...)
}

// HeatmapSensor retrieves a heatmap of native entity sensor measurements using the actor's authorization.
// The request is cloned; a missing schema source defaults to the actor's app.
func (a *Actor) HeatmapSensor(ctx context.Context, request *proto.HeatmapSensorRequest, options ...grpc.CallOption) (*proto.HeatmapSensorResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	if request == nil {
		return nil, errors.New("sensor request is nil")
	}
	req := protobuf.Clone(request).(*proto.HeatmapSensorRequest)
	req.Authorization = a.Authorization()
	if req.Schema != nil && req.Schema.Source == nil {
		req.Schema.Source = a.VendorApp()
	}
	return a.connection.HeatmapSensor(ctx, req, options...)
}

// SensorActivity retrieves calendar activity for a native entity sensor using the actor's authorization.
// The request is cloned; a missing schema source defaults to the actor's app.
func (a *Actor) SensorActivity(ctx context.Context, request *proto.SensorActivityRequest, options ...grpc.CallOption) (*proto.SensorActivityResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	if request == nil {
		return nil, errors.New("sensor request is nil")
	}
	req := protobuf.Clone(request).(*proto.SensorActivityRequest)
	req.Authorization = a.Authorization()
	if req.Schema != nil && req.Schema.Source == nil {
		req.Schema.Source = a.VendorApp()
	}
	return a.connection.SensorActivity(ctx, req, options...)
}

// PredictSensor predicts native entity sensor measurements using the actor's authorization.
// The request is cloned; a missing schema source defaults to the actor's app.
func (a *Actor) PredictSensor(ctx context.Context, request *proto.PredictSensorRequest, options ...grpc.CallOption) (*proto.PredictSensorResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	if request == nil {
		return nil, errors.New("sensor request is nil")
	}
	req := protobuf.Clone(request).(*proto.PredictSensorRequest)
	req.Authorization = a.Authorization()
	if req.Schema != nil && req.Schema.Source == nil {
		req.Schema.Source = a.VendorApp()
	}
	return a.connection.PredictSensor(ctx, req, options...)
}

// WriteSensors writes measurements independently of entity mutations using the actor's authorization.
// The caller's request is preserved.
func (a *Actor) WriteSensors(ctx context.Context, request *proto.WriteSensorsRequest, options ...grpc.CallOption) (*proto.GenericResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	if request == nil {
		return nil, errors.New("sensor request is nil")
	}
	req := protobuf.Clone(request).(*proto.WriteSensorsRequest)
	req.Authorization = a.Authorization()
	if req.Schema != nil && req.Schema.Source == nil {
		req.Schema.Source = a.VendorApp()
	}
	return a.connection.WriteSensors(ctx, req, options...)
}

// CreateSensorStream creates a short-lived browser sensor stream grant using the actor's authorization.
// The caller's request is preserved.
func (a *Actor) CreateSensorStream(ctx context.Context, request *proto.CreateSensorStreamRequest, options ...grpc.CallOption) (*proto.CreateSensorStreamResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	if request == nil {
		return nil, errors.New("sensor request is nil")
	}
	req := protobuf.Clone(request).(*proto.CreateSensorStreamRequest)
	req.Authorization = a.Authorization()
	if req.Schema != nil && req.Schema.Source == nil {
		req.Schema.Source = a.VendorApp()
	}
	return a.connection.CreateSensorStream(ctx, req, options...)
}

// RevokeSensorSession revokes sensor grants and connections for a login session using the actor's authorization.
// The caller's request is preserved.
func (a *Actor) RevokeSensorSession(ctx context.Context, request *proto.RevokeSensorSessionRequest, options ...grpc.CallOption) (*proto.GenericResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	if request == nil {
		return nil, errors.New("sensor request is nil")
	}
	req := protobuf.Clone(request).(*proto.RevokeSensorSessionRequest)
	req.Authorization = a.Authorization()

	return a.connection.RevokeSensorSession(ctx, req, options...)
}

// ListAppSensors lists sensor names in the actor's app and workspace.
func (a *Actor) ListAppSensors(ctx context.Context, options ...grpc.CallOption) (*proto.ListAppSensorsResponse, error) {
	if a == nil || a.connection == nil {
		return nil, errors.New("actor or connection is nil")
	}
	return a.connection.ListAppSensors(ctx, a.Authorization(), options...)
}
