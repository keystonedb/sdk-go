package reflector

import (
	proto2 "github.com/keystonedb/sdk-go/proto"
	"reflect"
)

type StringMap struct{}

func (e StringMap) ToProto(value reflect.Value) (*proto2.Value, error) {
	value = Deref(value)
	if mapVal, ok := value.Interface().(map[string]string); ok {
		ret := &proto2.Value{Array: proto2.NewRepeatedKeyValue()}
		for k, v := range mapVal {
			ret.Array.KeyValue[k] = []byte(v)
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e StringMap) SetValue(value *proto2.Value, onto reflect.Value) error {
	if value.Array == nil {
		return InvalidValueError
	}
	ret := make(map[string]string)
	for k, v := range value.Array.KeyValue {
		ret[k] = string(v)
	}
	onto.Set(reflect.ValueOf(ret))
	return nil
}
