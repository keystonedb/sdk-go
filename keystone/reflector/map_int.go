package reflector

import (
	"reflect"
	"strconv"

	"github.com/keystonedb/sdk-go/proto"
)

type IntMap struct{}

func (e IntMap) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if value.Kind() != reflect.Map || value.Type().Key().Kind() != reflect.String {
		return nil, UnsupportedTypeError
	}

	ret := &proto.Value{Array: proto.NewRepeatedValue(), KnownType: proto.Property_KeyValue}
	for _, key := range value.MapKeys() {
		k := key.String()
		v := value.MapIndex(key)

		var vStr string
		if v.CanInt() {
			vStr = strconv.FormatInt(v.Int(), 10)
		} else if v.CanUint() {
			vStr = strconv.FormatUint(v.Uint(), 10)
		} else {
			return nil, UnsupportedTypeError
		}
		ret.Array.KeyValue[k] = []byte(vStr)
	}
	return ret, nil
}

func (e IntMap) SetValue(value *proto.Value, onto reflect.Value) error {
	if value.Array == nil {
		return InvalidValueError
	}

	mapType := onto.Type()
	keyType := mapType.Key()
	elemType := mapType.Elem()

	if keyType.Kind() != reflect.String {
		return UnsupportedTypeError
	}

	res := reflect.MakeMapWithSize(mapType, len(value.Array.KeyValue))
	for k, v := range value.Array.KeyValue {
		mk := reflect.New(keyType).Elem()
		mk.SetString(k)

		mv := reflect.New(elemType).Elem()
		if elemType.Kind() >= reflect.Int && elemType.Kind() <= reflect.Int64 {
			i, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				return err
			}
			mv.SetInt(i)
		} else if elemType.Kind() >= reflect.Uint && elemType.Kind() <= reflect.Uint64 {
			u, err := strconv.ParseUint(string(v), 10, 64)
			if err != nil {
				return err
			}
			mv.SetUint(u)
		} else {
			return UnsupportedTypeError
		}
		res.SetMapIndex(mk, mv)
	}
	onto.Set(res)
	return nil
}

func (e IntMap) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_KeyValue}
}
