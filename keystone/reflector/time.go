package reflector

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"time"
)

type Time struct{}

func (e Time) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if tme, isTime := value.Interface().(time.Time); isTime {
		return &proto.Value{Time: timestamppb.New(tme)}, nil
	}
	return nil, UnsupportedTypeError
}

func (e Time) SetValue(value *proto.Value, onto reflect.Value) error {
	onto.Set(reflect.ValueOf(value.GetTime().AsTime()))
	return nil
}