package reflector

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"reflect"
)

type StringMap struct{}

func (e StringMap) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if mapVal, ok := value.Interface().(map[string]string); ok {
		ret := &proto.Value{Array: proto.NewRepeatedKeyValue()}
		for k, v := range mapVal {
			ret.Array.KeyValue[k] = []byte(v)
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e StringMap) SetValue(value *proto.Value, onto reflect.Value) error {
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
