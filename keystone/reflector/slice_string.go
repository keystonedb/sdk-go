package reflector

import (
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
)

type StringSlice struct{}

func (e StringSlice) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if slice, ok := value.Interface().([]string); ok {
		return &proto.Value{Array: &proto.RepeatedValue{Strings: slice}, KnownType: proto.Property_Strings}, nil
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

func (e StringSlice) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Strings}
}
