package reflector

import (
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
)

type Int struct {
	Kind reflect.Kind
	Type reflect.Type
}

func (e Int) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if value.CanInt() {
		return &proto.Value{Int: value.Int(), KnownType: proto.Property_Number}, nil
	}

	if value.CanUint() && value.Kind() != reflect.Uint64 && value.Kind() != reflect.Uintptr {
		return &proto.Value{Int: int64(value.Uint()), KnownType: proto.Property_Number}, nil
	}
	return nil, UnsupportedTypeError
}

func (e Int) cast(value reflect.Value) reflect.Value {
	if e.Type == nil {
		return value
	}
	return value.Convert(e.Type)
}

func (e Int) SetValue(value *proto.Value, onto reflect.Value) error {
	switch e.Kind {
	case reflect.Int:
		onto.Set(e.cast(reflect.ValueOf(int(value.Int))))
		return nil
	case reflect.Int8:
		onto.Set(e.cast(reflect.ValueOf(int8(value.Int))))
		return nil
	case reflect.Int16:
		onto.Set(e.cast(reflect.ValueOf(int16(value.Int))))
		return nil
	case reflect.Int32:
		onto.Set(e.cast(reflect.ValueOf(int32(value.Int))))
		return nil
	case reflect.Int64:
		onto.SetInt(value.Int)
		return nil
	case reflect.Uint8:
		onto.Set(e.cast(reflect.ValueOf(uint8(value.Int))))
		return nil
	case reflect.Uint16:
		onto.Set(e.cast(reflect.ValueOf(uint16(value.Int))))
		return nil
	case reflect.Uint32:
		onto.Set(e.cast(reflect.ValueOf(uint32(value.Int))))
		return nil
	default:
		return UnsupportedTypeError
	}
}

func (e Int) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Number}
}
