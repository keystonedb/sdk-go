package keystone

import "github.com/keystonedb/sdk-go/proto"

type Child struct {
	_childID   string
	_aggregate int64
	_type      string
	_data      map[string][]byte
}

func NewChild(from interface{}) *Child {
	c := &Child{
		_type: Type(from),
		_data: make(map[string][]byte),
	}

	if nc, ok := from.(NestedChild); ok {
		c.SetChildID(nc.ChildID())
		c._data = nc.KeystoneData()
	} else {
		// Only nested children can be created
		return nil
	}

	if nca, ok := from.(NestedChildAggregateValue); ok {
		c.SetAggregateValue(nca.AggregateValue())
	}

	return c
}

func (e *Child) keyType() string                 { return e._type }
func (e *Child) ChildID() string                 { return e._childID }
func (e *Child) SetChildID(id string)            { e._childID = id }
func (e *Child) AggregateValue() int64           { return e._aggregate }
func (e *Child) SetAggregateValue(val int64)     { e._aggregate = val }
func (e *Child) KeystoneData() map[string][]byte { return e._data }

func (e *Child) toProtoChild() *proto.EntityChild {
	return &proto.EntityChild{
		Type:  &proto.Key{Key: e._type},
		Cid:   e._childID,
		Value: e.AggregateValue(),
		Data:  e.KeystoneData(),
	}
}
