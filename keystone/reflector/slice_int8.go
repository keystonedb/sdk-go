package reflector

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"reflect"
)

type Int8Slice struct{}

func (e Int8Slice) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if slice, ok := value.Interface().([]int8); ok {
		ret := &proto.Value{Array: &proto.RepeatedValue{}}
		for _, i := range slice {
			ret.Array.Ints = append(ret.Array.Ints, int64(i))
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e Int8Slice) SetValue(value *proto.Value, onto reflect.Value) error {
	var slice []int8
	if value.Array != nil {
		for _, i := range value.Array.Ints {
			slice = append(slice, int8(i))
		}
	}
	onto.Set(reflect.ValueOf(slice))
	return nil
}
