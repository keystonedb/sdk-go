package reflector

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/keystonedb/sdk-go/proto"
)

type Float struct {
	Is32 bool
}

func (e Float) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if !value.IsValid() {
		return &proto.Value{Float: 0, KnownType: proto.Property_Float}, nil
	}
	if value.Kind() == reflect.Float32 && e.Is32 {
		newF, _ := strconv.ParseFloat(fmt.Sprintf("%f", value.Interface().(float32)), 64)
		return &proto.Value{Float: newF, KnownType: proto.Property_Float}, nil
	} else if value.Kind() == reflect.Float64 && !e.Is32 {
		return &proto.Value{Float: value.Float(), KnownType: proto.Property_Float}, nil
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &proto.Value{Float: float64(value.Int()), KnownType: proto.Property_Float}, nil
	}
	return nil, UnsupportedTypeError
}

func (e Float) SetValue(value *proto.Value, onto reflect.Value) error {
	if onto.Kind() == reflect.Pointer {
		if onto.IsNil() {
			onto.Set(reflect.New(onto.Type().Elem()))
		}
		onto = onto.Elem()
	}
	if e.Is32 {
		newF, _ := strconv.ParseFloat(fmt.Sprintf("%f", value.GetFloat()), 64)
		onto.Set(reflect.ValueOf(float32(newF)))
		return nil
	}
	onto.SetFloat(value.Float)
	return nil
}

func (e Float) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_Float}
}
