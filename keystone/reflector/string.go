package reflector

import (
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
)

type String struct{}

func (e String) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	return &proto.Value{Text: value.String(), KnownType: proto.Property_Text}, nil
}

func (e String) SetValue(value *proto.Value, onto reflect.Value) error {
	onto.SetString(value.GetText())
	return nil
}

func (e String) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Text}
}
