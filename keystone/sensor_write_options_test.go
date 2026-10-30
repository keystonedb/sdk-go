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
)

func TestWriteSensorsBackgroundOption(t *testing.T) {
	conn, mock, listener, server := MockConnection()
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	actor := conn.Actor("workspace", "", "user", "test")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, call := range []struct {
		name  string
		write func(context.Context, *proto.WriteSensorsRequest, ...grpc.CallOption) (*proto.GenericResponse, error)
	}{
		{"Connection", conn.WriteSensors},
		{"Actor", actor.WriteSensors},
	} {
		t.Run(call.name, func(t *testing.T) {
			for _, test := range []struct {
				name                    string
				requestBackground, want bool
				options                 []grpc.CallOption
			}{
				{"default synchronous", false, false, nil},
				{"preserves request", true, true, nil},
				{"enables background", false, true, []grpc.CallOption{WithSensorBackground(true)}},
				{"disables background", true, false, []grpc.CallOption{WithSensorBackground(false)}},
				{"last option wins", false, false, []grpc.CallOption{WithSensorBackground(true), WithSensorBackground(false)}},
				{"last option enables", false, true, []grpc.CallOption{WithSensorBackground(false), WithSensorBackground(true)}},
			} {
				t.Run(test.name, func(t *testing.T) {
					request := &proto.WriteSensorsRequest{AppScope: true, TtlSeconds: 60,
						Background:   test.requestBackground,
						Measurements: []*proto.EntitySensorMeasurement{NewSensorMeasurement("hits", 1)}}
					original := protobuf.Clone(request)
					expected := protobuf.Clone(request).(*proto.WriteSensorsRequest)
					expected.Background = test.want
					if call.name == "Actor" {
						expected.Authorization = actor.Authorization()
					}
					mock.WriteSensorsFunc = func(_ context.Context, got *proto.WriteSensorsRequest) (*proto.GenericResponse, error) {
						require.True(t, protobuf.Equal(expected, got), "unexpected request: %v", got)
						return &proto.GenericResponse{Success: true}, nil
					}
					response, err := call.write(ctx, request, test.options...)
					require.NoError(t, err)
					require.True(t, response.GetSuccess())
					require.True(t, protobuf.Equal(original, request), "caller request was modified")
					// Ordinary gRPC options still reach the transport when mixed with SDK options.
					options := append([]grpc.CallOption{grpc.MaxCallRecvMsgSize(1)}, test.options...)
					_, err = call.write(ctx, request, options...)
					require.Equal(t, codes.ResourceExhausted, status.Code(err))
				})
			}
			mock.WriteSensorsFunc = func(context.Context, *proto.WriteSensorsRequest) (*proto.GenericResponse, error) {
				return nil, status.Error(codes.InvalidArgument, "invalid measurement")
			}
			_, err := call.write(ctx, &proto.WriteSensorsRequest{}, WithSensorBackground(true))
			require.Equal(t, codes.InvalidArgument, status.Code(err))
			_, err = call.write(ctx, nil, WithSensorBackground(true))
			require.EqualError(t, err, "sensor request is nil")
		})
	}
}
