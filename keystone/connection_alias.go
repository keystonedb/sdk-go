package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
)

func (c *Connection) AttachAlias(ctx context.Context, in *proto.AliasMutationRequest, opts ...grpc.CallOption) (*proto.GenericResponse, error) {
	tl := c.timeLogConfig.NewLog("AttachAlias")
	response, err := c.client.AttachAlias(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}
func (c *Connection) RemoveAlias(ctx context.Context, in *proto.AliasMutationRequest, opts ...grpc.CallOption) (*proto.GenericResponse, error) {
	tl := c.timeLogConfig.NewLog("RemoveAlias")
	response, err := c.client.RemoveAlias(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}
func (c *Connection) ResolveAlias(ctx context.Context, in *proto.ResolveAliasRequest, opts ...grpc.CallOption) (*proto.ResolveAliasResponse, error) {
	tl := c.timeLogConfig.NewLog("ResolveAlias")
	response, err := c.client.ResolveAlias(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}
func (c *Connection) ListAliases(ctx context.Context, in *proto.ListAliasesRequest, opts ...grpc.CallOption) (*proto.ListAliasesResponse, error) {
	tl := c.timeLogConfig.NewLog("ListAliases")
	response, err := c.client.ListAliases(ctx, in, opts...)
	c.logger.TimedLog(tl)
	return response, err
}
