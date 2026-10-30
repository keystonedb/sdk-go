package keystone

import (
	"context"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// Simulate a caller that still implements the old sensor-provider methods.
type legacySensorEntity struct {
	BaseEntity
	Name    string
	cleared bool
}

func (e *legacySensorEntity) GetSensorMeasurements() []*proto.EntitySensorMeasurement {
	return []*proto.EntitySensorMeasurement{NewSensorMeasurement("legacy", 1)}
}

func (e *legacySensorEntity) ClearSensorMeasurements() error {
	e.cleared = true
	return nil
}

type sensorMutationClient struct {
	proto.KeystoneClient
	requests []*proto.MutateRequest
}

func (c *sensorMutationClient) Mutate(_ context.Context, request *proto.MutateRequest, _ ...grpc.CallOption) (*proto.MutateResponse, error) {
	c.requests = append(c.requests, request)
	return &proto.MutateResponse{Success: true, EntityId: request.GetEntityId()}, nil
}

func TestMutationsDoNotCollectSensorMeasurements(t *testing.T) {
	for _, remote := range []bool{false, true} {
		name := "Mutate"
		if remote {
			name = "RemoteMutate"
		}
		t.Run(name, func(t *testing.T) {
			client := &sensorMutationClient{}
			conn := NewConnection(client, "vendor", "app", "token")
			entity := &legacySensorEntity{Name: "Entity properties only"}
			entity.SetKeystoneID(ID("entity"))
			conn.RegisterTypes(entity)
			actor := conn.Actor("workspace", "", "user", "test")
			var err error
			if remote {
				err = actor.RemoteMutate(context.Background(), entity.GetKeystoneID(), entity)
			} else {
				err = actor.Mutate(context.Background(), entity)
			}
			require.NoError(t, err)
			require.Len(t, client.requests, 1)
			require.Empty(t, client.requests[0].GetMutation().GetMeasurements())
			require.False(t, entity.cleared, "entity mutation must not consume sensor state")
			if !remote {
				require.NotEmpty(t, client.requests[0].GetMutation().GetProperties())
			}
		})
	}
}
