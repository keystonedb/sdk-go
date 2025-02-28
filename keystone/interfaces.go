package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

type Marshaler interface {
	MarshalKeystone() (map[Property]*proto.Value, error)
	UnmarshalKeystone(map[Property]*proto.Value) error
}

type ValueMarshaler interface {
	MarshalValue() (*proto.Value, error)
	UnmarshalValue(*proto.Value) error
	PropertyDefinition() proto.PropertyDefinition
}

type Reflector interface {
	ToProto(reflect.Value) (*proto.Value, error)
	SetValue(*proto.Value, reflect.Value) error
	PropertyDefinition() proto.PropertyDefinition
}

type Entity interface {
	GetKeystoneID() string
	SetKeystoneID(id string)
}

type ChildEntity interface {
	GetKeystoneParentID() string
	GetKeystoneChildID() string
	SetKeystoneParentID(id string)
	SetKeystoneChildID(id string)
}

// NestedChild is an interface that defines a child struct - these are not standalone entities
type NestedChild interface {
	ChildID() string
	SetChildID(id string)
	KeystoneData() map[string][]byte
	KeystoneDataAppend() map[string][]byte
	KeystoneRemoveData() []string
}

// NestedChildAggregateValue defines the aggregate Value of a child entity
type NestedChildAggregateValue interface {
	AggregateValue() int64
	SetAggregateValue(val int64)
}

type MutationObserver interface {
	MutationSuccess(response *proto.MutateResponse)
}
