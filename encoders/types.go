package encoders

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"time"
)

var InvalidTimeError = errors.New("not a time.Time")

func Time(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if tme, isTime := value.Interface().(time.Time); isTime {
		return &proto.Value{Time: timestamppb.New(tme)}, nil
	}
	return nil, InvalidTimeError
}

var InvalidTimestampError = errors.New("not a timestamppb.Timestamp")

func Timestamp(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if tme, isTime := value.Interface().(timestamppb.Timestamp); isTime {
		return &proto.Value{Time: &tme}, nil
	}
	return nil, InvalidTimestampError
}
