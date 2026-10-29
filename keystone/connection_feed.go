package keystone

import (
	"context"

	"github.com/keystonedb/sdk-go/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (c *Connection) FeedPut(ctx context.Context, in *proto.FeedPutRequest, opts ...grpc.CallOption) (*proto.FeedPutResponse, error) {
	tl := c.timeLogConfig.NewLog("FeedPut", zap.String("feed_type", in.GetFeed().GetType()))
	response, err := c.client.FeedPut(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}

func (c *Connection) FeedRead(ctx context.Context, in *proto.FeedReadRequest, opts ...grpc.CallOption) (*proto.FeedReadResponse, error) {
	tl := c.timeLogConfig.NewLog("FeedRead", zap.String("feed_type", in.GetFeed().GetType()))
	response, err := c.client.FeedRead(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}

func (c *Connection) FeedRemove(ctx context.Context, in *proto.FeedRemoveRequest, opts ...grpc.CallOption) (*proto.GenericResponse, error) {
	tl := c.timeLogConfig.NewLog("FeedRemove", zap.String("feed_type", in.GetFeed().GetType()))
	response, err := c.client.FeedRemove(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}

func (c *Connection) FeedSummary(ctx context.Context, in *proto.FeedSummaryRequest, opts ...grpc.CallOption) (*proto.FeedSummaryResponse, error) {
	tl := c.timeLogConfig.NewLog("FeedSummary", zap.String("feed_type", in.GetFeed().GetType()))
	response, err := c.client.FeedSummary(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}
