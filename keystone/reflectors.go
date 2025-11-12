package keystone

import (
	"errors"
	"reflect"
	"time"

	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	marshalerType        = reflect.TypeFor[Marshaler]()
	valueMarshalerType   = reflect.TypeFor[ValueMarshaler]()
	timeType             = reflect.TypeFor[time.Time]()
	refTimeType          = reflect.TypeFor[*time.Time]()
	timestampType        = reflect.TypeFor[timestamppb.Timestamp]()
	refTimestampType     = reflect.TypeFor[*timestamppb.Timestamp]()
	mutationObserverType = reflect.TypeFor[MutationObserver]()
	NestedChildType      = reflect.TypeFor[NestedChild]()
)

var kindReflector = map[reflect.Kind]Reflector{
	reflect.String:  reflector.String{},
	reflect.Bool:    reflector.Bool{},
	reflect.Int:     reflector.Int{Kind: reflect.Int},
	reflect.Int8:    reflector.Int{Kind: reflect.Int8},
	reflect.Int16:   reflector.Int{Kind: reflect.Int16},
	reflect.Int32:   reflector.Int{Kind: reflect.Int32},
	reflect.Int64:   reflector.Int{Kind: reflect.Int64},
	reflect.Uint:    reflector.Int{Kind: reflect.Uint},
	reflect.Uint8:   reflector.Int{Kind: reflect.Uint8},
	reflect.Uint16:  reflector.Int{Kind: reflect.Uint16},
	reflect.Uint32:  reflector.Int{Kind: reflect.Uint32},
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
	reflect.Int8:   reflector.IntSlice{},
	reflect.Int16:  reflector.IntSlice{},
	reflect.Int32:  reflector.IntSlice{},
	reflect.Uint16: reflector.IntSlice{},
	reflect.Uint32: reflector.IntSlice{},
	reflect.Int64:  reflector.IntSlice{},
}

var typeReflector = map[reflect.Type]Reflector{
	timeType:         reflector.Time{},
	refTimeType:      reflector.Time{},
	timestampType:    reflector.Timestamp{},
	refTimestampType: reflector.Timestamp{},
}

func GetReflector(t reflect.Type, v reflect.Value) Reflector {
	// Unwrap pointer types so primitive pointers use the same reflectors
	for t.Kind() == reflect.Pointer {
		if v.IsValid() {
			v = reflector.Deref(v)
		}
		t = t.Elem()
	}

	if ref, ok := typeReflector[t]; ok {
		return ref
	}

	if ref, ok := kindReflector[t.Kind()]; ok {
		if intRef, isIntRef := ref.(reflector.Int); isIntRef {
			intRef.Type = t
			return intRef
		}
		return ref
	}

	if t.Kind() == reflect.Map && t.Key().Kind() == reflect.String {
		if ref, ok := mapKindReflector[t.Elem().Kind()]; ok {
			return ref
		}
		if t.Elem().Kind() == reflect.Slice && t.Elem().Elem().Kind() == reflect.Uint8 {
			return reflector.Map{}
		}
	}

	if t.Kind() == reflect.Slice {
		if t.Elem().Implements(NestedChildType) {
			// Do not marshal nested children
			return nil
		}

		if ref, ok := sliceKindReflector[t.Elem().Kind()]; ok {
			return ref
		} else {
			//TODO: Add support for other slice types, maybe json blobs?
		}
	}
	if t.Implements(valueMarshalerType) {
		// Ensure we don't pass an invalid reflect.Value to newValueMarshalReflector
		if !v.IsValid() {
			vz := reflect.New(t).Elem()
			return newValueMarshalReflector(vz)
		}
		return newValueMarshalReflector(v)
	}

	if t.Kind() != reflect.Pointer && reflect.PointerTo(t).Implements(valueMarshalerType) {
		vp := reflect.New(t)
		// Only set if we have a valid value; otherwise keep zero value
		if v.IsValid() {
			vp.Elem().Set(v)
		}
		return newValueMarshalReflector(vp)
	} else if t.Implements(valueMarshalerType) {
		// Duplicate guard for completeness when reaching here
		if !v.IsValid() {
			vz := reflect.New(t).Elem()
			return newValueMarshalReflector(vz)
		}
		return newValueMarshalReflector(v)
	}

	return nil
}

type valueMarshalReflector struct {
	marshal ValueMarshaler
}

func (v valueMarshalReflector) IsZero() bool {
	return v.marshal.IsZero()
}

func (v valueMarshalReflector) ToProto(value reflect.Value) (*proto.Value, error) {
	return v.marshal.MarshalValue()
}

func (v valueMarshalReflector) SetValue(value *proto.Value, onto reflect.Value) error {
	newVal := reflect.ValueOf(v.marshal)
	if newVal.IsZero() || newVal.IsNil() {
		newVal = reflect.New(reflect.TypeOf(v.marshal).Elem())
	}
	msh, ok := newVal.Interface().(ValueMarshaler)
	if !ok {
		return errors.New("could not convert to ValueMarshaler")
	}
	err := msh.UnmarshalValue(value)
	if err != nil {
		return err
	}
	if newVal.Kind() == reflect.Pointer && onto.Kind() != reflect.Pointer {
		onto.Set(newVal.Elem())
	} else {
		onto.Set(newVal)
	}
	return nil
}

func (v valueMarshalReflector) PropertyDefinition() proto.PropertyDefinition {
	return v.marshal.PropertyDefinition()
}

func newValueMarshalReflector(value reflect.Value) Reflector {
	m, ok := value.Interface().(ValueMarshaler)
	if !ok {
		return nil
	}
	return valueMarshalReflector{marshal: m}
}
