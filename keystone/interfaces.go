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
