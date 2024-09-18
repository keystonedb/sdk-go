package reflector

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"reflect"
)

type IntSlice struct{}

func (e IntSlice) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if slice, ok := value.Interface().([]int); ok {
		ret := &proto.Value{Array: &proto.RepeatedValue{}}
		for _, i := range slice {
			ret.Array.Ints = append(ret.Array.Ints, int64(i))
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e IntSlice) SetValue(value *proto.Value, onto reflect.Value) error {
	var slice []int
	if value.Array != nil {
		for _, i := range value.Array.Ints {
			slice = append(slice, int(i))
		}
	}
	onto.Set(reflect.ValueOf(slice))
	return nil
}
