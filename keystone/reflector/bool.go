package reflector

import (
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

type Bool struct{}

func (e Bool) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	return &proto.Value{Bool: value.Bool()}, nil
}

func (e Bool) SetValue(value *proto.Value, onto reflect.Value) error {
	onto.SetBool(value.GetBool())
	return nil
}
