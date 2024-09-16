package keystone

import (
	"github.com/keystonedb/sdk-go/sdk-go/proto"
)

type Unmarshaler interface {
	UnmarshalKeystone(map[Property]*proto.Value) error
}

type Marshaler interface {
	MarshalKeystone() (map[Property]*proto.Value, error)
}

type ValueMarshaler interface {
	MarshalValue() (*proto.Value, error)
}

type ValueUnmarshaler interface {
	UnmarshalValue(*proto.Value) error
}
