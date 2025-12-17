package keystone

import (
	"context"

	"github.com/keystonedb/sdk-go/proto"
)

type IncrementingID struct {
	eid         ID
	keys        []string
	includeKeys []string
	meta        map[string]string
}

func NewIncrementingID(eid ID, incrementKeys ...string) *IncrementingID {
	return &IncrementingID{
		eid:  eid,
		keys: incrementKeys,
	}
}

func (i *IncrementingID) WithRead(keys ...string) *IncrementingID {
	i.includeKeys = keys
	return i
}

func (i *IncrementingID) WithMeta(meta map[string]string) *IncrementingID {
	i.meta = meta
	return i
}

func (i *IncrementingID) Commit(a *Actor) (*proto.IIDResponse, error) {
	conn := a.Connection()
	req := &proto.IIDCreateRequest{
		Authorization: a.Authorization(),
		Eid:           i.eid.String(),
		Incr:          map[string]bool{},
		Meta:          i.meta,
	}
	for _, key := range i.keys {
		req.Incr[key] = true
	}
	for _, key := range i.includeKeys {
		req.Incr[key] = false
	}
	return conn.IID(context.Background(), req)
}
