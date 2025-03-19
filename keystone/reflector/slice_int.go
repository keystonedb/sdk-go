package reflector

import (
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

type IntSlice struct{}

func (e IntSlice) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if value.Type().Kind() == reflect.Slice {
		ret := &proto.Value{Array: &proto.RepeatedValue{}}
		if value.Len() > 0 {
			for i := 0; i < value.Len(); i++ {
				ret.Array.Ints = append(ret.Array.Ints, value.Index(i).Int())
			}
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e IntSlice) SetValue(value *proto.Value, onto reflect.Value) error {
	elemSlice := reflect.MakeSlice(reflect.SliceOf(onto.Type().Elem()), 0, len(value.Array.Ints))
	if value.Array != nil {
		v := reflect.New(onto.Type().Elem()).Elem()
		for _, i := range value.Array.Ints {
			v.SetInt(i)
			elemSlice = reflect.Append(elemSlice, v)
		}
	}
	onto.Set(elemSlice)
	return nil
}

func (e IntSlice) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Ints}
}
