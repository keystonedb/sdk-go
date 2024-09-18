package reflector

import (
	"github.com/keystonedb/sdk-go/keystone/proto"
	"reflect"
	"strconv"
)

type IntMap struct{}

func (e IntMap) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if mapVal, ok := value.Interface().(map[string]int); ok {
		ret := &proto.Value{Array: proto.NewRepeatedKeyValue()}
		for k, v := range mapVal {
			ret.Array.KeyValue[k] = []byte(strconv.Itoa(v))
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e IntMap) SetValue(value *proto.Value, onto reflect.Value) error {
	if value.Array == nil {
		return InvalidValueError
	}
	res := make(map[string]int)
	for k, v := range value.Array.KeyValue {
		i, err := strconv.Atoi(string(v))
		if err != nil {
			return err
		}
		res[k] = i
	}

	onto.Set(reflect.ValueOf(res))
	return nil
}
