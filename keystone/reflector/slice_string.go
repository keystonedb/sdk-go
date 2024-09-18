package reflector

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"reflect"
)

type StringSlice struct{}

func (e StringSlice) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if slice, ok := value.Interface().([]string); ok {
		return &proto.Value{Array: &proto.RepeatedValue{Strings: slice}}, nil
	}
	return nil, UnsupportedTypeError
}

func (e StringSlice) SetValue(value *proto.Value, onto reflect.Value) error {
	var slice []string
	if value.Array != nil {
		slice = value.Array.Strings
	}
	onto.Set(reflect.ValueOf(slice))
	return nil
}
