package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/encoders"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"time"
)

var (
	marshalerType      = reflect.TypeFor[Marshaler]()
	valueMarshalerType = reflect.TypeFor[ValueMarshaler]()
	timeType           = reflect.TypeFor[time.Time]()
	refTimeType        = reflect.TypeFor[*time.Time]()
	timestampType      = reflect.TypeFor[timestamppb.Timestamp]()
	refTimestampType   = reflect.TypeFor[*timestamppb.Timestamp]()
)

type encoderFunc func(reflect.Value) (*proto.Value, error)

func newTypeEncoder(t reflect.Type) encoderFunc {
	if t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(valueMarshalerType) {
		return addrValueUnmarshalFunc
	}

	if t.Implements(valueMarshalerType) {
		return valueUnmarshalFunc
	}

	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return encoders.String
	case reflect.Bool:
		return encoders.Bool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return encoders.Int
	case reflect.Float32:
		return encoders.Float32
	case reflect.Float64:
		return encoders.Float
	case reflect.Map:
		switch t.Elem().Kind() {
		case reflect.String:
			return encoders.StringMap
		case reflect.Int:
			return encoders.IntMap
		case reflect.Slice:
			switch t.Elem().Elem().Kind() {
			case reflect.Uint8:
				return encoders.Map
			}
			/*log.Println("Unsupported map slice type", t.Elem().Elem().Kind())
			default:
				log.Println("Unsupported map type", t.Elem().Kind())*/
		}
	case reflect.Slice:
		switch t.Elem().Kind() {
		case reflect.String:
			return encoders.StringSlice
		case reflect.Int:
			return encoders.IntSlice
		case reflect.Int8:
			return encoders.Int8Slice
		case reflect.Int16:
			return encoders.Int16Slice
		case reflect.Int32:
			return encoders.Int32Slice
		case reflect.Int64:
			return encoders.Int64Slice
		}
	case reflect.Struct:
		switch t {
		case timeType, refTimeType:
			return encoders.Time
		case timestampType, refTimestampType:
			return encoders.Timestamp
		}
	}
	return nil
}

func addrValueUnmarshalFunc(value reflect.Value) (*proto.Value, error) {
	vp := reflect.New(value.Type())
	vp.Elem().Set(value)
	return valueUnmarshalFunc(vp)
}

var InvalidValueMarshalerError = errors.New("not a ValueMarshaler")

func valueUnmarshalFunc(value reflect.Value) (*proto.Value, error) {
	if value.Kind() == reflect.Invalid || (value.Kind() == reflect.Pointer && value.IsNil()) {
		return nil, nil
	}
	m, ok := value.Interface().(ValueMarshaler)
	if !ok {
		return nil, InvalidValueMarshalerError
	}
	return m.MarshalValue()
}
