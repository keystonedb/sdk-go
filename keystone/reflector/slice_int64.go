package reflector

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"reflect"
)

type Int64Slice struct{}

func (e Int64Slice) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if slice, ok := value.Interface().([]int64); ok {
		return &proto.Value{Array: &proto.RepeatedValue{Ints: slice}}, nil
	}
	return nil, UnsupportedTypeError
}

func (e Int64Slice) SetValue(value *proto.Value, onto reflect.Value) error {
	var slice []int64
	if value.Array != nil {
		slice = value.Array.Ints
	}

	onto.Set(reflect.ValueOf(slice))
	return nil
}
