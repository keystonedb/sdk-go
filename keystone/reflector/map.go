package reflector

import (
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

type Map struct{}

func (e Map) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if mapVal, ok := value.Interface().(map[string][]byte); ok {
		ret := &proto.Value{Array: proto.NewRepeatedKeyValue()}
		for k, v := range mapVal {
			ret.Array.KeyValue[k] = v
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e Map) SetValue(value *proto.Value, onto reflect.Value) error {
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

func (e Map) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_KeyValue}
}
