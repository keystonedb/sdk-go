package reflector

import (
	proto2 "github.com/keystonedb/sdk-go/proto"
	"reflect"
)

type Map struct{}

func (e Map) ToProto(value reflect.Value) (*proto2.Value, error) {
	value = Deref(value)
	if mapVal, ok := value.Interface().(map[string][]byte); ok {
		ret := &proto2.Value{Array: proto2.NewRepeatedKeyValue()}
		for k, v := range mapVal {
			ret.Array.KeyValue[k] = v
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e Map) SetValue(value *proto2.Value, onto reflect.Value) error {
	if value.Array == nil {
		return InvalidValueError
	}
	ret := make(map[string][]byte)
	for k, v := range value.Array.KeyValue {
		ret[k] = v
	}
	onto.Set(reflect.ValueOf(ret))
	return nil
}
