package encoders

import (
	"errors"
	"fmt"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"reflect"
	"strconv"
)

func Float32264(f float32) float64 {
	newF, _ := strconv.ParseFloat(fmt.Sprintf("%f", float32(f)), 64)
	return newF
}

func String(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	return &proto.Value{Text: value.String()}, nil
}

func Bool(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	return &proto.Value{Bool: value.Bool()}, nil
}

func Int(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	return &proto.Value{Int: value.Int()}, nil
}

var InvalidFloat32Error = errors.New("not a float32")

func Float32(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if value.Kind() != reflect.Float32 {
		return nil, InvalidFloat32Error
	}
	return &proto.Value{Float: Float32264(value.Interface().(float32))}, nil
}

func Float(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	return &proto.Value{Float: value.Float()}, nil
}

func Bytes(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	return &proto.Value{Raw: value.Bytes()}, nil
}
