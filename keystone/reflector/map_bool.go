package reflector

import (
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
)

type BoolMap struct{}

func (e BoolMap) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if mapVal, ok := value.Interface().(map[string]bool); ok {
		ret := &proto.Value{Array: proto.NewRepeatedValue(), KnownType: proto.Property_KeyValue}
		for k, v := range mapVal {
			if v {
				ret.Array.KeyValue[k] = []byte("1")
			} else {
				ret.Array.KeyValue[k] = []byte("0")
			}
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e BoolMap) SetValue(value *proto.Value, onto reflect.Value) error {
	if value.Array == nil {
		return InvalidValueError
	}
	res := make(map[string]bool)
	for k, v := range value.Array.KeyValue {
		res[k] = string(v) == "1"
	}

	onto.Set(reflect.ValueOf(res))
	return nil
}

func (e BoolMap) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_KeyValue}
}
