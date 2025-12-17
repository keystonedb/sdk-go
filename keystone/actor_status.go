package keystone

import (
	"context"

	"github.com/keystonedb/sdk-go/proto"
)

func (a *Actor) ServerStatus() (*proto.StatusResponse, error) {
	return a.Connection().Status(context.Background(), a.Authorization())
}
