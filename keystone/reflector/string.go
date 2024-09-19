package reflector

import (
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
)

type String struct{}

func (e String) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	return &proto.Value{Text: value.String()}, nil
}

func (e String) SetValue(value *proto.Value, onto reflect.Value) error {
	onto.SetString(value.GetText())
	return nil
}
