package reflector

import (
	"reflect"
	"time"

	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Time struct{}

func (e Time) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if !value.IsValid() {
		value = reflect.New(reflect.TypeOf(&time.Time{}))
	}
	if tme, isTime := value.Interface().(time.Time); isTime {
		return &proto.Value{Time: timestamppb.New(tme), KnownType: proto.Property_Time}, nil
	}
	return nil, UnsupportedTypeError
}

func (e Time) SetValue(value *proto.Value, onto reflect.Value) error {
	onto.Set(reflect.ValueOf(value.GetTime().AsTime()))
	return nil
}

func (e Time) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Time}
}
