package reflector

import (
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
)

type Bool struct{}

func (e Bool) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	return &proto.Value{Bool: value.Bool(), KnownType: proto.Property_Boolean}, nil
}

func (e Bool) SetValue(value *proto.Value, onto reflect.Value) error {
	onto.SetBool(value.GetBool())
	return nil
}

func (e Bool) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Boolean}
}
