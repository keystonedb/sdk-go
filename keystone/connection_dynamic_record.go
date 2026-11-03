package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
)

func (c *Connection) GetDynamicRecord(ctx context.Context, in *proto.GetDynamicRecordRequest, opts ...grpc.CallOption) (*proto.DynamicRecord, error) {
	tl := c.timeLogConfig.NewLog("GetDynamicRecord")
	response, err := c.client.GetDynamicRecord(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}

func (c *Connection) MutateDynamicRecord(ctx context.Context, in *proto.MutateDynamicRecordRequest, opts ...grpc.CallOption) (*proto.DynamicRecordMutationResponse, error) {
	tl := c.timeLogConfig.NewLog("MutateDynamicRecord")
	response, err := c.client.MutateDynamicRecord(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}

func (c *Connection) DeleteDynamicRecord(ctx context.Context, in *proto.DeleteDynamicRecordRequest, opts ...grpc.CallOption) (*proto.DynamicRecordMutationResponse, error) {
	tl := c.timeLogConfig.NewLog("DeleteDynamicRecord")
	response, err := c.client.DeleteDynamicRecord(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}

func (c *Connection) ListDynamicRecordHistory(ctx context.Context, in *proto.ListDynamicRecordHistoryRequest, opts ...grpc.CallOption) (*proto.ListDynamicRecordHistoryResponse, error) {
	tl := c.timeLogConfig.NewLog("ListDynamicRecordHistory")
	response, err := c.client.ListDynamicRecordHistory(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}
