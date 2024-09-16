package encoders

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"reflect"
	"strconv"
)

var InvalidMapError = errors.New("not a map[string][]byte")

func Map(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if mapVal, ok := value.Interface().(map[string][]byte); ok {
		ret := &proto.Value{Array: proto.NewRepeatedKeyValue()}
		for k, v := range mapVal {
			ret.Array.KeyValue[k] = v
		}
		return ret, nil
	}
	return nil, InvalidMapError
}

var InvalidStringMapError = errors.New("not a map[string]string")

func StringMap(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if mapVal, ok := value.Interface().(map[string]string); ok {
		ret := &proto.Value{Array: proto.NewRepeatedKeyValue()}
		for k, v := range mapVal {
			ret.Array.KeyValue[k] = []byte(v)
		}
		return ret, nil
	}
	return nil, InvalidStringMapError
}

var InvalidIntMapError = errors.New("not a map[string]int")

func IntMap(value reflect.Value) (*proto.Value, error) {
	value = deRef(value)
	if mapVal, ok := value.Interface().(map[string]int); ok {
		ret := &proto.Value{Array: proto.NewRepeatedKeyValue()}
		for k, v := range mapVal {
			ret.Array.KeyValue[k] = []byte(strconv.Itoa(v))
		}
		return ret, nil
	}
	return nil, InvalidIntMapError
}
