package reflector

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"reflect"
)

type Bytes struct{}

func (e Bytes) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	return &proto.Value{Raw: value.Bytes()}, nil
}

func (e Bytes) SetValue(value *proto.Value, onto reflect.Value) error {
	onto.SetBytes(value.GetRaw())
	return nil
}