package keystone

import (
	"context"
	"github.com/keystonedb/sdk-go/proto"
)

func (a *Actor) Squid(sequenceKey string) (*proto.SquidResponse, error) {
	return a.Connection().SQUID(context.Background(), &proto.SquidRequest{
		Authorization: a.Authorization(),
		SequenceKey:   sequenceKey,
	})
}

func (a *Actor) SquidRetrieve(sequenceKey, squat string) (*proto.SquidResponse, error) {
	return a.Connection().SQUIDRecover(context.Background(), &proto.SquidRecoverRequest{
		Authorization: a.Authorization(),
		SequenceKey:   sequenceKey,
		Squat:         squat,
	})
}
