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
		return &proto.Value{Time: timestamppb.New(time.Time{}), KnownType: proto.Property_Time}, nil
	}
	if tme, isTime := value.Interface().(time.Time); isTime {
		return &proto.Value{Time: timestamppb.New(tme), KnownType: proto.Property_Time}, nil
	}
	return nil, UnsupportedTypeError
}

func (e Time) SetValue(value *proto.Value, onto reflect.Value) error {
	timeValue := value.GetTime().AsTime()
	if onto.Kind() == reflect.Pointer {
		if onto.IsNil() {
			onto.Set(reflect.New(onto.Type().Elem()))
		}
		onto = onto.Elem()
	}
	onto.Set(reflect.ValueOf(timeValue))
	return nil
}

func (e Time) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Time}
}
