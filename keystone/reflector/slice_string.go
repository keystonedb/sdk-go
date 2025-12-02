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
	if value.Kind() == reflect.Slice && value.Type().Elem().Kind() == reflect.String {
		slice := make([]string, value.Len())
		for i := 0; i < value.Len(); i++ {
			slice[i] = value.Index(i).String()
		}
		return &proto.Value{Array: &proto.RepeatedValue{Strings: slice}, KnownType: proto.Property_Strings}, nil
	}
	return nil, UnsupportedTypeError
}

func (e StringSlice) SetValue(value *proto.Value, onto reflect.Value) error {
	elemType := onto.Type().Elem()
	elemSlice := reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0)
	if value.Array != nil {
		strings := value.Array.Strings
		elemSlice = reflect.MakeSlice(reflect.SliceOf(elemType), 0, len(strings))
		v := reflect.New(elemType).Elem()
		for _, s := range strings {
			v.SetString(s)
			elemSlice = reflect.Append(elemSlice, v)
		}
	}
	onto.Set(elemSlice)
	return nil
}

func (e StringSlice) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Strings}
}
