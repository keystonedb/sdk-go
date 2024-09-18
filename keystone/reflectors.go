package keystone

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"github.com/keystonedb/sdk-go/keystone/reflector"
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

var kindReflector = map[reflect.Kind]Reflector{
	reflect.String:  reflector.String{},
	reflect.Bool:    reflector.Bool{},
	reflect.Int:     reflector.Int{Kind: reflect.Int},
	reflect.Int8:    reflector.Int{Kind: reflect.Int8},
	reflect.Int16:   reflector.Int{Kind: reflect.Int16},
	reflect.Int32:   reflector.Int{Kind: reflect.Int32},
	reflect.Int64:   reflector.Int{Kind: reflect.Int64},
	reflect.Float64: reflector.Float{},
	reflect.Float32: reflector.Float{Is32: true},
}

var mapKindReflector = map[reflect.Kind]Reflector{
	reflect.String: reflector.StringMap{},
	reflect.Int:    reflector.IntMap{},
}

var sliceKindReflector = map[reflect.Kind]Reflector{
	reflect.Uint8:  reflector.Bytes{},
	reflect.String: reflector.StringSlice{},
	reflect.Int:    reflector.IntSlice{},
	reflect.Int8:   reflector.Int8Slice{},
	reflect.Int16:  reflector.Int16Slice{},
	reflect.Int32:  reflector.Int32Slice{},
	reflect.Int64:  reflector.Int64Slice{},
}

var typeReflector = map[reflect.Type]Reflector{
	timeType:         reflector.Time{},
	refTimeType:      reflector.Time{},
	timestampType:    reflector.Timestamp{},
	refTimestampType: reflector.Timestamp{},
}

func GetReflector(t reflect.Type, v reflect.Value) Reflector {

	if ref, ok := typeReflector[t]; ok {
		return ref
	}

	if ref, ok := kindReflector[t.Kind()]; ok {
		return ref
	}

	if t.Kind() == reflect.Map {
		if ref, ok := mapKindReflector[t.Elem().Kind()]; ok {
			return ref
		}
		if t.Elem().Kind() == reflect.Slice && t.Elem().Elem().Kind() == reflect.Uint8 {
			return reflector.Map{}
		}
	}

	if t.Kind() == reflect.Slice {
		if ref, ok := sliceKindReflector[t.Elem().Kind()]; ok {
			return ref
		}
	}

	if t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(valueMarshalerType) {
		vp := reflect.New(t)
		vp.Elem().Set(v)
		return newValueMarshalReflector(vp)
	} else if t.Implements(valueMarshalerType) {
		return newValueMarshalReflector(v)
	}

	return nil
}

type valueMarshalReflector struct {
	marshal ValueMarshaler
}

func (v valueMarshalReflector) ToProto(value reflect.Value) (*proto.Value, error) {
	return v.marshal.MarshalValue()
}

func (v valueMarshalReflector) SetValue(value *proto.Value, onto reflect.Value) error {
	err := v.marshal.UnmarshalValue(value)
	if err != nil {
		return err
	}
	newVal := reflect.ValueOf(v.marshal)
	if newVal.Kind() == reflect.Pointer {
		onto.Set(newVal.Elem())
	} else {
		onto.Set(newVal)
	}
	return nil
}

func newValueMarshalReflector(value reflect.Value) Reflector {
	if value.Kind() == reflect.Invalid || (value.Kind() == reflect.Pointer && value.IsNil()) {
		return nil
	}
	m, ok := value.Interface().(ValueMarshaler)
	if !ok {
		return nil
	}
	return valueMarshalReflector{marshal: m}
}
