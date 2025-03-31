package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/grpc"
	"io"
)

type Key struct {
	VendorID string
	AppID    string
	Type     string
}

func NewKey(vendorID, appID, entityType string) *Key {
	return &Key{
		VendorID: vendorID,
		AppID:    appID,
		Type:     entityType,
	}
}

func OwnKey(key string) *Key {
	return NewKey("", "", key)
}

func (k *Key) toProto(a *Actor) *proto.Key {
	if k == nil {
		return nil
	}
	ret := &proto.Key{
		Source: &proto.VendorApp{
			VendorId: k.VendorID,
			AppId:    k.AppID,
		},
		Key: k.Type,
	}

	if k.VendorID == "" {
		ret.Source = a.Authorization().GetSource()
	} else if k.AppID == "" {
		ret.Source.AppId = a.Authorization().GetSource().GetAppId()
	}

	return ret
}

func (a *Actor) EventStream(ctx context.Context, handler func(response *proto.EventStreamResponse) error, name string, eventType *Key) error {
	req := &proto.EventStreamRequest{
		Authorization: a.Authorization(),
		StreamName:    name,
		AllWorkspaces: a.WorkspaceID() == noWorkspace,
		EventType:     eventType.toProto(a),
	}

	stream, err := a.Connection().EventStream(ctx, req)
	if err != nil {
		return err
	}

	defer func(stream grpc.ServerStreamingClient[proto.EventStreamResponse]) {
		_ = stream.CloseSend()
	}(stream)

	for {
		evt, recErr := stream.Recv()
		if recErr == io.EOF {
			return nil
		}
		if recErr != nil {
			return recErr
		}
		handleErr := handler(evt)
		if handleErr != nil {
			return handleErr
		}
	}
}
