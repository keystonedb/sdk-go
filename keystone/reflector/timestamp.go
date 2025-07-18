package reflector

import (
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Timestamp struct{}

func (e Timestamp) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if !value.IsValid() {
		value = reflect.New(reflect.TypeOf(&timestamppb.Timestamp{}))
	}
	if tme, isTime := value.Interface().(timestamppb.Timestamp); isTime {
		return &proto.Value{Time: &tme, KnownType: proto.Property_Time}, nil
	}
	return nil, UnsupportedTypeError
}

func (e Timestamp) SetValue(value *proto.Value, onto reflect.Value) error {
	onto.Set(reflect.ValueOf(value.GetTime()))
	return nil
}

func (e Timestamp) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Time}
}
