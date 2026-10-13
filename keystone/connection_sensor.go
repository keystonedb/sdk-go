package keystone

import (
	"context"

	"github.com/keystonedb/sdk-go/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ChartSensor reads native entity sensor measurements.
func (c *Connection) ChartSensor(ctx context.Context, request *proto.ChartSensorRequest, options ...grpc.CallOption) (*proto.ChartSensorResponse, error) {
	tl := c.timeLogConfig.NewLog("ChartSensor", zap.String("schema", request.GetSchema().GetKey()), zap.String("EntityId", request.GetEntityId()), zap.String("sensor", request.GetSensor()))
	resp, err := c.client.ChartSensor(ctx, request, options...)
	c.logger.TimedLog(tl)
	return resp, err
}

// HeatmapSensor retrieves a heatmap of native entity sensor measurements.
func (c *Connection) HeatmapSensor(ctx context.Context, request *proto.HeatmapSensorRequest, options ...grpc.CallOption) (*proto.HeatmapSensorResponse, error) {
	tl := c.timeLogConfig.NewLog("HeatmapSensor", zap.String("schema", request.GetSchema().GetKey()), zap.String("EntityId", request.GetEntityId()), zap.String("sensor", request.GetSensor()))
	resp, err := c.client.HeatmapSensor(ctx, request, options...)
	c.logger.TimedLog(tl)
	return resp, err
}

// SensorActivity retrieves calendar activity for a native entity sensor.
func (c *Connection) SensorActivity(ctx context.Context, request *proto.SensorActivityRequest, options ...grpc.CallOption) (*proto.SensorActivityResponse, error) {
	tl := c.timeLogConfig.NewLog("SensorActivity", zap.String("schema", request.GetSchema().GetKey()), zap.String("EntityId", request.GetEntityId()), zap.String("sensor", request.GetSensor()))
	resp, err := c.client.SensorActivity(ctx, request, options...)
	c.logger.TimedLog(tl)
	return resp, err
}

// PredictSensor predicts native entity sensor measurements.
func (c *Connection) PredictSensor(ctx context.Context, request *proto.PredictSensorRequest, options ...grpc.CallOption) (*proto.PredictSensorResponse, error) {
	tl := c.timeLogConfig.NewLog("PredictSensor", zap.String("schema", request.GetSchema().GetKey()), zap.String("EntityId", request.GetEntityId()), zap.String("sensor", request.GetSensor()))
	resp, err := c.client.PredictSensor(ctx, request, options...)
	c.logger.TimedLog(tl)
	return resp, err
}

// WriteSensors writes measurements independently of entity mutations.
func (c *Connection) WriteSensors(ctx context.Context, request *proto.WriteSensorsRequest, options ...grpc.CallOption) (*proto.GenericResponse, error) {
	tl := c.timeLogConfig.NewLog("WriteSensors", zap.String("schema", request.GetSchema().GetKey()), zap.String("EntityId", request.GetEntityId()))
	response, err := c.client.WriteSensors(ctx, request, options...)
	c.logger.TimedLog(tl)
	return response, err
}

// CreateSensorStream creates a short-lived browser sensor stream grant.
func (c *Connection) CreateSensorStream(ctx context.Context, request *proto.CreateSensorStreamRequest, options ...grpc.CallOption) (*proto.CreateSensorStreamResponse, error) {
	tl := c.timeLogConfig.NewLog("CreateSensorStream", zap.String("schema", request.GetSchema().GetKey()), zap.String("EntityId", request.GetEntityId()))
	response, err := c.client.CreateSensorStream(ctx, request, options...)
	c.logger.TimedLog(tl)
	return response, err
}

// RevokeSensorSession revokes sensor grants and connections for a login session.
func (c *Connection) RevokeSensorSession(ctx context.Context, request *proto.RevokeSensorSessionRequest, options ...grpc.CallOption) (*proto.GenericResponse, error) {
	tl := c.timeLogConfig.NewLog("RevokeSensorSession", zap.String("session", request.GetSessionId()))
	response, err := c.client.RevokeSensorSession(ctx, request, options...)
	c.logger.TimedLog(tl)
	return response, err
}

// ListAppSensors lists sensor names in the authenticated app and workspace.
func (c *Connection) ListAppSensors(ctx context.Context, authorization *proto.Authorization, options ...grpc.CallOption) (*proto.ListAppSensorsResponse, error) {
	tl := c.timeLogConfig.NewLog("ListAppSensors")
	response, err := c.client.ListAppSensors(ctx, authorization, options...)
	c.logger.TimedLog(tl)
	return response, err
}
