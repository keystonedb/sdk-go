package reflector

import (
	"reflect"

	"github.com/keystonedb/sdk-go/proto"
)

type StringMap struct{}

func (e StringMap) ToProto(value reflect.Value) (*proto.Value, error) {
	value = Deref(value)
	if mapVal, ok := value.Interface().(map[string]string); ok {
		ret := &proto.Value{Array: proto.NewRepeatedValue()}
		for k, v := range mapVal {
			ret.Array.KeyValue[k] = []byte(v)
		}
		return ret, nil
	}
	// Support maps with string-like values (e.g., custom types whose underlying type is string)
	if value.Kind() == reflect.Map && value.Type().Key().Kind() == reflect.String && value.Type().Elem().Kind() == reflect.String {
		ret := &proto.Value{Array: proto.NewRepeatedValue()}
		for _, key := range value.MapKeys() {
			k := key.String()
			v := value.MapIndex(key).String()
			ret.Array.KeyValue[k] = []byte(v)
		}
		return ret, nil
	}
	return nil, UnsupportedTypeError
}

func (e StringMap) SetValue(value *proto.Value, onto reflect.Value) error {
	if value.Array == nil {
		return InvalidValueError
	}
	// Build a map matching the destination type, supporting string-like element types.
	mapType := onto.Type()
	keyType := mapType.Key()
	elemType := mapType.Elem()
	// Only support string-like keys and values
	if keyType.Kind() != reflect.String || elemType.Kind() != reflect.String {
		return UnsupportedTypeError
	}
	ret := reflect.MakeMapWithSize(mapType, len(value.Array.KeyValue))
	for k, v := range value.Array.KeyValue {
		mk := reflect.New(keyType).Elem()
		mk.SetString(k)
		mv := reflect.New(elemType).Elem()
		mv.SetString(string(v))
		ret.SetMapIndex(mk, mv)
	}
	onto.Set(ret)
	return nil
}

func (e StringMap) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_KeyValue}
}
