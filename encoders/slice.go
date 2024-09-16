package encoders

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"reflect"
)

var InvalidStringSliceError = errors.New("not a []string")

func StringSlice(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if slice, ok := value.Interface().([]string); ok {
		return &proto.Value{Array: &proto.RepeatedValue{Strings: slice}}, nil
	}
	return nil, InvalidStringSliceError
}

var InvalidInt64SliceError = errors.New("not a []int64")

func Int64Slice(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if slice, ok := value.Interface().([]int64); ok {
		return &proto.Value{Array: &proto.RepeatedValue{Ints: slice}}, nil
	}
	return nil, InvalidInt64SliceError
}

var InvalidIntSliceError = errors.New("not a []int")

func IntSlice(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if slice, ok := value.Interface().([]int); ok {
		ret := &proto.Value{Array: &proto.RepeatedValue{}}
		for _, i := range slice {
			ret.Array.Ints = append(ret.Array.Ints, int64(i))
		}
		return ret, nil
	}
	return nil, InvalidIntSliceError
}

var InvalidInt32SliceError = errors.New("not a []int32")

func Int32Slice(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if slice, ok := value.Interface().([]int32); ok {
		ret := &proto.Value{Array: &proto.RepeatedValue{}}
		for _, i := range slice {
			ret.Array.Ints = append(ret.Array.Ints, int64(i))
		}
		return ret, nil
	}
	return nil, InvalidInt32SliceError
}

var InvalidInt16SliceError = errors.New("not a []int16")

func Int16Slice(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if slice, ok := value.Interface().([]int16); ok {
		ret := &proto.Value{Array: &proto.RepeatedValue{}}
		for _, i := range slice {
			ret.Array.Ints = append(ret.Array.Ints, int64(i))
		}
		return ret, nil
	}
	return nil, InvalidInt16SliceError
}

var InvalidInt8SliceError = errors.New("not a []int8")

func Int8Slice(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if slice, ok := value.Interface().([]int8); ok {
		ret := &proto.Value{Array: &proto.RepeatedValue{}}
		for _, i := range slice {
			ret.Array.Ints = append(ret.Array.Ints, int64(i))
		}
		return ret, nil
	}
	return nil, InvalidInt8SliceError
}
