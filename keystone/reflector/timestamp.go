package reflector

import (
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
)

type Timestamp struct{}

func (e Timestamp) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if tme, isTime := value.Interface().(timestamppb.Timestamp); isTime {
		return &proto.Value{Time: &tme}, nil
	}
	return nil, UnsupportedTypeError
}

func (e Timestamp) SetValue(value *proto.Value, onto reflect.Value) error {
	onto.Set(reflect.ValueOf(value.GetTime()))
	return nil
}
