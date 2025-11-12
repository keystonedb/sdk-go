package reflector

import (
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
)

type Bool struct{}

func (e Bool) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if !value.IsValid() {
		return &proto.Value{Bool: false, KnownType: proto.Property_Boolean}, nil
	}
	return &proto.Value{Bool: value.Bool(), KnownType: proto.Property_Boolean}, nil
}

func (e Bool) SetValue(value *proto.Value, onto reflect.Value) error {
	if onto.Kind() == reflect.Pointer {
		if onto.IsNil() {
			onto.Set(reflect.New(onto.Type().Elem()))
		}
		onto = onto.Elem()
	}
	onto.SetBool(value.GetBool())
	return nil
}

func (e Bool) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Boolean}
}
