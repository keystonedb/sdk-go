package reflector

import (
	"fmt"
	"github.com/keystonedb/sdk-go/proto"
	"reflect"
	"strconv"
)

type Float struct {
	Is32 bool
}

func (e Float) ToProto(value reflect.Value) (*proto.Value, error) {
	if value.Kind() == reflect.Float32 && e.Is32 {
		newF, _ := strconv.ParseFloat(fmt.Sprintf("%f", value.Interface().(float32)), 64)
		return &proto.Value{Float: newF}, nil
	} else if value.Kind() == reflect.Float64 && !e.Is32 {
		return &proto.Value{Float: value.Float()}, nil
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &proto.Value{Float: float64(value.Int())}, nil
	}
	return nil, UnsupportedTypeError
}

func (e Float) SetValue(value *proto.Value, onto reflect.Value) error {
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
